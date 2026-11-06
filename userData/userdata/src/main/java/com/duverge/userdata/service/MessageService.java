package com.duverge.userdata.service;

import java.util.Optional;
import java.util.function.Supplier;

import org.springframework.stereotype.Service;

import com.duverge.userdata.entity.Message;
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

    public Message createMessage(String messageId, String status) {
        Message message = new Message(messageId, status);
        return timedDatabase("save", () -> messageRepository.save(message));
    }

    private <T> T timedDatabase(String operation, Supplier<T> action) {
        return databaseExecutor.execute(() -> Timer.builder("userdata.database.operation")
                .description("User database operation duration")
                .tag("operation", operation)
                .register(meterRegistry)
                .record(action));
    }
}
