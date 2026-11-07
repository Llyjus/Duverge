package com.duverge.userdata.service;

import static org.junit.jupiter.api.Assertions.assertDoesNotThrow;
import static org.junit.jupiter.api.Assertions.assertThrows;

import java.util.HashSet;
import java.util.Optional;
import java.util.Set;

import org.junit.jupiter.api.Test;

import com.duverge.userdata.entity.Message;
import com.duverge.userdata.error.DuplicateMessageIdException;

class IdempotencyServiceTests {

    @Test
    void allowsAnUnseenMessageId() {
        StubMessageService messageService = new StubMessageService();
        IdempotencyService idempotencyService = new IdempotencyService(messageService);

        assertDoesNotThrow(() -> idempotencyService.ensureNewMessageId("new-id"));
        org.junit.jupiter.api.Assertions.assertEquals("new-id", messageService.lastQueriedId);
    }

    @Test
    void rejectsAnAlreadyRecordedMessageIdWithTheTypedError() {
        StubMessageService messageService = new StubMessageService("used-id");
        IdempotencyService idempotencyService = new IdempotencyService(messageService);

        assertThrows(DuplicateMessageIdException.class,
                () -> idempotencyService.ensureNewMessageId("used-id"));
        org.junit.jupiter.api.Assertions.assertEquals("used-id", messageService.lastQueriedId);
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
}
