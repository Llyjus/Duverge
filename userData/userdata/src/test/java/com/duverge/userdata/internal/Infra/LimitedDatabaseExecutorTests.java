package com.duverge.userdata.internal.Infra;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.time.Duration;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;

import com.duverge.userdata.error.DatabaseCapacityUnavailableException;

import org.junit.jupiter.api.Test;

import io.micrometer.core.instrument.simple.SimpleMeterRegistry;

class LimitedDatabaseExecutorTests {

    @Test
    void capsActiveOperationsAndRejectsWhenWaitingQueueIsFull() throws Exception {
        var registry = new SimpleMeterRegistry();
        var executor = new LimitedDatabaseExecutor(registry, 1, Duration.ofSeconds(2), 1);
        var started = new CountDownLatch(1);
        var release = new CountDownLatch(1);
        var active = new AtomicInteger();
        var activeMax = new AtomicInteger();

        try (ExecutorService workers = Executors.newFixedThreadPool(2)) {
            Future<?> first = workers.submit(() -> executor.execute(() -> {
                int now = active.incrementAndGet();
                activeMax.accumulateAndGet(now, Math::max);
                started.countDown();
                await(release);
                active.decrementAndGet();
                return null;
            }));
            assertTrue(started.await(1, TimeUnit.SECONDS));

            Future<?> second = workers.submit(() -> executor.execute(() -> {
                int now = active.incrementAndGet();
                activeMax.accumulateAndGet(now, Math::max);
                active.decrementAndGet();
                return null;
            }));
            awaitGauge(registry, "userdata.database.waiting", 1);

            assertThrows(DatabaseCapacityUnavailableException.class,
                    () -> executor.execute(() -> null));
            assertEquals(1.0, registry.get("userdata.database.permit.queue.full").counter().count());

            release.countDown();
            first.get(1, TimeUnit.SECONDS);
            second.get(1, TimeUnit.SECONDS);
        } finally {
            release.countDown();
        }

        assertEquals(1, activeMax.get());
        assertEquals(0.0, registry.get("userdata.database.active").gauge().value());
        assertEquals(0.0, registry.get("userdata.database.waiting").gauge().value());
    }

    @Test
    void waitingOperationTimesOutAndPermitIsReleasedAfterFailure() throws Exception {
        var registry = new SimpleMeterRegistry();
        var executor = new LimitedDatabaseExecutor(registry, 1, Duration.ofMillis(100), 1);
        var started = new CountDownLatch(1);
        var release = new CountDownLatch(1);

        try (ExecutorService worker = Executors.newSingleThreadExecutor()) {
            Future<?> activeOperation = worker.submit(() -> executor.execute(() -> {
                started.countDown();
                await(release);
                return null;
            }));
            assertTrue(started.await(1, TimeUnit.SECONDS));

            assertThrows(DatabaseCapacityUnavailableException.class,
                    () -> executor.execute(() -> null));
            assertEquals(1.0, registry.get("userdata.database.permit.timeout").counter().count());
            assertEquals(0.0, registry.get("userdata.database.waiting").gauge().value());

            release.countDown();
            activeOperation.get(1, TimeUnit.SECONDS);
        } finally {
            release.countDown();
        }

        assertThrows(IllegalStateException.class,
                () -> executor.execute(() -> { throw new IllegalStateException("test failure"); }));
        assertEquals(0.0, registry.get("userdata.database.active").gauge().value());
        assertEquals("ok", executor.execute(() -> "ok"));
    }

    private static void await(CountDownLatch latch) {
        try {
            if (!latch.await(2, TimeUnit.SECONDS)) {
                throw new IllegalStateException("timed out waiting for test latch");
            }
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw new IllegalStateException("test worker interrupted", e);
        }
    }

    private static void awaitGauge(SimpleMeterRegistry registry, String name, double expected)
            throws InterruptedException {
        long deadline = System.nanoTime() + TimeUnit.SECONDS.toNanos(1);
        while (System.nanoTime() < deadline) {
            if (registry.get(name).gauge().value() == expected) {
                return;
            }
            Thread.sleep(5);
        }
        assertEquals(expected, registry.get(name).gauge().value(), "gauge did not reach expected value");
    }
}
