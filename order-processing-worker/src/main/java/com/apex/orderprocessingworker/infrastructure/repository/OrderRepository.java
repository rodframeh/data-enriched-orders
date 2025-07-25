package com.apex.orderprocessingworker.infrastructure.repository;

import com.apex.orderprocessingworker.domain.entity.Order;
import org.springframework.data.mongodb.repository.ReactiveMongoRepository;
import org.springframework.data.mongodb.repository.Query;
import org.springframework.stereotype.Repository;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;

import java.time.LocalDateTime;

@Repository
public interface OrderRepository extends ReactiveMongoRepository<Order, String> {

    Mono<Order> findByOrderId(String orderId);

    Flux<Order> findByCustomerId(String customerId);

    @Query(value = "{ 'customerId': ?0 }", count = true)
    Mono<Long> countByCustomerId(String customerId);
}