package com.duverge.userdata.service;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

import com.duverge.userdata.entity.User;
import com.duverge.userdata.internal.Infra.LimitedDatabaseExecutor;
import com.duverge.userdata.internal.Infra.LimitedPasswordEncoder;
import com.duverge.userdata.repository.UserRepository;
import io.micrometer.core.instrument.simple.SimpleMeterRegistry;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import java.time.Duration;
import java.lang.reflect.Proxy;
import java.util.HashMap;
import java.util.Map;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

class UserServicePasswordMetricsTests {
    private UserRepository userRepository;
    private SimpleMeterRegistry meterRegistry;
    private UserService userService;
    private LimitedPasswordEncoder passwordEncoder;
    private Map<String, User> users;

    @BeforeEach
    void setUp() {
        users = new HashMap<>();
        userRepository = (UserRepository) Proxy.newProxyInstance(
                UserRepository.class.getClassLoader(), new Class<?>[] { UserRepository.class },
                (proxy, method, args) -> switch (method.getName()) {
                    case "findByAccountId" -> users.get(args[0]);
                    case "save" -> {
                        User user = (User) args[0];
                        users.put(user.getAccountId(), user);
                        yield user;
                    }
                    case "toString" -> "InMemoryUserRepositoryStub";
                    case "hashCode" -> System.identityHashCode(proxy);
                    case "equals" -> proxy == args[0];
                    default -> null;
                });
        meterRegistry = new SimpleMeterRegistry();
        passwordEncoder = new LimitedPasswordEncoder(meterRegistry, 10, Duration.ofSeconds(1), 10);
        userService = new UserService(userRepository, meterRegistry, passwordEncoder,
                new LimitedDatabaseExecutor(meterRegistry, 10, Duration.ofSeconds(1), 10));
    }

    @Test
    void encodeGaugeReturnsToZeroAfterSuccessfulHash() {
        assertTrue(userService.setNewUser("new-user", "test-password"));

        assertInFlight("encode", 0);
        assertEquals(1, meterRegistry.get("userdata.password.operation").tag("operation", "encode").timer().count());
        assertEquals(1, meterRegistry.get("userdata.password.permit.wait").timer().count());
    }

    @Test
    void verifyGaugeReturnsToZeroAfterSuccessfulVerification() {
        User user = new User();
        user.setAccountId("existing-user");
        user.setPassword(new BCryptPasswordEncoder().encode("test-password"));
        users.put("existing-user", user);

        assertTrue(userService.checkUser("existing-user", "test-password"));
        assertFalse(userService.checkUser("existing-user", "wrong-password"));

        assertInFlight("verify", 0);
        assertEquals(2, meterRegistry.get("userdata.password.operation").tag("operation", "verify").timer().count());
        assertEquals(2, meterRegistry.get("userdata.password.permit.wait").timer().count());
    }

    private void assertInFlight(String operation, double expected) {
        assertEquals(expected, meterRegistry.get("password.hash.in.flight")
                .tag("operation", operation).gauge().value());
    }

}
