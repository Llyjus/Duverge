package com.duverge.userdata.config;

import java.time.Duration;

import com.duverge.userdata.internal.LimitedPasswordEncoder;
import io.micrometer.core.instrument.MeterRegistry;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.security.web.SecurityFilterChain;

@Configuration
public class SecurityConfig {

    @Bean
    PasswordEncoder passwordEncoder(
            MeterRegistry meterRegistry,
            @Value("${userdata.password-hash.max-concurrency:10}") int maxConcurrency,
            @Value("${userdata.password-hash.acquire-timeout:5s}") Duration acquireTimeout,
            @Value("${userdata.password-hash.max-waiters:100}") int maxWaiters) {
        return new LimitedPasswordEncoder(meterRegistry, maxConcurrency, acquireTimeout, maxWaiters);
    }

    @Bean
    SecurityFilterChain securityFilterChain(HttpSecurity http) throws Exception {
        http
                .authorizeHttpRequests(authorize -> authorize
                        .requestMatchers("/actuator/health", "/actuator/prometheus").permitAll()
                        .anyRequest().authenticated())
                .csrf(csrf -> csrf.ignoringRequestMatchers("/actuator/prometheus"));

        return http.build();
    }
}
