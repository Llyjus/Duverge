package com.duverge.userdata.service;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import com.duverge.userdata.entity.User;
import com.duverge.userdata.internal.Encoder;
import com.duverge.userdata.repository.UserRepository;
import io.micrometer.core.instrument.simple.SimpleMeterRegistry;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

class UserServicePasswordMetricsTests {
    private UserRepository userRepository;
    private SimpleMeterRegistry meterRegistry;
    private UserService userService;

    @BeforeEach
    void setUp() {
        userRepository = mock(UserRepository.class);
        meterRegistry = new SimpleMeterRegistry();
        userService = new UserService(userRepository, meterRegistry);
    }

    @Test
    void encodeGaugeReturnsToZeroAfterSuccessfulHash() {
        when(userRepository.findByAccountId("new-user")).thenReturn(null);

        userService.setNewUser("new-user", "test-password");

        assertInFlight("encode", 0);
        assertEquals(1, meterRegistry.get("userdata.password.operation").tag("operation", "encode").timer().count());
    }

    @Test
    void verifyGaugeReturnsToZeroAfterSuccessfulVerification() {
        User user = new User();
        user.setPassword(Encoder.encode("test-password"));
        when(userRepository.findByAccountId("existing-user")).thenReturn(user);

        userService.checkUser("existing-user", "test-password");

        assertInFlight("verify", 0);
        assertEquals(1, meterRegistry.get("userdata.password.operation").tag("operation", "verify").timer().count());
    }

    @Test
    void encodeAndVerifyGaugesReturnToZeroWhenHashActionThrows() {
        RuntimeException failure = new RuntimeException("simulated hash failure");

        assertThrows(RuntimeException.class, () -> userService.trackEncodeHash(() -> { throw failure; }));
        assertThrows(RuntimeException.class, () -> userService.trackVerifyHash(() -> { throw failure; }));

        assertInFlight("verify", 0);
        assertInFlight("encode", 0);
        assertInFlightMax("verify", 1);
        assertInFlightMax("encode", 1);
    }

    @Test
    void trackHashCountsAnActiveTaskAndRestoresAfterCompletion() {
        String result = userService.trackEncodeHash(() -> userService.trackEncodeHash(() -> {
            assertInFlight("encode", 2);
            assertInFlightMax("encode", 2);
            return "done";
        }));

        assertEquals("done", result);
        assertInFlight("encode", 0);
        assertInFlightMax("encode", 2);
    }

    private void assertInFlight(String operation, double expected) {
        assertEquals(expected, meterRegistry.get("password.hash.in.flight")
                .tag("operation", operation).gauge().value());
    }

    private void assertInFlightMax(String operation, double expected) {
        assertEquals(expected, meterRegistry.get("password.hash.in.flight.max")
                .tag("operation", operation).gauge().value());
    }
}
