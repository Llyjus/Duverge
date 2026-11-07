package com.duverge.userdata.service;

import java.lang.reflect.Proxy;
import java.time.Duration;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.transaction.PlatformTransactionManager;
import org.springframework.transaction.TransactionDefinition;
import org.springframework.transaction.TransactionStatus;
import org.springframework.transaction.support.SimpleTransactionStatus;

import com.duverge.userdata.entity.User;
import com.duverge.userdata.internal.infra.LimitedDatabaseExecutor;
import com.duverge.userdata.internal.infra.LimitedPasswordEncoder;
import com.duverge.userdata.internal.infra.redis.RedisService;
import com.duverge.userdata.repository.UserRepository;
import com.duverge.userdata.repository.MessageRepository;

import io.micrometer.core.instrument.simple.SimpleMeterRegistry;

class UserServicePasswordMetricsTests {
    private UserRepository userRepository;
    private MessageRepository messageRepository;
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
        messageRepository = (MessageRepository) Proxy.newProxyInstance(
                MessageRepository.class.getClassLoader(), new Class<?>[] { MessageRepository.class },
                (proxy, method, args) -> switch (method.getName()) {
                    case "save" -> args[0];
                    case "findByMessageId" -> Optional.empty();
                    case "toString" -> "InMemoryMessageRepositoryStub";
                    case "hashCode" -> System.identityHashCode(proxy);
                    case "equals" -> proxy == args[0];
                    default -> null;
                });
        MessageService messageService = new MessageService(messageRepository, meterRegistry,
                new LimitedDatabaseExecutor(meterRegistry, 10, Duration.ofSeconds(1), 10));
        userService = new UserService(userRepository, meterRegistry, passwordEncoder,
                new LimitedDatabaseExecutor(meterRegistry, 10, Duration.ofSeconds(1), 10), messageService,
                new IdempotencyService(messageService, new TestRedisService()), new TestTransactionManager());
    }

    @Test
    void encodeGaugeReturnsToZeroAfterSuccessfulHash() {
        assertTrue(userService.setNewUser("message-1", "new-user", "test-password"));

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

        assertTrue(userService.checkUser("message-2", "existing-user", "test-password"));
        assertFalse(userService.checkUser("message-3", "existing-user", "wrong-password"));

        assertInFlight("verify", 0);
        assertEquals(2, meterRegistry.get("userdata.password.operation").tag("operation", "verify").timer().count());
        assertEquals(2, meterRegistry.get("userdata.password.permit.wait").timer().count());
    }

    private void assertInFlight(String operation, double expected) {
        assertEquals(expected, meterRegistry.get("password.hash.in.flight")
                .tag("operation", operation).gauge().value());
    }

    private static final class TestTransactionManager implements PlatformTransactionManager {
        @Override
        public TransactionStatus getTransaction(TransactionDefinition definition) {
            return new SimpleTransactionStatus();
        }

        @Override
        public void commit(TransactionStatus status) {
        }

        @Override
        public void rollback(TransactionStatus status) {
        }
    }

    private static final class TestRedisService extends RedisService {
        private TestRedisService() {
            super(null);
        }

        @Override
        public String getMessageStatus(String messageId) {
            return MESSAGE_ID_UNCOMPLETED;
        }
    }

}
