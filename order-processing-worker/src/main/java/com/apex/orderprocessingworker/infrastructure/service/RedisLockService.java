package com.apex.orderprocessingworker.infrastructure.service;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.data.redis.core.ReactiveRedisTemplate;
import org.springframework.data.redis.core.script.RedisScript;
import org.springframework.stereotype.Service;
import reactor.core.publisher.Mono;

import java.time.Duration;
import java.util.List;

@Service
public class RedisLockService {

    private static final Logger logger = LoggerFactory.getLogger(RedisLockService.class);

    private final ReactiveRedisTemplate<String, String> redisTemplate;
    private final Duration defaultLeaseTime;
    private final Duration defaultWaitTime;

    // Lua script para operaciones atómicas de lock
    private static final String LOCK_SCRIPT = """
            if redis.call('exists', KEYS[1]) == 0 then
                redis.call('setex', KEYS[1], ARGV[2], ARGV[1])
                return 1
            else
                return 0
            end
            """;

    private static final String UNLOCK_SCRIPT = """
            if redis.call('get', KEYS[1]) == ARGV[1] then
                return redis.call('del', KEYS[1])
            else
                return 0
            end
            """;

    public RedisLockService(
            ReactiveRedisTemplate<String, String> redisTemplate,
            @Value("${app.lock.default-lease-time}") Duration defaultLeaseTime,
            @Value("${app.lock.default-wait-time}") Duration defaultWaitTime) {
        this.redisTemplate = redisTemplate;
        this.defaultLeaseTime = defaultLeaseTime;
        this.defaultWaitTime = defaultWaitTime;
    }

    public Mono<Boolean> acquireLock(String lockKey, String lockValue) {
        return acquireLock(lockKey, lockValue, defaultLeaseTime);
    }

    public Mono<Boolean> acquireLock(String lockKey, String lockValue, Duration leaseTime) {
        logger.debug("Attempting to acquire lock: {} with value: {}", lockKey, lockValue);

        RedisScript<Long> script = RedisScript.of(LOCK_SCRIPT, Long.class);

        return redisTemplate.execute(script,
                        List.of(lockKey),
                        lockValue,
                        String.valueOf(leaseTime.getSeconds()))
                .next()
                .map(result -> result == 1L)
                .doOnSuccess(acquired -> {
                    if (acquired) {
                        logger.debug("Lock acquired successfully: {}", lockKey);
                    } else {
                        logger.debug("Failed to acquire lock: {}", lockKey);
                    }
                })
                .doOnError(error -> logger.error("Error acquiring lock {}: {}", lockKey, error.getMessage()));
    }

    public Mono<Boolean> releaseLock(String lockKey, String lockValue) {
        logger.debug("Attempting to release lock: {} with value: {}", lockKey, lockValue);

        RedisScript<Long> script = RedisScript.of(UNLOCK_SCRIPT, Long.class);

        return redisTemplate.execute(script,
                        List.of(lockKey),
                        lockValue)
                .map(result -> result == 1L)
                .next()
                .doOnSuccess(released -> {
                    if (released) {
                        logger.debug("Lock released successfully: {}", lockKey);
                    } else {
                        logger.debug("Failed to release lock (may have expired): {}", lockKey);
                    }
                })
                .doOnError(error -> logger.error("Error releasing lock {}: {}", lockKey, error.getMessage()));
    }

    public Mono<Boolean> isLocked(String lockKey) {
        return redisTemplate.hasKey(lockKey)
                .doOnSuccess(exists -> logger.debug("Lock {} exists: {}", lockKey, exists))
                .doOnError(error -> logger.error("Error checking lock {}: {}", lockKey, error.getMessage()));
    }

    public Mono<Void> withLock(String lockKey, String lockValue, Mono<Void> operation) {
        return withLock(lockKey, lockValue, defaultLeaseTime, operation);
    }

    public Mono<Void> withLock(String lockKey, String lockValue, Duration leaseTime, Mono<Void> operation) {
        return acquireLock(lockKey, lockValue, leaseTime)
                .flatMap(acquired -> {
                    if (acquired) {
                        return operation
                                .doFinally(signalType ->
                                        releaseLock(lockKey, lockValue)
                                                .doOnError(error -> logger.error("Error releasing lock in finally: {}", error.getMessage()))
                                                .subscribe()
                                );
                    } else {
                        return Mono.error(new RuntimeException("Could not acquire lock: " + lockKey));
                    }
                });
    }
}