package com.duverge.userdata.service;

import static org.junit.jupiter.api.Assertions.assertDoesNotThrow;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertNull;

import org.junit.jupiter.api.Test;

import com.duverge.userdata.error.DuplicateMessageIdException;
import com.duverge.userdata.error.MessageIdStateException;
import com.duverge.userdata.internal.infra.redis.RedisService;

class IdempotencyServiceTests {

    @Test
    void allowsAnUncompletedMessageIdFromRedis() {
        StubRedisService redisService = new StubRedisService(RedisService.MESSAGE_ID_UNCOMPLETED);
        IdempotencyService idempotencyService = new IdempotencyService(redisService);

        assertDoesNotThrow(() -> idempotencyService.ensureNewMessageId("new-id"));
        assertEquals("new-id", redisService.lastReadId);
    }

    @Test
    void rejectsACompletedMessageIdAsAlreadyProcessed() {
        StubRedisService redisService = new StubRedisService(RedisService.MESSAGE_ID_COMPLETED);
        IdempotencyService idempotencyService = new IdempotencyService(redisService);

        assertThrows(DuplicateMessageIdException.class,
                () -> idempotencyService.ensureNewMessageId("completed-id"));
        assertEquals("completed-id", redisService.lastReadId);
    }

    @Test
    void rejectsMissingOrUnknownRedisStates() {
        StubRedisService redisService = new StubRedisService(null);
        IdempotencyService idempotencyService = new IdempotencyService(redisService);

        assertThrows(MessageIdStateException.class,
                () -> idempotencyService.ensureNewMessageId("missing-id"));
        assertThrows(MessageIdStateException.class,
                () -> new IdempotencyService(new StubRedisService("unexpected"))
                        .ensureNewMessageId("unknown-state-id"));
    }

    @Test
    void rejectsBlankMessageIdWithoutQueryingRedis() {
        StubRedisService redisService = new StubRedisService(RedisService.MESSAGE_ID_UNCOMPLETED);

        assertThrows(MessageIdStateException.class,
                () -> new IdempotencyService(redisService).ensureNewMessageId(" "));
        assertNull(redisService.lastReadId);
    }

    private static final class StubRedisService extends RedisService {
        private final String status;
        private String lastReadId;

        private StubRedisService(String status) {
            super(null);
            this.status = status;
        }

        @Override
        public String getMessageStatus(String messageId) {
            lastReadId = messageId;
            return status;
        }
    }
}
