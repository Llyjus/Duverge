package com.duverge.userdata.service;

import java.util.concurrent.atomic.AtomicInteger;

import org.springframework.stereotype.Service;

import com.duverge.userdata.entity.User;
import com.duverge.userdata.internal.Encoder;
import com.duverge.userdata.repository.UserRepository;

import io.micrometer.core.instrument.Gauge;
import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.Timer;

@Service
public class UserService {

    final UserRepository userRepository;
    private final MeterRegistry meterRegistry;
    // Atomic counters to track in-flight password hash operations
    private final AtomicInteger encodeInFlight = new AtomicInteger();
    private final AtomicInteger verifyInFlight = new AtomicInteger();
    private final AtomicInteger encodeInFlightMax = new AtomicInteger();
    private final AtomicInteger verifyInFlightMax = new AtomicInteger();

    public UserService(UserRepository userRepository, MeterRegistry meterRegistry) {
        this.userRepository = userRepository;
        this.meterRegistry = meterRegistry;

        // Register gauges for in-flight password hash operations
        Gauge.builder("password.hash.in.flight", encodeInFlight, AtomicInteger::get)
                .tag("operation", "encode")
                .description("Password hash operations currently in progress")
                .register(meterRegistry);
        Gauge.builder("password.hash.in.flight", verifyInFlight, AtomicInteger::get)
                .tag("operation", "verify")
                .description("Password hash operations currently in progress")
                .register(meterRegistry);
        Gauge.builder("password.hash.in.flight.max", encodeInFlightMax, AtomicInteger::get)
                .tag("operation", "encode")
                .description("Maximum concurrent password hash operations since service start")
                .register(meterRegistry);
        Gauge.builder("password.hash.in.flight.max", verifyInFlightMax, AtomicInteger::get)
                .tag("operation", "verify")
                .description("Maximum concurrent password hash operations since service start")
                .register(meterRegistry);
    }

    public boolean setNewUser(String accountId, String password) {

        User user = new User();

        User userCheck = timedDatabase("find_by_account_id", () -> userRepository.findByAccountId(accountId));
        if (userCheck != null) {
            return false;
        }

        user.setAccountId(accountId);
        // Hash the password and set it on the user object
        user.setPassword(timedPassword("encode", () -> trackEncodeHash(() -> Encoder.encode(password))));

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
        return timedPassword("verify", () -> trackVerifyHash(() -> Encoder.matches(password, user.getPassword())));
    }

    public User getUser(String accountId) {
        return userRepository.findByAccountId(accountId);
    }

    public User getUserById(Long id) {
        return userRepository.findById(id).orElse(null);
    }

    private <T> T timedPassword(String operation, java.util.function.Supplier<T> action) {
        return Timer.builder("userdata.password.operation")
                .description("Password hash and verification duration")
                .tag("operation", operation)
                .register(meterRegistry)
                .record(action);
    }

    // Track the number of in-flight password hash operations
    <T> T trackEncodeHash(java.util.function.Supplier<T> action) {
        return trackHash(encodeInFlight, encodeInFlightMax, action);
    }

    <T> T trackVerifyHash(java.util.function.Supplier<T> action) {
        return trackHash(verifyInFlight, verifyInFlightMax, action);
    }

    private <T> T trackHash(AtomicInteger inFlight, AtomicInteger inFlightMax,
            java.util.function.Supplier<T> action) {
        int active = inFlight.incrementAndGet();
        inFlightMax.accumulateAndGet(active, Math::max);
        try {
            return action.get();
        } finally {
            inFlight.decrementAndGet();
        }
    }

    private <T> T timedDatabase(String operation, java.util.function.Supplier<T> action) {
        return Timer.builder("userdata.database.operation")
                .description("User database operation duration")
                .tag("operation", operation)
                .register(meterRegistry)
                .record(action);
    }
}
