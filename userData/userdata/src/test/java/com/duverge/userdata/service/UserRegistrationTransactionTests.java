package com.duverge.userdata.service;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertNotEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.time.Duration;
import java.util.concurrent.CyclicBarrier;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.data.jpa.test.autoconfigure.DataJpaTest;
import org.springframework.boot.test.context.TestConfiguration;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Import;
import org.springframework.dao.DataIntegrityViolationException;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

import com.duverge.userdata.entity.Message;
import com.duverge.userdata.entity.User;
import com.duverge.userdata.error.DuplicateMessageIdException;
import com.duverge.userdata.internal.infra.LimitedDatabaseExecutor;
import com.duverge.userdata.internal.infra.redis.RedisService;
import com.duverge.userdata.repository.MessageRepository;
import com.duverge.userdata.repository.UserRepository;

import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.simple.SimpleMeterRegistry;

@DataJpaTest(properties = {
        "spring.jpa.hibernate.ddl-auto=create-drop",
        "spring.jpa.show-sql=false"
})
@Import({ UserService.class, MessageService.class, IdempotencyService.class,
        UserRegistrationTransactionTests.TestBeans.class })
@Transactional(propagation = Propagation.NOT_SUPPORTED)
class UserRegistrationTransactionTests {

    @Autowired
    private UserRepository userRepository;
    @Autowired
    private MessageRepository messageRepository;
    @Autowired
    private UserService userService;
    @Autowired
    private TestPasswordEncoder testPasswordEncoder;

    @BeforeEach
    void clearDatabase() {
        messageRepository.deleteAll();
        userRepository.deleteAll();
        testPasswordEncoder.reset();
    }

    @Test
    void registrationCommitsUserAndMessageWithCreatedAt() {
        assertTrue(userService.setNewUser("register-ok", "alice", "secret"));

        User user = userRepository.findByAccountId("alice");
        Message message = messageRepository.findByMessageId("register-ok").orElseThrow();
        assertNotNull(user);
        assertTrue(new BCryptPasswordEncoder().matches("secret", user.getPassword()));
        assertNotNull(message.getCreatedAt());
        assertEquals(1, messageRepository.count());
    }

    @Test
    void messageIdConflictAfterPrecheckRollsBackTheNewUser() {
        testPasswordEncoder.runAfterNextEncode(
                () -> messageRepository.saveAndFlush(new Message("already-used-id")));

        assertThrows(DuplicateMessageIdException.class,
                () -> userService.setNewUser("already-used-id", "bob", "secret"));

        assertNull(userRepository.findByAccountId("bob"));
        assertEquals(0, userRepository.count());
        assertEquals(1, messageRepository.count());
    }

    @Test
    void passwordEncodingFailureDoesNotWriteUserOrMessage() {
        testPasswordEncoder.failNextEncoding();

        assertThrows(IllegalStateException.class,
                () -> userService.setNewUser("encode-failed", "alice", "secret"));

        assertNull(userRepository.findByAccountId("alice"));
        assertTrue(messageRepository.findByMessageId("encode-failed").isEmpty());
    }

    @Test
    void duplicateAccountReturnsFalseAndStillRecordsItsRequest() {
        assertTrue(userService.setNewUser("register-first", "alice", "secret"));

        assertFalse(userService.setNewUser("register-duplicate", "alice", "another-secret"));

        assertNotNull(messageRepository.findByMessageId("register-duplicate").orElse(null));
        assertEquals(1, userRepository.count());
        assertEquals(2, messageRepository.count());
    }

    @Test
    void loginSuccessAndFailureBothPersistMessageBeforeReturning() {
        assertTrue(userService.setNewUser("register-login-user", "alice", "secret"));

        assertTrue(userService.checkUser("login-success", "alice", "secret"));
        assertFalse(userService.checkUser("login-failure", "alice", "wrong"));
        assertFalse(userService.checkUser("login-missing-user", "nobody", "secret"));

        assertNotNull(messageRepository.findByMessageId("login-success").orElse(null));
        assertNotNull(messageRepository.findByMessageId("login-failure").orElse(null));
        assertNotNull(messageRepository.findByMessageId("login-missing-user").orElse(null));
        assertEquals(4, messageRepository.count());
    }

    @Test
    void concurrentRegistrationsForOneAccountCreateOnlyOneUserAndBothMessages() throws Exception {
        testPasswordEncoder.synchronizeNextTwoEncodes();
        ExecutorService workers = Executors.newFixedThreadPool(2);
        try {
            var first = workers.submit(() -> userService.setNewUser("concurrent-1", "alice", "secret"));
            var second = workers.submit(() -> userService.setNewUser("concurrent-2", "alice", "secret"));

            boolean firstCreated = first.get(15, TimeUnit.SECONDS);
            boolean secondCreated = second.get(15, TimeUnit.SECONDS);

            assertNotEquals(firstCreated, secondCreated);
            assertEquals(1, userRepository.count());
            assertEquals(2, messageRepository.count());
            assertNotNull(messageRepository.findByMessageId("concurrent-1").orElse(null));
            assertNotNull(messageRepository.findByMessageId("concurrent-2").orElse(null));
        } finally {
            testPasswordEncoder.clearEncodeBarrier();
            workers.shutdownNow();
        }
    }

    @Test
    void reusingMessageIdIsReportedAsMessageConflictAndDoesNotMasqueradeAsDuplicateAccount() {
        assertTrue(userService.setNewUser("same-id", "alice", "secret"));
        testPasswordEncoder.failNextEncoding();

        assertThrows(DuplicateMessageIdException.class,
                () -> userService.setNewUser("same-id", "alice", "secret"));
        assertThrows(DuplicateMessageIdException.class,
                () -> userService.checkUser("same-id", "alice", "secret"));

        assertEquals(1, userRepository.count());
        assertEquals(1, messageRepository.count());
    }

    @TestConfiguration(proxyBeanMethods = false)
    static class TestBeans {
        @Bean
        MeterRegistry testMeterRegistry() {
            return new SimpleMeterRegistry();
        }

        @Bean
        LimitedDatabaseExecutor limitedDatabaseExecutor(MeterRegistry meterRegistry) {
            return new LimitedDatabaseExecutor(meterRegistry, 10, Duration.ofSeconds(1), 10);
        }

        @Bean
        TestPasswordEncoder passwordEncoder() {
            return new TestPasswordEncoder();
        }

        @Bean
        RedisService redisService() {
            return new TestRedisService();
        }
    }

    static class TestRedisService extends RedisService {
        TestRedisService() {
            super(null);
        }

        @Override
        public String getMessageStatus(String messageId) {
            return MESSAGE_ID_UNCOMPLETED;
        }
    }

    static class TestPasswordEncoder implements org.springframework.security.crypto.password.PasswordEncoder {
        private final BCryptPasswordEncoder delegate = new BCryptPasswordEncoder();
        private final AtomicBoolean failNextEncode = new AtomicBoolean();
        private volatile CyclicBarrier encodeBarrier;
        private volatile Runnable afterNextEncode;

        void failNextEncoding() {
            failNextEncode.set(true);
        }

        void synchronizeNextTwoEncodes() {
            encodeBarrier = new CyclicBarrier(2);
        }

        void clearEncodeBarrier() {
            encodeBarrier = null;
        }

        void runAfterNextEncode(Runnable action) {
            afterNextEncode = action;
        }

        void reset() {
            failNextEncode.set(false);
            encodeBarrier = null;
            afterNextEncode = null;
        }

        @Override
        public String encode(CharSequence rawPassword) {
            if (failNextEncode.compareAndSet(true, false)) {
                throw new IllegalStateException("simulated password encoder failure");
            }
            CyclicBarrier barrier = encodeBarrier;
            if (barrier != null) {
                try {
                    barrier.await(10, TimeUnit.SECONDS);
                } catch (Exception e) {
                    throw new IllegalStateException("password encoder test barrier failed", e);
                }
            }
            String encoded = delegate.encode(rawPassword);
            Runnable action = afterNextEncode;
            if (action != null) {
                afterNextEncode = null;
                action.run();
            }
            return encoded;
        }

        @Override
        public boolean matches(CharSequence rawPassword, String encodedPassword) {
            return delegate.matches(rawPassword, encodedPassword);
        }
    }
}
