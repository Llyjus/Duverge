package com.duverge.userdata.service;

import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

@Service
public class RedisService {

    private final StringRedisTemplate redisTemplate;

    public RedisService(StringRedisTemplate redisTemplate) {
        this.redisTemplate = redisTemplate;
    }

    public void setLoginSession(String sessionId, String accountId) {
        redisTemplate.opsForValue().set(sessionId, 
                                        accountId);
    }

    public String getLoginSession(String sessionId) {
        if (sessionId == null || sessionId.isEmpty()) {
            return null;
        };
        return redisTemplate.opsForValue().get(sessionId);
    }
}