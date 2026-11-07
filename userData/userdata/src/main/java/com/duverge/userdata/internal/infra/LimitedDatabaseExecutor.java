package com.duverge.userdata.internal.infra;

import java.time.Duration;
import java.util.concurrent.Semaphore;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.function.Supplier;

import com.duverge.userdata.error.DatabaseCapacityUnavailableException;

import io.micrometer.core.instrument.Counter;
import io.micrometer.core.instrument.Gauge;
import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.Timer;

/** Bounds concurrent database operations and the number/time of waiting callers. */
public class LimitedDatabaseExecutor {

    private final Semaphore permits;
    private final int maxWaiters;
    private final long acquireTimeoutMillis;
    private final AtomicInteger active = new AtomicInteger();
    private final AtomicInteger activeMax = new AtomicInteger();
    private final AtomicInteger waiting = new AtomicInteger();
    private final Timer waitTimer;
    private final Counter timeoutCounter;
    private final Counter queueFullCounter;

    public LimitedDatabaseExecutor(
            MeterRegistry registry,
            int maxConcurrency,
            Duration acquireTimeout,
            int maxWaiters) {

        // Check arguments before creation of the Semaphore
        if (maxConcurrency < 1) {
            throw new IllegalArgumentException("database max-concurrency must be positive");
        }
        if (acquireTimeout.isNegative() || acquireTimeout.isZero() || acquireTimeout.toMillis() < 1) {
            throw new IllegalArgumentException("database acquire-timeout must be at least 1ms");
        }
        if (maxWaiters < 0) {
            throw new IllegalArgumentException("database max-waiters must be zero or positive");
        }

        // Create the semaphore with fairness 
        // to ensure that waiting threads are granted permits in the order
        // they requested them. 
        // This prevents starvation of waiting threads.
        this.permits = new Semaphore(maxConcurrency, true);
        this.maxWaiters = maxWaiters;
        this.acquireTimeoutMillis = acquireTimeout.toMillis();

        // Register metrics
        Gauge.builder("userdata.database.active", active, AtomicInteger::get)
                .description("Database operations currently executing")
                .register(registry);
        Gauge.builder("userdata.database.active.max", activeMax, AtomicInteger::get)
                .description("Maximum simultaneous database operations since process start")
                .register(registry);
        Gauge.builder("userdata.database.waiting", waiting, AtomicInteger::get)
                .description("Database operations waiting for an admission permit")
                .register(registry);

        // Register timers and counters for waiting time, timeouts, and queue full events
        this.waitTimer = Timer.builder("userdata.database.permit.wait")
                .description("Time spent waiting for a database admission permit")
                .register(registry);
        this.timeoutCounter = Counter.builder("userdata.database.permit.timeout")
                .description("Database operations rejected after permit acquisition timed out")
                .register(registry);
        this.queueFullCounter = Counter.builder("userdata.database.permit.queue.full")
                .description("Database operations rejected because the bounded waiting queue was full")
                .register(registry);
    }

    // Executes the given operation, enforcing concurrency limits and waiting timeouts.
    public <T> T execute(Supplier<T> operation) {
        long waitStarted = System.nanoTime();
        // Attempt to acquire a permit immediately without waiting
        boolean acquired = tryAcquireImmediately();

        // If immediate acquisition fails, 
        // attempt to acquire with timeout and bounded waiting
        if (!acquired) {
            // Increment the waiting count and check if we can reserve a waiting slot
            if (!reserveWaitSlot()) {
                waitTimer.record(System.nanoTime() - waitStarted, TimeUnit.NANOSECONDS);
                queueFullCounter.increment();
                throw new DatabaseCapacityUnavailableException("Database waiting limit reached");
            }
            try {
                try {
                    // Attempt to acquire a permit with timeout
                    acquired = permits.tryAcquire(acquireTimeoutMillis, TimeUnit.MILLISECONDS);
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    waitTimer.record(System.nanoTime() - waitStarted, TimeUnit.NANOSECONDS);
                    throw new DatabaseCapacityUnavailableException("Interrupted while waiting for database capacity", e);
                }
            } finally {
                // Decrement the waiting count regardless whether
                // we acquired a permit or not, to ensure accurate metrics
                waiting.decrementAndGet();
            }
        }

        waitTimer.record(System.nanoTime() - waitStarted, TimeUnit.NANOSECONDS);
        // If we still haven't acquired a permit, throw an exception
        if (!acquired) {
            timeoutCounter.increment();
            throw new DatabaseCapacityUnavailableException("Timed out waiting for database capacity");
        }

        // Increment the active count and update the maximum active count 
        int activeNow = active.incrementAndGet();
        activeMax.accumulateAndGet(activeNow, Math::max);
        try {
            // Get the result
            return operation.get();
        } finally {
            // Decrement the active count and release permit
            active.decrementAndGet();
            permits.release();
        }
    }

    // Attempts to reserve a waiting slot for a caller. 
    // Returns true if successful, false if the waiting limit is reached.
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

    private boolean tryAcquireImmediately() {
        try {
            // The untimed Semaphore.tryAcquire() ignores fairness and can
            // repeatedly bypass callers already waiting in the bounded queue.
            return permits.tryAcquire(0, TimeUnit.NANOSECONDS);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw new DatabaseCapacityUnavailableException(
                    "Interrupted while waiting for database capacity", e);
        }
    }
}
