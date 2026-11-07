package com.duverge.userdata.service;

import static org.junit.jupiter.api.Assertions.assertDoesNotThrow;
import static org.junit.jupiter.api.Assertions.assertThrows;

import java.util.HashSet;
import java.util.Optional;
import java.util.Set;

import org.junit.jupiter.api.Test;

import com.duverge.userdata.entity.Message;
import com.duverge.userdata.error.DuplicateMessageIdException;
import com.duverge.userdata.error.MessageIdStateException;
import com.duverge.userdata.internal.infra.redis.RedisService;

class IdempotencyServiceTests {

    @Test
    void allowsAnUnseenMessageId() {
        StubMessageService messageService = new StubMessageService();
        IdempotencyService idempotencyService = new IdempotencyService(messageService,
                new StubRedisService(RedisService.MESSAGE_ID_UNCOMPLETED));

        assertDoesNotThrow(() -> idempotencyService.ensureNewMessageId("new-id"));
        org.junit.jupiter.api.Assertions.assertEquals("new-id", messageService.lastQueriedId);
    }

    @Test
    void rejectsAnAlreadyRecordedMessageIdWithTheTypedError() {
        StubMessageService messageService = new StubMessageService("used-id");
        IdempotencyService idempotencyService = new IdempotencyService(messageService,
                new StubRedisService(RedisService.MESSAGE_ID_UNCOMPLETED));

        assertThrows(DuplicateMessageIdException.class,
                () -> idempotencyService.ensureNewMessageId("used-id"));
        org.junit.jupiter.api.Assertions.assertEquals("used-id", messageService.lastQueriedId);
    }

    @Test
    void rejectsMessageIdsMissingFromRedisOrNotInUncompletedState() {
        StubMessageService messageService = new StubMessageService();

        assertThrows(MessageIdStateException.class,
                () -> new IdempotencyService(messageService, new StubRedisService(null))
                        .ensureNewMessageId("missing-id"));
        assertThrows(MessageIdStateException.class,
                () -> new IdempotencyService(messageService, new StubRedisService("completed"))
                        .ensureNewMessageId("completed-id"));
        org.junit.jupiter.api.Assertions.assertEquals(null, messageService.lastQueriedId);
    }

    private static final class StubMessageService extends MessageService {
        private final Set<String> existingIds = new HashSet<>();
        private String lastQueriedId;

        private StubMessageService(String... existingIds) {
            super(null, null, null);
            this.existingIds.addAll(Set.of(existingIds));
        }

        @Override
        public Optional<Message> findByMessageId(String messageId) {
            lastQueriedId = messageId;
            return existingIds.contains(messageId)
                    ? Optional.of(new Message(messageId))
                    : Optional.empty();
        }
    }

    private static final class StubRedisService extends RedisService {
        private final String status;

        private StubRedisService(String status) {
            super(null);
            this.status = status;
        }

        @Override
        public String getMessageStatus(String messageId) {
            return status;
        }
    }
}
