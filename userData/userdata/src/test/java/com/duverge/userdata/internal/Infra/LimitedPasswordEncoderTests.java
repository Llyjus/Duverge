package com.duverge.userdata.internal.Infra;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.time.Duration;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

import com.duverge.userdata.error.PasswordHashUnavailableException;

import org.junit.jupiter.api.Test;

import io.micrometer.core.instrument.simple.SimpleMeterRegistry;

class LimitedPasswordEncoderTests {

    @Test
    void capsCombinedEncodeAndVerifyExecutionAndTracksWaitingRequests() throws Exception {
        var registry = new SimpleMeterRegistry();
        var encoder = new LimitedPasswordEncoder(registry, 2, Duration.ofSeconds(3), 10);
        var executor = Executors.newFixedThreadPool(5);
        var entered = new CountDownLatch(2);
        var release = new CountDownLatch(1);
        var currentlyActive = new AtomicInteger();
        var maximumObserved = new AtomicInteger();
        List<Future<?>> tasks = new ArrayList<>();

        try {
            for (int i = 0; i < 5; i++) {
                final String operation = i % 2 == 0 ? "encode" : "verify";
                tasks.add(executor.submit(() -> encoder.withPermit(operation, () -> {
                    int active = currentlyActive.incrementAndGet();
                    maximumObserved.accumulateAndGet(active, Math::max);
                    entered.countDown();
                    try {
                        try {
                            release.await();
                        } catch (InterruptedException e) {
                            Thread.currentThread().interrupt();
                            throw new RuntimeException(e);
                        }
                        return true;
                    } finally {
                        currentlyActive.decrementAndGet();
                    }
                })));
            }

            assertTrue(entered.await(2, TimeUnit.SECONDS));
            awaitGauge(registry, "password.hash.waiting", 3);
            assertEquals(2, registry.get("password.hash.active").gauge().value());
            assertEquals(3, registry.get("password.hash.waiting").gauge().value());
            assertEquals(2, registry.get("password.hash.active.max").gauge().value());
            assertEquals(2, maximumObserved.get());
            release.countDown();
            for (Future<?> task : tasks) {
                task.get(2, TimeUnit.SECONDS);
            }
            assertEquals(0, registry.get("password.hash.active").gauge().value());
            assertEquals(0, registry.get("password.hash.waiting").gauge().value());
            assertEquals(5, registry.get("userdata.password.permit.wait").timer().count());
        } finally {
            release.countDown();
            executor.shutdownNow();
        }
    }

    @Test
    void releasesPermitAfterSuccessfulAndExceptionalActions() {
        var registry = new SimpleMeterRegistry();
        var encoder = new LimitedPasswordEncoder(registry, 1, Duration.ofSeconds(1), 2);

        assertEquals("ok", encoder.withPermit("encode", () -> "ok"));
        assertThrows(IllegalStateException.class,
                () -> encoder.withPermit("verify", () -> { throw new IllegalStateException("expected"); }));

        assertEquals(0, registry.get("password.hash.active").gauge().value());
        assertEquals(0, registry.get("password.hash.waiting").gauge().value());
        assertEquals(1, registry.get("password.hash.active.max").gauge().value());
        assertEquals(1, registry.get("password.hash.in.flight.max").tag("operation", "encode").gauge().value());
        assertEquals(1, registry.get("password.hash.in.flight.max").tag("operation", "verify").gauge().value());
        assertEquals(1, registry.get("userdata.password.operation").tag("operation", "encode").timer().count());
        assertEquals(1, registry.get("userdata.password.operation").tag("operation", "verify").timer().count());
        // A released permit can be acquired again after either result.
        assertTrue(encoder.withPermit("encode", () -> true));
    }

    @Test
    void timesOutAndCountsFailureWithoutLeakingWaitingOrActiveState() throws Exception {
        var registry = new SimpleMeterRegistry();
        var encoder = new LimitedPasswordEncoder(registry, 1, Duration.ofMillis(100), 2);
        var executor = Executors.newSingleThreadExecutor();
        var entered = new CountDownLatch(1);
        var release = new CountDownLatch(1);

        try {
            Future<?> holder = executor.submit(() -> encoder.withPermit("encode", () -> {
                entered.countDown();
                try {
                    release.await();
                    return true;
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    throw new RuntimeException(e);
                }
            }));
            assertTrue(entered.await(1, TimeUnit.SECONDS));

            assertThrows(PasswordHashUnavailableException.class,
                    () -> encoder.withPermit("verify", () -> true));
            assertEquals(1, registry.get("password.hash.permit.timeout").counter().count());
            assertEquals(0, registry.get("password.hash.waiting").gauge().value());
            assertEquals(1, registry.get("password.hash.active").gauge().value());
            release.countDown();
            holder.get(1, TimeUnit.SECONDS);
            assertEquals(0, registry.get("password.hash.active").gauge().value());
        } finally {
            release.countDown();
            executor.shutdownNow();
        }
    }

    @Test
    void rejectsWhenBoundedWaitingAreaIsFull() throws Exception {
        var registry = new SimpleMeterRegistry();
        var encoder = new LimitedPasswordEncoder(registry, 1, Duration.ofSeconds(2), 1);
        var executor = Executors.newFixedThreadPool(2);
        var entered = new CountDownLatch(1);
        var release = new CountDownLatch(1);

        try {
            Future<?> holder = executor.submit(() -> encoder.withPermit("encode", () -> {
                entered.countDown();
                try {
                    release.await();
                    return true;
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    throw new RuntimeException(e);
                }
            }));
            assertTrue(entered.await(1, TimeUnit.SECONDS));
            Future<?> queued = executor.submit(() -> encoder.withPermit("verify", () -> true));
            awaitGauge(registry, "password.hash.waiting", 1);

            assertThrows(PasswordHashUnavailableException.class,
                    () -> encoder.withPermit("encode", () -> true));
            assertEquals(1, registry.get("password.hash.permit.queue.full").counter().count());
            assertEquals(1, registry.get("password.hash.waiting").gauge().value());
            release.countDown();
            holder.get(1, TimeUnit.SECONDS);
            queued.get(1, TimeUnit.SECONDS);
            assertEquals(0, registry.get("password.hash.waiting").gauge().value());
        } finally {
            release.countDown();
            executor.shutdownNow();
        }
    }

    private static void awaitGauge(SimpleMeterRegistry registry, String name, double expected) throws InterruptedException {
        long deadline = System.nanoTime() + TimeUnit.SECONDS.toNanos(1);
        while (System.nanoTime() < deadline) {
            if (registry.get(name).gauge().value() == expected) {
                return;
            }
            Thread.sleep(5);
        }
        assertEquals(expected, registry.get(name).gauge().value());
    }
}
