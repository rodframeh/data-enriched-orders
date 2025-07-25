package com.apex.orderprocessingworker.infrastructure.service;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.data.redis.core.ReactiveRedisTemplate;
import org.springframework.stereotype.Service;
import reactor.core.publisher.Mono;

import java.time.Duration;
import java.time.LocalDateTime;

@Service
public class RetryService {

    private static final Logger logger = LoggerFactory.getLogger(RetryService.class);
    private static final String RETRY_KEY_PREFIX = "retry:";
    private static final String FAILED_KEY_PREFIX = "failed:";

    private final ReactiveRedisTemplate<String, String> redisTemplate;
    private final ObjectMapper objectMapper;
    private final int maxAttempts;
    private final Duration initialDelay;
    private final Duration maxDelay;
    private final double multiplier;

    public RetryService(
            ReactiveRedisTemplate<String, String> redisTemplate,
            ObjectMapper objectMapper,
            @Value("${app.retry.max-attempts}") int maxAttempts,
            @Value("${app.retry.initial-delay}") Duration initialDelay,
            @Value("${app.retry.max-delay}") Duration maxDelay,
            @Value("${app.retry.multiplier}") double multiplier) {
        this.redisTemplate = redisTemplate;
        this.objectMapper = objectMapper;
        this.maxAttempts = maxAttempts;
        this.initialDelay = initialDelay;
        this.maxDelay = maxDelay;
        this.multiplier = multiplier;
    }

    public Mono<Void> storeFailedMessage(String orderId, String message, String errorMessage) {
        FailedMessage failedMessage = new FailedMessage(
                orderId,
                message,
                errorMessage,
                1,
                LocalDateTime.now(),
                LocalDateTime.now()
        );

        return storeMessage(RETRY_KEY_PREFIX + orderId, failedMessage)
                .doOnSuccess(v -> logger.info("Stored failed message for retry: {}", orderId))
                .doOnError(error -> logger.error("Error storing failed message {}: {}", orderId, error.getMessage()));
    }

    public Mono<Void> incrementRetryCount(String orderId) {
        String key = RETRY_KEY_PREFIX + orderId;

        return getFailedMessage(key)
                .flatMap(failedMessage -> {
                    if (failedMessage.attemptCount() >= maxAttempts) {
                        // Move to permanent failed storage
                        return moveToFailedStorage(orderId, failedMessage)
                                .then(deleteRetryMessage(key));
                    } else {
                        // Increment retry count
                        FailedMessage updatedMessage = new FailedMessage(
                                failedMessage.orderId(),
                                failedMessage.originalMessage(),
                                failedMessage.errorMessage(),
                                failedMessage.attemptCount() + 1,
                                failedMessage.firstFailedAt(),
                                LocalDateTime.now()
                        );

                        Duration nextDelay = calculateNextDelay(updatedMessage.attemptCount());

                        return storeMessage(key, updatedMessage)
                                .then(redisTemplate.expire(key, nextDelay))
                                .then();
                    }
                })
                .doOnSuccess(v -> logger.debug("Incremented retry count for order: {}", orderId))
                .doOnError(error -> logger.error("Error incrementing retry count for {}: {}", orderId, error.getMessage()));
    }

    public Mono<Boolean> shouldRetry(String orderId) {
        return redisTemplate.hasKey(RETRY_KEY_PREFIX + orderId)
                .doOnSuccess(exists -> logger.debug("Order {} should retry: {}", orderId, exists));
    }

    public Mono<FailedMessage> getRetryMessage(String orderId) {
        return getFailedMessage(RETRY_KEY_PREFIX + orderId);
    }

    private Mono<FailedMessage> getFailedMessage(String key) {
        return redisTemplate.opsForValue().get(key)
                .map(json -> {
                    try {
                        return objectMapper.readValue(json, FailedMessage.class);
                    } catch (JsonProcessingException e) {
                        throw new RuntimeException("Error deserializing failed message", e);
                    }
                });
    }

    private Mono<Void> storeMessage(String key, FailedMessage message) {
        try {
            String json = objectMapper.writeValueAsString(message);
            return redisTemplate.opsForValue().set(key, json).then();
        } catch (JsonProcessingException e) {
            return Mono.error(new RuntimeException("Error serializing failed message", e));
        }
    }

    private Mono<Void> moveToFailedStorage(String orderId, FailedMessage failedMessage) {
        logger.warn("Moving order {} to permanent failed storage after {} attempts",
                orderId, failedMessage.attemptCount());

        return storeMessage(FAILED_KEY_PREFIX + orderId, failedMessage)
                .doOnSuccess(v -> logger.info("Moved order {} to permanent failed storage", orderId));
    }

    private Mono<Void> deleteRetryMessage(String key) {
        return redisTemplate.delete(key).then();
    }

    private Duration calculateNextDelay(int attemptCount) {
        long delayMs = (long) (initialDelay.toMillis() * Math.pow(multiplier, attemptCount - 1));
        Duration delay = Duration.ofMillis(Math.min(delayMs, maxDelay.toMillis()));

        logger.debug("Calculated next delay for attempt {}: {}ms", attemptCount, delay.toMillis());
        return delay;
    }

    public record FailedMessage(
            String orderId,
            String originalMessage,
            String errorMessage,
            int attemptCount,
            LocalDateTime firstFailedAt,
            LocalDateTime lastAttemptAt
    ) {
    }
}
