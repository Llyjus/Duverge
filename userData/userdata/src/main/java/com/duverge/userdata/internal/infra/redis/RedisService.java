package com.duverge.userdata.internal.infra.redis;

import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;
import java.time.Duration;

@Service
public class RedisService {
    public static final String MESSAGE_ID_KEY_PREFIX = "message_id:";
    public static final String MESSAGE_ID_UNCOMPLETED = "uncompleted";
    public static final String MESSAGE_ID_COMPLETED = "completed";
    public static final Duration MESSAGE_ID_TTL = Duration.ofHours(1);

    private final StringRedisTemplate redisTemplate;

    public RedisService(StringRedisTemplate redisTemplate) {
        this.redisTemplate = redisTemplate;
    }

    /** Returns the current Redis status for a Gateway-reserved Message ID. */
    public String getMessageStatus(String messageId) {
        if (messageId == null || messageId.isBlank()) {
            return null;
        }
        return redisTemplate.opsForValue().get(MESSAGE_ID_KEY_PREFIX + messageId);
    }

    /** True only while Gateway's reservation still has the expected initial value. */
    public boolean isMessageUncompleted(String messageId) {
        return MESSAGE_ID_UNCOMPLETED.equals(getMessageStatus(messageId));
    }

    /** Marks a Gateway-reserved Message ID complete after its database work commits. */
    public void markMessageCompleted(String messageId) {
        if (messageId == null || messageId.isBlank()) {
            throw new IllegalArgumentException("Message ID is required");
        }
        redisTemplate.opsForValue().set(
                MESSAGE_ID_KEY_PREFIX + messageId,
                MESSAGE_ID_COMPLETED,
                MESSAGE_ID_TTL);
    }
}
