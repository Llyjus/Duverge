package com.duverge.userdata.internal;

import java.security.SecureRandom;
import java.util.Base64;


public class RandomSession {
    public String generateSession() {
        SecureRandom random = new SecureRandom();

        byte[] bytes = new byte[32];
        random.nextBytes(bytes);

        String sessionId = Base64.getUrlEncoder()
                .withoutPadding()
                .encodeToString(bytes);
        return sessionId;
    }
}
