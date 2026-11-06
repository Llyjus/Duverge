package com.duverge.userdata.service;

import org.springframework.dao.DataIntegrityViolationException;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.PlatformTransactionManager;
import org.springframework.transaction.support.TransactionTemplate;

import com.duverge.userdata.entity.User;
import com.duverge.userdata.error.DatabaseConstraintClassifier;
import com.duverge.userdata.internal.infra.LimitedDatabaseExecutor;
import com.duverge.userdata.repository.UserRepository;

import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.Timer;

@Service
public class UserService {

    final UserRepository userRepository;
    private final MeterRegistry meterRegistry;
    private final PasswordEncoder passwordEncoder;
    private final LimitedDatabaseExecutor databaseExecutor;
    private final MessageService messageService;
    private final IdempotencyService idempotencyService;
    private final TransactionTemplate transactionTemplate;

    public UserService(
            UserRepository userRepository,
            MeterRegistry meterRegistry,
            PasswordEncoder passwordEncoder,
            LimitedDatabaseExecutor databaseExecutor,
            MessageService messageService,
            IdempotencyService idempotencyService,
            PlatformTransactionManager transactionManager) {
        this.userRepository = userRepository;
        this.meterRegistry = meterRegistry;
        this.passwordEncoder = passwordEncoder;
        this.databaseExecutor = databaseExecutor;
        this.messageService = messageService;
        this.idempotencyService = idempotencyService;
        this.transactionTemplate = new TransactionTemplate(transactionManager);
    }

    public boolean setNewUser(String messageId, String accountId, String password) {
        idempotencyService.ensureNewMessageId(messageId);
        User existingUser = timedDatabase(
                "find_by_account_id", () -> userRepository.findByAccountId(accountId));
        if (existingUser != null) {
            persistMessage(messageId);
            return false;
        }

        // Hash before opening the database transaction so BCrypt does not hold
        // a JDBC connection while it runs.
        String encodedPassword = passwordEncoder.encode(password);
        try {
            return transactionTemplate.execute(transactionStatus -> {
                User concurrentlyCreated = timedDatabase(
                        "find_by_account_id", () -> userRepository.findByAccountId(accountId));
                if (concurrentlyCreated != null) {
                    messageService.createMessage(messageId);
                    return false;
                }

                User user = new User();
                user.setAccountId(accountId);
                user.setPassword(encodedPassword);
                timedDatabase("save", () -> userRepository.save(user));
                messageService.createMessage(messageId);
                return true;
            });
        } catch (DataIntegrityViolationException e) {
            // Only the account constraint indicates a concurrent registration.
            // Other integrity errors (including Message ID conflicts) must not
            // be mistaken for an already-registered account.
            if (!DatabaseConstraintClassifier.violates(e, "uk_users_account_id")) {
                throw e;
            }
            User concurrentlyCreated = timedDatabase(
                    "find_by_account_id", () -> userRepository.findByAccountId(accountId));
            if (concurrentlyCreated == null) {
                throw e;
            }
            persistMessage(messageId);
            return false;
        }
    }

    public boolean setUser(String accountId) {
        User user = timedDatabase("find_by_account_id", () -> userRepository.findByAccountId(accountId));
        if (user == null) {
            return false;
        }
        user.setAccountId(accountId);
        // deal with false of connection
        return true;
    }

    public boolean checkUser(String messageId, String accountId, String password) {
        idempotencyService.ensureNewMessageId(messageId);
        User user = timedDatabase("find_by_account_id", () -> userRepository.findByAccountId(accountId));
        boolean valid = user != null && passwordEncoder.matches(password, user.getPassword());

        // Login only reads the user row. Persist its message record in a short
        // transaction after password verification and before returning.
        persistMessage(messageId);
        return valid;
    }

    private void persistMessage(String messageId) {
        transactionTemplate.execute(transactionStatus -> {
            messageService.createMessage(messageId);
            return null;
        });
    }

    public User getUser(String accountId) {
        return timedDatabase("find_by_account_id", () -> userRepository.findByAccountId(accountId));
    }

    public User getUserById(Long id) {
        return timedDatabase("find_by_id", () -> userRepository.findById(id).orElse(null));
    }

    private <T> T timedDatabase(String operation, java.util.function.Supplier<T> action) {
        return databaseExecutor.execute(() -> Timer.builder("userdata.database.operation")
                .description("User database operation duration")
                .tag("operation", operation)
                .register(meterRegistry)
                .record(action));
    }
}
