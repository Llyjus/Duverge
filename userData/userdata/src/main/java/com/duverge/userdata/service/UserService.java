package com.duverge.userdata.service;

import org.springframework.stereotype.Service;

import com.duverge.userdata.entity.User;
import com.duverge.userdata.repository.UserRepository;

import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.Timer;
import org.springframework.security.crypto.password.PasswordEncoder;

@Service
public class UserService {

    final UserRepository userRepository;
    private final MeterRegistry meterRegistry;
    private final PasswordEncoder passwordEncoder;

    public UserService(UserRepository userRepository, MeterRegistry meterRegistry, PasswordEncoder passwordEncoder) {
        this.userRepository = userRepository;
        this.meterRegistry = meterRegistry;
        this.passwordEncoder = passwordEncoder;
    }

    public boolean setNewUser(String accountId, String password) {

        User user = new User();

        User userCheck = timedDatabase("find_by_account_id", () -> userRepository.findByAccountId(accountId));
        if (userCheck != null) {
            return false;
        }

        user.setAccountId(accountId);
        // Hash the password and set it on the user object
        user.setPassword(passwordEncoder.encode(password));

        timedDatabase("save", () -> userRepository.save(user));
        // deal with false of connection

        return true;
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

    public boolean checkUser(String accountId, String password) {
        User user = timedDatabase("find_by_account_id", () -> userRepository.findByAccountId(accountId));
        if (user == null) {
            return false;
        }
        return passwordEncoder.matches(password, user.getPassword());
    }

    public User getUser(String accountId) {
        return userRepository.findByAccountId(accountId);
    }

    public User getUserById(Long id) {
        return userRepository.findById(id).orElse(null);
    }

    private <T> T timedDatabase(String operation, java.util.function.Supplier<T> action) {
        return Timer.builder("userdata.database.operation")
                .description("User database operation duration")
                .tag("operation", operation)
                .register(meterRegistry)
                .record(action);
    }
}
