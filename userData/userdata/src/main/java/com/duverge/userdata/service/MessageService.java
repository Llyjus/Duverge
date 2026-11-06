package com.duverge.userdata.service;

import java.util.Optional;
import java.util.function.Supplier;
import java.time.LocalDateTime;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.dao.DataIntegrityViolationException;

import com.duverge.userdata.entity.Message;
import com.duverge.userdata.error.DatabaseConstraintClassifier;
import com.duverge.userdata.error.DuplicateMessageIdException;
import com.duverge.userdata.internal.infra.LimitedDatabaseExecutor;
import com.duverge.userdata.repository.MessageRepository;

import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.Timer;

@Service
public class MessageService {

    private final MessageRepository messageRepository;
    private final MeterRegistry meterRegistry;
    private final LimitedDatabaseExecutor databaseExecutor;

    public MessageService(
            MessageRepository messageRepository,
            MeterRegistry meterRegistry,
            LimitedDatabaseExecutor databaseExecutor) {
        this.messageRepository = messageRepository;
        this.meterRegistry = meterRegistry;
        this.databaseExecutor = databaseExecutor;
    }

    public Optional<Message> findByMessageId(String messageId) {
        return timedDatabase("find_by_message_id", () -> messageRepository.findByMessageId(messageId));
    }

    @Transactional
    public Message createMessage(String messageId) {
        Message message = new Message(messageId);
        try {
            // Flush here so a duplicate ID is classified at this boundary,
            // rather than leaking out later as an ambiguous commit failure.
            return timedDatabase("save", () -> messageRepository.saveAndFlush(message));
        } catch (DataIntegrityViolationException e) {
            if (DatabaseConstraintClassifier.violates(e, "uk_messages_id")) {
                throw new DuplicateMessageIdException("Message ID has already been recorded", e);
            }
            throw e;
        }
    }

    @Transactional
    public int deleteMessagesOlderThanOneDay() {
        LocalDateTime cutoff = LocalDateTime.now().minusDays(1);
        return timedDatabase("delete_older_than_one_day", () -> messageRepository.deleteCreatedBefore(cutoff));
    }

    private <T> T timedDatabase(String operation, Supplier<T> action) {
        return databaseExecutor.execute(() -> Timer.builder("userdata.database.operation")
                .description("User database operation duration")
                .tag("operation", operation)
                .register(meterRegistry)
                .record(action));
    }
}
