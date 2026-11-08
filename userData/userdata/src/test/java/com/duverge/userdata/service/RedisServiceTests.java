package com.duverge.userdata.service;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.lang.reflect.Proxy;
import java.time.Duration;

import org.junit.jupiter.api.Test;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.data.redis.core.ValueOperations;

import com.duverge.userdata.internal.infra.redis.RedisService;

class RedisServiceTests {

    @Test
    void retrievesGatewayMessageStatusUsingItsRedisKeyAndChecksUncompletedState() {
        StubStringRedisTemplate redisTemplate = new StubStringRedisTemplate("uncompleted");
        RedisService redisService = new RedisService(redisTemplate);

        assertEquals("uncompleted", redisService.getMessageStatus("request-123"));
        assertEquals("message_id:request-123", redisTemplate.lastReadKey);
        assertTrue(redisService.isMessageUncompleted("request-123"));

        redisTemplate.status = "completed";
        assertFalse(redisService.isMessageUncompleted("request-123"));

        redisService.markMessageCompleted("request-123");
        assertEquals("message_id:request-123", redisTemplate.lastWrittenKey);
        assertEquals("completed", redisTemplate.status);
        assertEquals(Duration.ofHours(1), redisTemplate.lastWriteTtl);
    }

    @Test
    void doesNotQueryRedisForBlankMessageIds() {
        StubStringRedisTemplate redisTemplate = new StubStringRedisTemplate("uncompleted");
        RedisService redisService = new RedisService(redisTemplate);

        assertNull(redisService.getMessageStatus(" "));
        assertFalse(redisService.isMessageUncompleted(""));
        assertNull(redisTemplate.lastReadKey);
    }

    private static final class StubStringRedisTemplate extends StringRedisTemplate {
        private String status;
        private String lastReadKey;
        private String lastWrittenKey;
        private Duration lastWriteTtl;

        private StubStringRedisTemplate(String status) {
            this.status = status;
        }

        @Override
        @SuppressWarnings("unchecked")
        public ValueOperations<String, String> opsForValue() {
            return (ValueOperations<String, String>) Proxy.newProxyInstance(
                    ValueOperations.class.getClassLoader(),
                    new Class<?>[] { ValueOperations.class },
                    (proxy, method, args) -> {
                        if (method.getName().equals("get")) {
                            lastReadKey = (String) args[0];
                            return status;
                        }
                        if (method.getName().equals("set")) {
                            lastWrittenKey = (String) args[0];
                            status = (String) args[1];
                            if (args.length == 3) {
                                lastWriteTtl = (Duration) args[2];
                            }
                            return null;
                        }
                        if (method.getName().equals("toString")) {
                            return "StubValueOperations";
                        }
                        return null;
                    });
        }
    }
}
