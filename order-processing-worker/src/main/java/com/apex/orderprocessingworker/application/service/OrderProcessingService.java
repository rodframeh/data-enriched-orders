package com.apex.orderprocessingworker.application.service;

import com.apex.orderprocessingworker.domain.entity.Order;
import com.apex.orderprocessingworker.infrastructure.model.ExternalApiModels;
import com.apex.orderprocessingworker.infrastructure.model.OrderMessage;
import com.apex.orderprocessingworker.infrastructure.repository.OrderRepository;
import com.apex.orderprocessingworker.infrastructure.service.ExternalApiService;
import com.apex.orderprocessingworker.infrastructure.service.RedisLockService;
import com.apex.orderprocessingworker.infrastructure.service.RetryService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;

import java.util.List;
import java.util.UUID;

@Service
public class OrderProcessingService {

    private static final Logger logger = LoggerFactory.getLogger(OrderProcessingService.class);
    private static final String LOCK_PREFIX = "order_lock:";

    private final OrderRepository orderRepository;
    private final ExternalApiService externalApiService;
    private final RedisLockService lockService;
    private final RetryService retryService;

    public OrderProcessingService(
            OrderRepository orderRepository,
            ExternalApiService externalApiService,
            RedisLockService lockService,
            RetryService retryService) {
        this.orderRepository = orderRepository;
        this.externalApiService = externalApiService;
        this.lockService = lockService;
        this.retryService = retryService;
    }

    public Mono<Void> processOrder(OrderMessage orderMessage) {
        String lockKey = LOCK_PREFIX + orderMessage.orderId();
        String lockValue = UUID.randomUUID().toString();

        logger.info("Starting order processing for orderId: {}", orderMessage.orderId());

        return lockService.withLock(lockKey, lockValue,
                validateAndEnrichOrder(orderMessage)
                        .flatMap(this::saveOrder)
                        .doOnSuccess(savedOrder -> logger.info("Successfully processed order: {}", savedOrder.orderId()))
                        .doOnError(error -> logger.error("Error processing order {}: {}", orderMessage.orderId(), error.getMessage()))
                        .then()
        ).onErrorResume(error -> {
            logger.error("Failed to process order {}: {}", orderMessage.orderId(), error.getMessage());
            return handleProcessingError(orderMessage, error);
        });
    }

    private Mono<Order> validateAndEnrichOrder(OrderMessage orderMessage) {
        logger.debug("Validating and enriching order: {}", orderMessage.orderId());

        // Validate customer
        Mono<ExternalApiModels.CustomerResponse> customerMono =
                externalApiService.getCustomer(orderMessage.customerId())
                        .doOnNext(customer -> validateCustomer(customer));

        // Validate and enrich products
        Mono<List<Order.OrderProduct>> productsMono =
                Flux.fromIterable(orderMessage.products())
                        .flatMap(productItem ->
                                externalApiService.getProduct(productItem.productId())
                                        .doOnNext(product -> validateProduct(product))
                                        .map(product -> new Order.OrderProduct(
                                                product.id(),
                                                product.name(),
                                                product.description(),
                                                product.price(),
                                                productItem.quantity()
                                        ))
                        )
                        .collectList();

        return Mono.zip(customerMono, productsMono)
                .map(tuple -> Order.create(
                        orderMessage.orderId(),
                        orderMessage.customerId(),
                        tuple.getT2()
                ))
                .doOnSuccess(order -> logger.debug("Order validation and enrichment completed for: {}", orderMessage.orderId()));
    }

    private void validateCustomer(ExternalApiModels.CustomerResponse customer) {
        if (!customer.active()) {
            throw new IllegalArgumentException("Customer is not active: " + customer.id());
        }

        if (customer.status() == ExternalApiModels.CustomerStatus.BLOCKED) {
            throw new IllegalArgumentException("Customer is blocked: " + customer.id());
        }

        logger.debug("Customer validation passed for: {}", customer.id());
    }

    private void validateProduct(ExternalApiModels.ProductResponse product) {
        if (!product.active()) {
            throw new IllegalArgumentException("Product is not active: " + product.id());
        }

        if (product.price() == null || product.price().intValue() <= 0) {
            throw new IllegalArgumentException("Invalid product price: " + product.id());
        }

        logger.debug("Product validation passed for: {}", product.id());
    }

    private Mono<Order> saveOrder(Order order) {
        logger.debug("Saving order to database: {}", order.orderId());

        return orderRepository.save(order)
                .doOnSuccess(savedOrder -> logger.debug("Order saved successfully: {}", savedOrder.orderId()))
                .doOnError(error -> logger.error("Error saving order {}: {}", order.orderId(), error.getMessage()));
    }

    private Mono<Void> handleProcessingError(OrderMessage orderMessage, Throwable error) {
        String errorMessage = error.getMessage() != null ? error.getMessage() : error.getClass().getSimpleName();

        return retryService.shouldRetry(orderMessage.orderId())
                .flatMap(shouldRetry -> {
                    if (shouldRetry) {
                        logger.info("Incrementing retry count for order: {}", orderMessage.orderId());
                        return retryService.incrementRetryCount(orderMessage.orderId());
                    } else {
                        logger.info("Storing failed message for order: {}", orderMessage.orderId());
                        return retryService.storeFailedMessage(
                                orderMessage.orderId(),
                                orderMessage.toString(),
                                errorMessage
                        );
                    }
                })
                .doOnSuccess(v -> logger.debug("Error handling completed for order: {}", orderMessage.orderId()))
                .doOnError(handlingError -> logger.error("Error handling failed for order {}: {}",
                        orderMessage.orderId(), handlingError.getMessage()));
    }

    public Mono<Order> findByOrderId(String orderId) {
        return orderRepository.findByOrderId(orderId);
    }

    public Flux<Order> findByCustomerId(String customerId) {
        return orderRepository.findByCustomerId(customerId);
    }
}