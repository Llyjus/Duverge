package com.duverge.userdata.service;

import org.springframework.stereotype.Service;

import com.duverge.userdata.error.DuplicateMessageIdException;
import com.duverge.userdata.error.MessageIdStateException;
import com.duverge.userdata.internal.infra.redis.RedisService;

/** Validates request Message IDs against the Gateway-managed Redis state. */
@Service
public class IdempotencyService {
    private final RedisService redisService;

    public IdempotencyService(RedisService redisService) {
        this.redisService = redisService;
    }

    /**
     * Rejects a previously recorded ID before business work starts. The unique
     * database constraint remains the final guard against concurrent requests.
     */
    public void ensureNewMessageId(String messageId) {
        if (messageId == null || messageId.isBlank()) {
            throw new MessageIdStateException("Message ID is required");
        }
        String status = redisService.getMessageStatus(messageId);
        if (RedisService.MESSAGE_ID_COMPLETED.equals(status)) {
            throw new DuplicateMessageIdException("Message ID has already been processed", null);
        }
        if (!RedisService.MESSAGE_ID_UNCOMPLETED.equals(status)) {
            throw new MessageIdStateException("Message ID is missing or is not uncompleted in Redis");
        }
    }

    /** Updates Redis only when called after the associated database transaction commits. */
    public void markMessageCompleted(String messageId) {
        redisService.markMessageCompleted(messageId);
    }
}
