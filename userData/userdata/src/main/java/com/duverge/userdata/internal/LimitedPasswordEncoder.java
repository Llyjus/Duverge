package com.duverge.userdata.internal;

import java.time.Duration;
import java.util.concurrent.Semaphore;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.function.Supplier;

import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;

import io.micrometer.core.instrument.Counter;
import io.micrometer.core.instrument.Gauge;
import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.Timer;

/** Central admission control for all password encode and verify operations. */
public class LimitedPasswordEncoder implements PasswordEncoder {

    private final PasswordEncoder delegate = new BCryptPasswordEncoder();
    private final Semaphore permits;
    private final int maxWaiters;
    private final long acquireTimeoutMillis;
    private final AtomicInteger active = new AtomicInteger();
    private final AtomicInteger activeMax = new AtomicInteger();
    private final AtomicInteger waiting = new AtomicInteger();
    private final AtomicInteger encodeActive = new AtomicInteger();
    private final AtomicInteger verifyActive = new AtomicInteger();
    private final AtomicInteger encodeActiveMax = new AtomicInteger();
    private final AtomicInteger verifyActiveMax = new AtomicInteger();
    private final Timer waitTimer;
    private final Timer encodeTimer;
    private final Timer verifyTimer;
    private final Counter timeoutCounter;
    private final Counter queueFullCounter;

    public LimitedPasswordEncoder(
            MeterRegistry registry,
            int maxConcurrency,
            Duration acquireTimeout,
            int maxWaiters) {
        if (maxConcurrency < 0) {
            throw new IllegalArgumentException("max-concurrency must be zero (unlimited) or positive");
        }
        if (acquireTimeout.isNegative() || acquireTimeout.isZero() || acquireTimeout.toMillis() < 1) {
            throw new IllegalArgumentException("acquire-timeout must be at least 1ms");
        }
        if (maxWaiters < 0) {
            throw new IllegalArgumentException("max-waiters must be zero or positive");
        }

        // Zero is reserved for the no-limit benchmark baseline.
        this.permits = new Semaphore(maxConcurrency == 0 ? Integer.MAX_VALUE : maxConcurrency, true);
        this.maxWaiters = maxWaiters;
        this.acquireTimeoutMillis = acquireTimeout.toMillis();

        Gauge.builder("password.hash.active", active, AtomicInteger::get)
                .description("BCrypt operations currently executing")
                .register(registry);
        Gauge.builder("password.hash.active.max", activeMax, AtomicInteger::get)
                .description("Maximum simultaneous BCrypt operations since process start")
                .register(registry);
        Gauge.builder("password.hash.waiting", waiting, AtomicInteger::get)
                .description("Requests currently waiting for a BCrypt permit")
                .register(registry);
        registerOperationGauges(registry, "encode", encodeActive, encodeActiveMax);
        registerOperationGauges(registry, "verify", verifyActive, verifyActiveMax);

        this.waitTimer = Timer.builder("userdata.password.permit.wait")
                .description("Time spent acquiring a BCrypt execution permit")
                .register(registry);
        this.encodeTimer = Timer.builder("userdata.password.operation")
                .description("Password hash and verification execution duration")
                .tag("operation", "encode")
                .register(registry);
        this.verifyTimer = Timer.builder("userdata.password.operation")
                .description("Password hash and verification execution duration")
                .tag("operation", "verify")
                .register(registry);
        this.timeoutCounter = Counter.builder("password.hash.permit.timeout")
                .description("BCrypt operations rejected after permit acquisition timed out")
                .register(registry);
        this.queueFullCounter = Counter.builder("password.hash.permit.queue.full")
                .description("BCrypt operations rejected because the waiting limit was reached")
                .register(registry);
    }

    @Override
    public String encode(CharSequence rawPassword) {
        return withPermit("encode", () -> delegate.encode(rawPassword));
    }

    @Override
    public boolean matches(CharSequence rawPassword, String encodedPassword) {
        return withPermit("verify", () -> delegate.matches(rawPassword, encodedPassword));
    }

    <T> T withPermit(String operation, Supplier<T> action) {
        AtomicInteger operationActive = operationGauge(operation);
        AtomicInteger operationMax = operationMaxGauge(operation);
        Timer executionTimer = operationTimer(operation);
        long waitStarted = System.nanoTime();
        boolean acquired = permits.tryAcquire();

        if (!acquired) {
            if (!reserveWaitSlot()) {
                waitTimer.record(System.nanoTime() - waitStarted, TimeUnit.NANOSECONDS);
                queueFullCounter.increment();
                throw new PasswordHashUnavailableException("Password hash waiting limit reached");
            }
            try {
                try {
                    acquired = permits.tryAcquire(acquireTimeoutMillis, TimeUnit.MILLISECONDS);
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    waitTimer.record(System.nanoTime() - waitStarted, TimeUnit.NANOSECONDS);
                    throw new PasswordHashUnavailableException("Interrupted while waiting for password hash capacity", e);
                }
            } finally {
                waiting.decrementAndGet();
            }
        }

        waitTimer.record(System.nanoTime() - waitStarted, TimeUnit.NANOSECONDS);
        if (!acquired) {
            timeoutCounter.increment();
            throw new PasswordHashUnavailableException("Timed out waiting for password hash capacity");
        }

        int activeNow = active.incrementAndGet();
        operationActive.incrementAndGet();
        activeMax.accumulateAndGet(activeNow, Math::max);
        operationMax.accumulateAndGet(operationActive.get(), Math::max);
        try {
            return executionTimer.record(action);
        } finally {
            operationActive.decrementAndGet();
            active.decrementAndGet();
            permits.release();
        }
    }

    private void registerOperationGauges(MeterRegistry registry, String operation,
            AtomicInteger current, AtomicInteger maximum) {
        Gauge.builder("password.hash.in.flight", current, AtomicInteger::get)
                .tag("operation", operation)
                .description("BCrypt operations currently executing by operation")
                .register(registry);
        Gauge.builder("password.hash.in.flight.max", maximum, AtomicInteger::get)
                .tag("operation", operation)
                .description("Maximum simultaneous BCrypt operations by operation")
                .register(registry);
    }

    private boolean reserveWaitSlot() {
        if (maxWaiters == 0) {
            return false;
        }
        while (true) {
            int current = waiting.get();
            if (current >= maxWaiters) {
                return false;
            }
            if (waiting.compareAndSet(current, current + 1)) {
                return true;
            }
        }
    }

    private AtomicInteger operationGauge(String operation) {
        return switch (operation) {
            case "encode" -> encodeActive;
            case "verify" -> verifyActive;
            default -> throw new IllegalArgumentException("Unsupported password operation: " + operation);
        };
    }

    private AtomicInteger operationMaxGauge(String operation) {
        return switch (operation) {
            case "encode" -> encodeActiveMax;
            case "verify" -> verifyActiveMax;
            default -> throw new IllegalArgumentException("Unsupported password operation: " + operation);
        };
    }

    private Timer operationTimer(String operation) {
        return switch (operation) {
            case "encode" -> encodeTimer;
            case "verify" -> verifyTimer;
            default -> throw new IllegalArgumentException("Unsupported password operation: " + operation);
        };
    }
}
