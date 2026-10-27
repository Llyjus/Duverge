package com.duverge.userdata.config;

import java.time.Duration;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import com.duverge.userdata.internal.Infra.LimitedDatabaseExecutor;

import io.micrometer.core.instrument.MeterRegistry;

@Configuration
public class DatabaseAdmissionConfig {

    @Bean
    LimitedDatabaseExecutor limitedDatabaseExecutor(
            MeterRegistry meterRegistry,
            @Value("${userdata.database.max-concurrency:10}") int maxConcurrency,
            @Value("${userdata.database.acquire-timeout:1s}") Duration acquireTimeout,
            @Value("${userdata.database.max-waiters:10}") int maxWaiters) {
        return new LimitedDatabaseExecutor(meterRegistry, maxConcurrency, acquireTimeout, maxWaiters);
    }
}
