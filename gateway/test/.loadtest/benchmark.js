// Each VU iteration runs one complete lifecycle with fresh accounts. Test
// accounts are persisted, and CreateRoom keeps the creator in the room after
// the guest exits, so use a fresh RUN_ID and clean test data after the run.
import http from 'k6/http';
import { check } from 'k6';
import { Trend } from 'k6/metrics';
import exec from 'k6/execution';
import { Socket } from 'k6/x/tcp';

const BASE_URL = (__ENV.BASE_URL || 'http://127.0.0.1:26526').replace(/\/$/, '');
const TCP_HOST = __ENV.TCP_HOST || '127.0.0.1';
const TCP_PORT = Number(__ENV.TCP_PORT || 26527);
const ACCOUNT_PREFIX = (__ENV.ACCOUNT_PREFIX || 'k6').replace(/[^a-zA-Z0-9_-]/g, '_');
const RUN_ID = (__ENV.RUN_ID || String(Date.now())).replace(/[^a-zA-Z0-9_-]/g, '_');
const TEST_PASSWORD = __ENV.TEST_PASSWORD || `K6-${RUN_ID}-load-test`;
const SOCKET_TIMEOUT_MS = Number(__ENV.SOCKET_TIMEOUT_MS || 15000);
const MAX_FRAME_BYTES = 64 * 1024;
const START_TIME_25 = __ENV.START_TIME_25 || '120s';
const START_TIME_50 = __ENV.START_TIME_50 || '240s';
const ITERATIONS = Number(__ENV.ITERATIONS || 1);
if (!Number.isInteger(ITERATIONS) || ITERATIONS < 1) {
  throw new Error('ITERATIONS must be a positive integer');
}

const chatRoundTrip = new Trend('tcp_room_chat_round_trip', true);

export const options = {
  scenarios: {
    vus_10: {
      executor: 'per-vu-iterations',
      vus: 10,
      iterations: ITERATIONS,
      startTime: '0s',
      maxDuration: '110s',
      env: { LOAD_LEVEL: '10' },
    },
    vus_25: {
      executor: 'per-vu-iterations',
      vus: 25,
      iterations: ITERATIONS,
      startTime: START_TIME_25,
      maxDuration: '110s',
      env: { LOAD_LEVEL: '25' },
    },
    vus_50: {
      executor: 'per-vu-iterations',
      vus: 50,
      iterations: ITERATIONS,
      startTime: START_TIME_50,
      maxDuration: '110s',
      env: { LOAD_LEVEL: '50' },
    },
  },
  summaryTrendStats: ['avg', 'med', 'p(90)', 'p(95)', 'p(99)', 'max'],
};

function checkOrThrow(response, name, predicate) {
  const passed = check(response, { [name]: predicate });
  if (!passed) {
    throw new Error(`${name} failed; HTTP ${response.status}: ${response.body}`);
  }
}

function postJSON(path, operation, payload) {
  return http.post(`${BASE_URL}${path}`, JSON.stringify(payload), {
    headers: { 'Content-Type': 'application/json' },
    tags: { operation },
    timeout: '30s',
  });
}

function register(accountId) {
  const response = postJSON('/login', 'register', {
    type: 'register',
    accountId,
    password: TEST_PASSWORD,
  });

  // A 401 means that this generated account already exists (for example when
  // rerunning with the same RUN_ID). The following login verifies the account.
  checkOrThrow(response, 'registered or existing test account', (r) => r.status === 200 || r.status === 401);
}

function login(accountId) {
  const response = postJSON('/login', 'login', {
    type: 'login',
    accountId,
    password: TEST_PASSWORD,
  });
  checkOrThrow(response, 'login succeeded', (r) => r.status === 200);

  const result = response.json();
  if (!result.success || !result.sessionId) {
    throw new Error(`login returned no session for ${accountId}`);
  }
  return result.sessionId;
}

function concatenate(left, right) {
  const result = new Uint8Array(left.length + right.length);
  result.set(left, 0);
  result.set(right, left.length);
  return result;
}

function encodeFrame(message) {
  const body = new TextEncoder().encode(JSON.stringify(message));
  if (body.length === 0 || body.length > MAX_FRAME_BYTES) {
    throw new Error(`TCP message size out of range: ${body.length}`);
  }

  const frame = new Uint8Array(body.length + 4);
  frame[0] = (body.length >>> 24) & 0xff;
  frame[1] = (body.length >>> 16) & 0xff;
  frame[2] = (body.length >>> 8) & 0xff;
  frame[3] = body.length & 0xff;
  frame.set(body, 4);
  return frame.buffer;
}

function makeTCPClient() {
  const state = {
    socket: new Socket(),
    buffer: new Uint8Array(0),
    packets: [],
    waiters: [],
    error: null,
  };

  state.socket.on('data', (data) => {
    state.buffer = concatenate(state.buffer, data);

    while (state.buffer.length >= 4) {
      const bodyLength = (
        (state.buffer[0] * 0x1000000) +
        (state.buffer[1] << 16) +
        (state.buffer[2] << 8) +
        state.buffer[3]
      ) >>> 0;

      if (bodyLength === 0 || bodyLength > MAX_FRAME_BYTES) {
        state.error = `invalid TCP frame length ${bodyLength}`;
        return;
      }
      if (state.buffer.length < bodyLength + 4) return;

      const body = state.buffer.slice(4, bodyLength + 4);
      state.buffer = state.buffer.slice(bodyLength + 4);
      try {
        state.packets.push(JSON.parse(new TextDecoder().decode(body)));
      } catch (error) {
        state.error = `invalid JSON from TCP server: ${error}`;
        break;
      }
    }

    // Resolve pending packet waits from the socket callback. k6's sleep()
    // blocks the JS event loop, so polling it can prevent this callback from
    // running until the timeout expires.
    for (let i = state.waiters.length - 1; i >= 0; i--) {
      const waiter = state.waiters[i];
      const index = state.packets.findIndex(waiter.predicate);
      if (index >= 0) {
        state.waiters.splice(i, 1);
        waiter.resolve(state.packets.splice(index, 1)[0]);
      }
    }
  });

  state.socket.on('error', (error) => {
    state.error = String(error);
    state.waiters.splice(0).forEach((waiter) => waiter.reject(new Error(state.error)));
  });
  state.socket.on('timeout', () => {
    state.error = `TCP socket timed out after ${SOCKET_TIMEOUT_MS}ms`;
    state.waiters.splice(0).forEach((waiter) => waiter.reject(new Error(state.error)));
  });
  state.socket.setTimeout(SOCKET_TIMEOUT_MS);
  return state;
}

async function waitForPacket(client, predicate, label) {
  if (client.error) throw new Error(`${label}: ${client.error}`);
  const index = client.packets.findIndex(predicate);
  if (index >= 0) return client.packets.splice(index, 1)[0];

  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      const waiterIndex = client.waiters.indexOf(waiter);
      if (waiterIndex >= 0) client.waiters.splice(waiterIndex, 1);
      reject(new Error(`timed out waiting for ${label}`));
    }, SOCKET_TIMEOUT_MS);
    const waiter = {
      predicate,
      resolve: (packet) => {
        clearTimeout(timer);
        resolve(packet);
      },
      reject: (error) => {
        clearTimeout(timer);
        reject(new Error(`${label}: ${error.message}`));
      },
    };
    client.waiters.push(waiter);
  });
}

async function connectAndAuthenticate(sessionId) {
  const client = makeTCPClient();
  await client.socket.connect(TCP_PORT, TCP_HOST);
  await client.socket.write(encodeFrame({
    version: '1.0',
    type: 'auth',
    data: { token: sessionId },
  }));

  const response = await waitForPacket(client, (packet) => packet.type === 'response', 'TCP authentication response');
  if (!response.success) throw new Error(`TCP authentication failed: ${JSON.stringify(response.data)}`);
  return client;
}

async function sendRoomChat(sender, recipient, roomId, content) {
  const started = Date.now();
  await sender.socket.write(encodeFrame({
    version: '1.0',
    request_id: `k6-${Date.now()}`,
    type: 'chat',
    data: {
      type: 'room',
      data: { room_id: roomId, type: 'text', content },
    },
  }));

  const senderMessage = await waitForPacket(sender, (packet) => packet.type === 'chat', 'sender room chat');
  const senderAck = await waitForPacket(sender, (packet) => packet.type === 'response', 'sender chat acknowledgement');
  const recipientMessage = await waitForPacket(recipient, (packet) => packet.type === 'chat', 'recipient room chat');

  const valid = senderMessage.success && recipientMessage.success && senderAck.success &&
    senderMessage.data.content === content && recipientMessage.data.content === content;
  check(valid, { 'room chat delivered to sender and recipient': (value) => value });
  if (!valid) throw new Error('room chat was not delivered to both room members');
  chatRoundTrip.add(Date.now() - started);
}

export default async function () {
  const level = __ENV.LOAD_LEVEL || 'manual';
  const scenarioName = exec.scenario.name;
  const accountStem = `${ACCOUNT_PREFIX}_${RUN_ID}_${scenarioName}_vu${__VU}_iter${__ITER}`;
  const hostId = `${accountStem}_host`;
  const guestId = `${accountStem}_guest`;
  let hostTCP;
  let guestTCP;
  let guestSession = '';
  let roomId = '';
  let guestJoined = false;

  try {
    register(hostId);
    register(guestId);

    const hostSession = login(hostId);
    guestSession = login(guestId);
    hostTCP = await connectAndAuthenticate(hostSession);
    guestTCP = await connectAndAuthenticate(guestSession);

    const createResponse = postJSON('/room/user', 'room_create', {
      type: 'createRoom',
      sessionId: hostSession,
    });
    checkOrThrow(createResponse, 'room created', (r) => r.status === 200);
    roomId = createResponse.json().roomId;
    if (!roomId) throw new Error('room creation returned no roomId');

    const joinResponse = postJSON('/room/room', 'room_join', {
      type: 'joinRoom',
      sessionId: guestSession,
      roomId,
    });
    checkOrThrow(joinResponse, 'guest joined room', (r) => r.status === 200 && r.json().result === true);
    guestJoined = true;

    await sendRoomChat(hostTCP, guestTCP, roomId, `k6 ${level} VU room chat`);

    const leaveResponse = postJSON('/room/room', 'room_leave', {
      type: 'leaveRoom',
      sessionId: guestSession,
      roomId,
    });
    checkOrThrow(leaveResponse, 'guest left room', (r) => r.status === 200 && r.json().result === true);
    guestJoined = false;
  } finally {
    // Try to release the guest if a later step failed after joining.
    if (guestJoined && roomId) {
      postJSON('/room/room', 'room_leave_cleanup', {
        type: 'leaveRoom',
        sessionId: guestSession,
        roomId,
      });
    }
    if (hostTCP) hostTCP.socket.destroy();
    if (guestTCP) guestTCP.socket.destroy();
  }
}
