package com.apex.orderprocessingworker.infrastructure.service;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.data.redis.core.ReactiveRedisTemplate;
import org.springframework.data.redis.core.ReactiveValueOperations;
import reactor.core.publisher.Mono;
import reactor.test.StepVerifier;

import java.time.Duration;
import java.time.LocalDateTime;

import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class RetryServiceTest {

    @Mock
    private ReactiveRedisTemplate<String, String> redisTemplate;

    @Mock
    private ReactiveValueOperations<String, String> valueOperations;

    private RetryService retryService;
    private ObjectMapper objectMapper;

    @BeforeEach
    void setUp() {
        objectMapper = new ObjectMapper();
        objectMapper.findAndRegisterModules(); // For LocalDateTime serialization

        int maxAttempts = 3;
        Duration initialDelay = Duration.ofSeconds(1);
        Duration maxDelay = Duration.ofMinutes(5);
        double multiplier = 2.0;

        retryService = new RetryService(
                redisTemplate,
                objectMapper,
                maxAttempts,
                initialDelay,
                maxDelay,
                multiplier
        );

        when(redisTemplate.opsForValue()).thenReturn(valueOperations);
    }

    @Test
    void storeFailedMessage_ShouldStoreMessageInRedis() {
        // Given
        String orderId = "order-123";
        String message = "original message";
        String errorMessage = "error occurred";

        when(valueOperations.set(anyString(), anyString()))
                .thenReturn(Mono.just(true));

        // When & Then
        StepVerifier.create(retryService.storeFailedMessage(orderId, message, errorMessage))
                .expectComplete()
                .verify();

        verify(valueOperations).set(eq("retry:order-123"), anyString());
    }

    @Test
    void incrementRetryCount_ShouldIncrementWhenBelowMaxAttempts() throws Exception {
        // Given
        String orderId = "order-123";
        RetryService.FailedMessage existingMessage = new RetryService.FailedMessage(
                orderId,
                "original message",
                "error",
                1,
                LocalDateTime.now().minusHours(1),
                LocalDateTime.now().minusHours(1)
        );

        String existingJson = objectMapper.writeValueAsString(existingMessage);

        when(valueOperations.get("retry:order-123"))
                .thenReturn(Mono.just(existingJson));
        when(valueOperations.set(anyString(), anyString()))
                .thenReturn(Mono.just(true));
        when(redisTemplate.expire(anyString(), any(Duration.class)))
                .thenReturn(Mono.just(true));

        // When & Then
        StepVerifier.create(retryService.incrementRetryCount(orderId))
                .expectComplete()
                .verify();

        verify(valueOperations).set(eq("retry:order-123"), anyString());
        verify(redisTemplate).expire(eq("retry:order-123"), any(Duration.class));
    }

    @Test
    void incrementRetryCount_ShouldMoveToFailedStorageWhenMaxAttemptsReached() throws Exception {
        // Given
        String orderId = "order-123";
        RetryService.FailedMessage maxAttemptsMessage = new RetryService.FailedMessage(
                orderId,
                "original message",
                "error",
                3, // max attempts reached
                LocalDateTime.now().minusHours(1),
                LocalDateTime.now().minusMinutes(5)
        );

        String existingJson = objectMapper.writeValueAsString(maxAttemptsMessage);

        when(valueOperations.get("retry:order-123"))
                .thenReturn(Mono.just(existingJson));
        when(valueOperations.set(anyString(), anyString()))
                .thenReturn(Mono.just(true));
        when(redisTemplate.delete(anyString()))
                .thenReturn(Mono.just(1L));

        // When & Then
        StepVerifier.create(retryService.incrementRetryCount(orderId))
                .expectComplete()
                .verify();

        verify(valueOperations).set(eq("failed:order-123"), anyString());
        verify(redisTemplate).delete("retry:order-123");
    }

    @Test
    void incrementRetryCount_ShouldHandleMessageNotFound() {
        // Given
        String orderId = "nonexistent-order";

        when(valueOperations.get("retry:nonexistent-order"))
                .thenReturn(Mono.empty());

        // When & Then
        StepVerifier.create(retryService.incrementRetryCount(orderId))
                .expectComplete()
                .verify();

        verify(valueOperations).get("retry:nonexistent-order");
        verify(valueOperations, never()).set(anyString(), anyString());
    }

      @Test
    void getRetryMessage_ShouldReturnMessageWhenExists() throws Exception {
        // Given
        String orderId = "order-123";
        RetryService.FailedMessage expectedMessage = new RetryService.FailedMessage(
                orderId,
                "original message",
                "error",
                2,
                LocalDateTime.now().minusHours(1),
                LocalDateTime.now().minusMinutes(30)
        );

        String messageJson = objectMapper.writeValueAsString(expectedMessage);

        when(valueOperations.get("retry:order-123"))
                .thenReturn(Mono.just(messageJson));

        // When & Then
        StepVerifier.create(retryService.getRetryMessage(orderId))
                .expectNextMatches(failedMessage ->
                        failedMessage.orderId().equals(orderId) &&
                                failedMessage.attemptCount() == 2 &&
                                failedMessage.originalMessage().equals("original message"))
                .expectComplete()
                .verify();

        verify(valueOperations).get("retry:order-123");
    }

    @Test
    void getRetryMessage_ShouldReturnEmptyWhenNotExists() {
        // Given
        String orderId = "nonexistent-order";

        when(valueOperations.get("retry:nonexistent-order"))
                .thenReturn(Mono.empty());

        // When & Then
        StepVerifier.create(retryService.getRetryMessage(orderId))
                .expectComplete()
                .verify();

        verify(valueOperations).get("retry:nonexistent-order");
    }

    @Test
    void getRetryMessage_ShouldHandleDeserializationError() {
        // Given
        String orderId = "order-123";
        String invalidJson = "invalid json";

        when(valueOperations.get("retry:order-123"))
                .thenReturn(Mono.just(invalidJson));

        // When & Then
        StepVerifier.create(retryService.getRetryMessage(orderId))
                .expectError(RuntimeException.class)
                .verify();
    }

    @Test
    void incrementRetryCount_ShouldCalculateCorrectDelays() throws Exception {
        // Given
        String orderId = "order-123";

        // Test attempt 1 -> 2
        RetryService.FailedMessage attempt1 = new RetryService.FailedMessage(
                orderId, "message", "error", 1,
                LocalDateTime.now().minusHours(1),
                LocalDateTime.now().minusHours(1)
        );

        when(valueOperations.get("retry:order-123"))
                .thenReturn(Mono.just(objectMapper.writeValueAsString(attempt1)));
        when(valueOperations.set(anyString(), anyString()))
                .thenReturn(Mono.just(true));
        when(redisTemplate.expire(anyString(), any(Duration.class)))
                .thenReturn(Mono.just(true));

        // When
        StepVerifier.create(retryService.incrementRetryCount(orderId))
                .expectComplete()
                .verify();

        // Then - Should use 2 second delay for attempt 2 (1 * 2^(2-1) = 2)
        verify(redisTemplate).expire(eq("retry:order-123"), eq(Duration.ofSeconds(2)));
    }

    @Test
    void storeFailedMessage_ShouldHandleRedisError() {
        // Given
        String orderId = "order-123";

        when(valueOperations.set(anyString(), anyString()))
                .thenReturn(Mono.error(new RuntimeException("Redis connection failed")));

        // When & Then
        StepVerifier.create(retryService.storeFailedMessage(orderId, "message", "error"))
                .expectError(RuntimeException.class)
                .verify();
    }
}