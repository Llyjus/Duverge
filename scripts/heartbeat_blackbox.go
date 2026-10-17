package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type tcpMessage struct {
	Version string          `json:"version"`
	Type    string          `json:"type"`
	Data    json.RawMessage `json:"data"`
	Success bool            `json:"success"`
}

type tcpClient struct {
	conn net.Conn
}

func (c *tcpClient) send(v any) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(body)))
	if _, err = c.conn.Write(append(header[:], body...)); err != nil {
		return err
	}
	return nil
}

func (c *tcpClient) receive() (tcpMessage, error) {
	var header [4]byte
	if _, err := io.ReadFull(c.conn, header[:]); err != nil {
		return tcpMessage{}, err
	}
	body := make([]byte, binary.BigEndian.Uint32(header[:]))
	if _, err := io.ReadFull(c.conn, body); err != nil {
		return tcpMessage{}, err
	}
	var message tcpMessage
	return message, json.Unmarshal(body, &message)
}

func httpJSON(baseURL, path string, payload any, output any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("%s returned %d: %s", path, res.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	if output != nil && len(responseBody) > 0 {
		return json.Unmarshal(responseBody, output)
	}
	return nil
}

func require(condition bool, format string, args ...any) {
	if !condition {
		panic(fmt.Sprintf(format, args...))
	}
}

func login(baseURL, account, password string) string {
	var registered struct {
		Success bool `json:"success"`
	}
	require(httpJSON(baseURL, "/login", map[string]any{
		"type": "register", "accountId": account, "password": password,
	}, &registered) == nil && registered.Success, "registration failed for %s", account)

	var response struct {
		Success   bool   `json:"success"`
		SessionID string `json:"sessionId"`
	}
	require(httpJSON(baseURL, "/login", map[string]any{
		"type": "login", "accountId": account, "password": password,
	}, &response) == nil && response.Success && response.SessionID != "", "login failed for %s", account)
	return response.SessionID
}

func authenticate(address, sessionID string) *tcpClient {
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	require(err == nil, "TCP connection failed: %v", err)
	client := &tcpClient{conn: conn}
	require(client.send(map[string]any{
		"version": "1.0", "type": "auth",
		"data": map[string]string{"token": sessionID},
	}) == nil, "TCP auth send failed")
	return client
}

func heartbeat(client *tcpClient, label string) {
	require(client.send(map[string]any{"version": "1.0", "type": "heartbeat", "data": nil}) == nil,
		"%s heartbeat send failed", label)
	response, err := client.receive()
	require(err == nil && response.Success, "%s heartbeat failed: err=%v response=%+v", label, err, response)
	fmt.Printf("PASS %-28s response success\n", label)
}

func seedStatus(userID string) error {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:6379", 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	value := fmt.Sprintf(`{"user_id":%q,"status":"in_lobby"}`, userID)
	command := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\nuser_status:%s\r\n$%d\r\n%s\r\n",
		len(userID)+12, userID, len(value), value)
	if _, err := io.WriteString(conn, command); err != nil {
		return err
	}
	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return err
	}
	if !strings.HasPrefix(string(response), "+OK") {
		return fmt.Errorf("redis SET failed: %s", response)
	}
	return nil
}

func chat(client *tcpClient, message any, label string) {
	require(client.send(message) == nil, "%s send failed", label)
	response, err := client.receive()
	require(err == nil, "%s response read failed: %v", label, err)
	fmt.Printf("PASS %-28s received %s\n", label, response.Type)
}

func main() {
	baseURL := "http://127.0.0.1:26526"
	tcpAddress := "127.0.0.1:26527"
	if value := os.Getenv("BASE_URL"); value != "" {
		baseURL = value
	}
	if value := os.Getenv("TCP_ADDRESS"); value != "" {
		tcpAddress = value
	}
	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	password := "heartbeat-test-password"
	user1 := "heartbeat_" + suffix + "_one"
	user2 := "heartbeat_" + suffix + "_two"

	healthResponse, err := http.Get(baseURL + "/health")
	require(err == nil, "health check failed: %v", err)
	require(healthResponse.StatusCode == http.StatusOK, "health check returned %d", healthResponse.StatusCode)
	healthResponse.Body.Close()
	fmt.Println("PASS health                         gateway is reachable")

	session1 := login(baseURL, user1, password)
	session2 := login(baseURL, user2, password)
	fmt.Println("PASS login                          two users authenticated")

	client1 := authenticate(tcpAddress, session1)
	defer client1.conn.Close()
	client2 := authenticate(tcpAddress, session2)
	defer client2.conn.Close()
	fmt.Println("PASS TCP auth                       two connections authenticated")
	if os.Getenv("SEED_STATUS") == "1" {
		require(seedStatus(user1) == nil && seedStatus(user2) == nil, "status fixture setup failed")
		fmt.Println("INFO status fixture                seeded JSON status for room service compatibility")
	}

	var room struct {
		Success bool   `json:"success"`
		RoomID  string `json:"roomId"`
	}
	require(httpJSON(baseURL, "/room/user", map[string]any{
		"type": "createRoom", "sessionId": session1,
	}, &room) == nil && room.RoomID != "", "room creation failed")
	fmt.Printf("PASS room creation                 room=%s\n", room.RoomID)

	heartbeat(client1, "heartbeat t=0 user1")
	heartbeat(client2, "heartbeat t=0 user2")
	time.Sleep(10 * time.Second)
	heartbeat(client1, "heartbeat t=10s user1")
	heartbeat(client2, "heartbeat t=10s user2")
	time.Sleep(10 * time.Second)
	heartbeat(client1, "heartbeat t=20s user1")
	heartbeat(client2, "heartbeat t=20s user2")
	fmt.Println("PASS heartbeat chain                users stayed connected through 20 seconds")

	var join struct {
		Result bool `json:"result"`
	}
	require(httpJSON(baseURL, "/room/room", map[string]any{
		"type": "joinRoom", "sessionId": session2, "roomId": room.RoomID,
	}, &join) == nil && join.Result, "second user failed to join room")
	fmt.Println("PASS room join                      second user joined")
	heartbeat(client1, "heartbeat in room user1")
	heartbeat(client2, "heartbeat in room user2")
	fmt.Println("PASS room presence                  both users renewed status while in room")

	roomChat := map[string]any{
		"version": "1.0", "type": "chat",
		"data": map[string]any{"type": "room", "data": map[string]string{
			"room_id": room.RoomID, "type": "text", "content": "heartbeat room chat",
		}},
	}
	require(client1.send(roomChat) == nil, "room chat send failed")
	first, err := client1.receive()
	require(err == nil && first.Type == "chat", "sender did not receive room chat: %v %+v", err, first)
	second, err := client1.receive()
	require(err == nil && second.Success, "sender did not receive room chat response: %v %+v", err, second)
	received, err := client2.receive()
	require(err == nil && received.Type == "chat", "recipient did not receive room chat: %v %+v", err, received)
	fmt.Println("PASS room chat                      both room members received message")

	privateChat := map[string]any{
		"version": "1.0", "type": "chat",
		"data": map[string]any{"type": "private", "data": map[string]string{
			"to": user2, "type": "text", "content": "heartbeat private chat",
		}},
	}
	require(client1.send(privateChat) == nil, "private chat send failed")
	senderResponse, err := client1.receive()
	require(err == nil && senderResponse.Success, "private chat sender response failed: %v %+v", err, senderResponse)
	recipientMessage, err := client2.receive()
	require(err == nil && recipientMessage.Type == "chat", "private chat recipient failed: %v %+v", err, recipientMessage)
	fmt.Println("PASS private chat                   recipient received message")

	fmt.Println("PASS black-box flow                 login -> room -> heartbeat -> chat -> private chat")
}
