package com.apex.orderprocessingworker.infrastructure.service;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.data.redis.core.ReactiveRedisTemplate;
import org.springframework.data.redis.core.script.RedisScript;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;
import reactor.test.StepVerifier;

import java.time.Duration;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class RedisLockServiceTest {

    @Mock
    private ReactiveRedisTemplate<String, String> redisTemplate;

    private RedisLockService redisLockService;

    @BeforeEach
    void setUp() {
        Duration defaultLeaseTime = Duration.ofMinutes(5);
        Duration defaultWaitTime = Duration.ofSeconds(10);
        redisLockService = new RedisLockService(redisTemplate, defaultLeaseTime, defaultWaitTime);
    }

    @Test
    void acquireLock_ShouldReturnTrueWhenLockAcquired() {
        // Given
        String lockKey = "test-lock";
        String lockValue = "test-value";

        when(redisTemplate.execute(any(RedisScript.class), anyList(), any(String.class), any(String.class)))
                .thenReturn(Flux.just(1L));

        // When & Then
        StepVerifier.create(redisLockService.acquireLock(lockKey, lockValue))
                .expectNext(true)
                .expectComplete()
                .verify();

        verify(redisTemplate).execute(any(RedisScript.class), eq(List.of(lockKey)), eq(lockValue), anyString());
    }

    @Test
    void acquireLock_ShouldReturnFalseWhenLockNotAcquired() {
        // Given
        String lockKey = "test-lock";
        String lockValue = "test-value";

        when(redisTemplate.execute(any(RedisScript.class), anyList(), any(String.class), any(String.class)))
                .thenReturn(Flux.just(0L));

        // When & Then
        StepVerifier.create(redisLockService.acquireLock(lockKey, lockValue))
                .expectNext(false)
                .expectComplete()
                .verify();

        verify(redisTemplate).execute(any(RedisScript.class), eq(List.of(lockKey)), eq(lockValue), anyString());
    }

    @Test
    void acquireLock_ShouldHandleRedisError() {
        // Given
        String lockKey = "test-lock";
        String lockValue = "test-value";

        when(redisTemplate.execute(any(RedisScript.class), anyList(), any(String.class), any(String.class)))
                .thenReturn(Flux.error(new RuntimeException("Redis connection failed")));

        // When & Then
        StepVerifier.create(redisLockService.acquireLock(lockKey, lockValue))
                .expectError(RuntimeException.class)
                .verify();
    }

    @Test
    void acquireLockWithCustomLeaseTime_ShouldUseProvidedDuration() {
        // Given
        String lockKey = "test-lock";
        String lockValue = "test-value";
        Duration customLeaseTime = Duration.ofMinutes(10);

        when(redisTemplate.execute(any(RedisScript.class), anyList(), any(String.class), any(String.class)))
                .thenReturn(Flux.just(1L));

        // When
        StepVerifier.create(redisLockService.acquireLock(lockKey, lockValue, customLeaseTime))
                .expectNext(true)
                .expectComplete()
                .verify();

        // Then
        ArgumentCaptor<String> leaseTimeCaptor = ArgumentCaptor.forClass(String.class);
        verify(redisTemplate).execute(
                any(RedisScript.class),
                eq(List.of(lockKey)),
                eq(lockValue),
                leaseTimeCaptor.capture()
        );

        assertEquals("600", leaseTimeCaptor.getValue()); // 10 minutes = 600 seconds
    }

    @Test
    void releaseLock_ShouldReturnTrueWhenLockReleased() {
        // Given
        String lockKey = "test-lock";
        String lockValue = "test-value";

        when(redisTemplate.execute(any(RedisScript.class), anyList(), any(String.class)))
                .thenReturn(Flux.just(1L));

        // When & Then
        StepVerifier.create(redisLockService.releaseLock(lockKey, lockValue))
                .expectNext(true)
                .expectComplete()
                .verify();

        verify(redisTemplate).execute(any(RedisScript.class), eq(List.of(lockKey)), eq(lockValue));
    }

    @Test
    void releaseLock_ShouldHandleRedisError() {
        // Given
        String lockKey = "test-lock";
        String lockValue = "test-value";

        when(redisTemplate.execute(any(RedisScript.class), anyList(), any(String.class)))
                .thenReturn(Flux.error(new RuntimeException("Redis connection failed")));

        // When & Then
        StepVerifier.create(redisLockService.releaseLock(lockKey, lockValue))
                .expectError(RuntimeException.class)
                .verify();
    }

    @Test
    void isLocked_ShouldReturnTrueWhenLockExists() {
        // Given
        String lockKey = "test-lock";

        when(redisTemplate.hasKey(lockKey))
                .thenReturn(Mono.just(true));

        // When & Then
        StepVerifier.create(redisLockService.isLocked(lockKey))
                .expectNext(true)
                .expectComplete()
                .verify();

        verify(redisTemplate).hasKey(lockKey);
    }

    @Test
    void isLocked_ShouldReturnFalseWhenLockDoesNotExist() {
        // Given
        String lockKey = "test-lock";

        when(redisTemplate.hasKey(lockKey))
                .thenReturn(Mono.just(false));

        // When & Then
        StepVerifier.create(redisLockService.isLocked(lockKey))
                .expectNext(false)
                .expectComplete()
                .verify();

        verify(redisTemplate).hasKey(lockKey);
    }


    @Test
    void withLock_ShouldFailWhenLockNotAcquired() {
        // Given
        String lockKey = "test-lock";
        String lockValue = "test-value";
        Mono<Void> operation = Mono.empty();

        when(redisTemplate.execute(any(RedisScript.class), anyList(), any(String.class), any(String.class)))
                .thenReturn(Flux.just(0L)); // Lock not acquired

        // When & Then
        StepVerifier.create(redisLockService.withLock(lockKey, lockValue, operation))
                .expectErrorMatches(throwable ->
                        throwable instanceof RuntimeException &&
                                throwable.getMessage().contains("Could not acquire lock: " + lockKey))
                .verify();

        // Verify lock was attempted but operation wasn't executed
        verify(redisTemplate, times(1)).execute(any(RedisScript.class), anyList(), anyString(), anyString());
    }

    @Test
    void withLockCustomLeaseTime_ShouldUseProvidedDuration() {
        // Given
        String lockKey = "test-lock";
        String lockValue = "test-value";
        Duration customLeaseTime = Duration.ofMinutes(2);
        Mono<Void> operation = Mono.empty();

        when(redisTemplate.execute(any(RedisScript.class), anyList(), any(String.class), any(String.class)))
                .thenReturn(Flux.just(1L)); // Lock acquired
        when(redisTemplate.execute(any(RedisScript.class), anyList(), any(String.class)))
                .thenReturn(Flux.just(1L)); // Lock released

        // When
        StepVerifier.create(redisLockService.withLock(lockKey, lockValue, customLeaseTime, operation))
                .expectComplete()
                .verify();

        // Then
        ArgumentCaptor<String> leaseTimeCaptor = ArgumentCaptor.forClass(String.class);
        verify(redisTemplate).execute(
                any(RedisScript.class),
                eq(List.of(lockKey)),
                eq(lockValue),
                leaseTimeCaptor.capture()
        );

        assertEquals("120", leaseTimeCaptor.getValue()); // 2 minutes = 120 seconds
    }
}
