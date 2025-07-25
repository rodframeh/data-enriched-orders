package com.apex.orderprocessingworker.application.service;

import com.apex.orderprocessingworker.domain.entity.Order;
import com.apex.orderprocessingworker.infrastructure.model.ExternalApiModels;
import com.apex.orderprocessingworker.infrastructure.model.OrderMessage;
import com.apex.orderprocessingworker.infrastructure.repository.OrderRepository;
import com.apex.orderprocessingworker.infrastructure.service.ExternalApiService;
import com.apex.orderprocessingworker.infrastructure.service.RedisLockService;
import com.apex.orderprocessingworker.infrastructure.service.RetryService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;
import reactor.test.StepVerifier;

import java.math.BigDecimal;
import java.time.Duration;
import java.util.List;

import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class OrderProcessingServiceTest {

    @Mock
    private OrderRepository orderRepository;

    @Mock
    private ExternalApiService externalApiService;

    @Mock
    private RedisLockService lockService;

    @Mock
    private RetryService retryService;

    @InjectMocks
    private OrderProcessingService orderProcessingService;

    private OrderMessage validOrderMessage;
    private ExternalApiModels.CustomerResponse validCustomer;
    private ExternalApiModels.ProductResponse validProduct;
    private Order expectedOrder;

    @BeforeEach
    void setUp() {
        // Setup test data
        validOrderMessage = new OrderMessage(
                "order-123",
                "customer-456",
                List.of(new OrderMessage.ProductItem("product-789", 2))
        );

        validCustomer = new ExternalApiModels.CustomerResponse(
                "customer-456",
                "John Doe",
                "john@example.com",
                "+1234567890",
                true,
                ExternalApiModels.CustomerStatus.ACTIVE
        );

        validProduct = new ExternalApiModels.ProductResponse(
                "product-789",
                "Laptop",
                "High-performance laptop",
                new BigDecimal("999.00"),
                "Electronics",
                true
        );

        expectedOrder = Order.create(
                "order-123",
                "customer-456",
                List.of(new Order.OrderProduct(
                        "product-789",
                        "Laptop",
                        "High-performance laptop",
                        new BigDecimal("999.00"),
                        2
                ))
        );
    }

    @Test
    void processOrder_ShouldSuccessfullyProcessValidOrder() {
        // Given
        when(lockService.withLock(anyString(), anyString(), any(Mono.class)))
                .thenAnswer(invocation -> invocation.getArgument(2));
        when(externalApiService.getCustomer("customer-456"))
                .thenReturn(Mono.just(validCustomer));
        when(externalApiService.getProduct("product-789"))
                .thenReturn(Mono.just(validProduct));
        when(orderRepository.save(any(Order.class)))
                .thenReturn(Mono.just(expectedOrder));

        // When & Then
        StepVerifier.create(orderProcessingService.processOrder(validOrderMessage))
                .expectComplete()
                .verify(Duration.ofSeconds(5));

        verify(externalApiService).getCustomer("customer-456");
        verify(externalApiService).getProduct("product-789");
        verify(orderRepository).save(any(Order.class));
    }

    @Test
    void processOrder_ShouldFailWhenCustomerIsInactive() {
        // Given
        ExternalApiModels.CustomerResponse inactiveCustomer = new ExternalApiModels.CustomerResponse(
                "customer-456",
                "John Doe",
                "john@example.com",
                "+1234567890",
                false, // inactive
                ExternalApiModels.CustomerStatus.INACTIVE
        );

        when(lockService.withLock(anyString(), anyString(), any(Mono.class)))
                .thenAnswer(invocation -> invocation.getArgument(2));
        when(externalApiService.getCustomer("customer-456"))
                .thenReturn(Mono.just(inactiveCustomer));
        when(retryService.shouldRetry(anyString()))
                .thenReturn(Mono.just(false));
        when(retryService.storeFailedMessage(anyString(), anyString(), anyString()))
                .thenReturn(Mono.empty());

        // When & Then
        StepVerifier.create(orderProcessingService.processOrder(validOrderMessage))
                .expectComplete()
                .verify(Duration.ofSeconds(5));

        verify(externalApiService).getCustomer("customer-456");
        verify(retryService).storeFailedMessage(eq("order-123"), anyString(), contains("Customer is not active"));
        verify(orderRepository, never()).save(any(Order.class));
    }

    @Test
    void processOrder_ShouldFailWhenCustomerIsBlocked() {
        // Given
        ExternalApiModels.CustomerResponse blockedCustomer = new ExternalApiModels.CustomerResponse(
                "customer-456",
                "John Doe",
                "john@example.com",
                "+1234567890",
                true,
                ExternalApiModels.CustomerStatus.BLOCKED
        );

        when(lockService.withLock(anyString(), anyString(), any(Mono.class)))
                .thenAnswer(invocation -> invocation.getArgument(2));
        when(externalApiService.getCustomer("customer-456"))
                .thenReturn(Mono.just(blockedCustomer));
        when(retryService.shouldRetry(anyString()))
                .thenReturn(Mono.just(false));
        when(retryService.storeFailedMessage(anyString(), anyString(), anyString()))
                .thenReturn(Mono.empty());

        // When & Then
        StepVerifier.create(orderProcessingService.processOrder(validOrderMessage))
                .expectComplete()
                .verify(Duration.ofSeconds(5));

        verify(externalApiService).getCustomer("customer-456");
        verify(retryService).storeFailedMessage(eq("order-123"), anyString(), contains("Customer is blocked"));
    }

    @Test
    void processOrder_ShouldFailWhenProductIsInactive() {
        // Given
        ExternalApiModels.ProductResponse inactiveProduct = new ExternalApiModels.ProductResponse(
                "product-789",
                "Laptop",
                "High-performance laptop",
                new BigDecimal("999.00"),
                "Electronics",
                false // inactive
        );

        when(lockService.withLock(anyString(), anyString(), any(Mono.class)))
                .thenAnswer(invocation -> invocation.getArgument(2));
        when(externalApiService.getCustomer("customer-456"))
                .thenReturn(Mono.just(validCustomer));
        when(externalApiService.getProduct("product-789"))
                .thenReturn(Mono.just(inactiveProduct));
        when(retryService.shouldRetry(anyString()))
                .thenReturn(Mono.just(false));
        when(retryService.storeFailedMessage(anyString(), anyString(), anyString()))
                .thenReturn(Mono.empty());

        // When & Then
        StepVerifier.create(orderProcessingService.processOrder(validOrderMessage))
                .expectComplete()
                .verify(Duration.ofSeconds(5));

        verify(externalApiService).getCustomer("customer-456");
        verify(externalApiService).getProduct("product-789");
        verify(retryService).storeFailedMessage(eq("order-123"), anyString(), contains("Product is not active"));
    }

    @Test
    void processOrder_ShouldFailWhenProductHasInvalidPrice() {
        // Given
        ExternalApiModels.ProductResponse invalidPriceProduct = new ExternalApiModels.ProductResponse(
                "product-789",
                "Laptop",
                "High-performance laptop",
                BigDecimal.ZERO, // invalid price
                "Electronics",
                true
        );

        when(lockService.withLock(anyString(), anyString(), any(Mono.class)))
                .thenAnswer(invocation -> invocation.getArgument(2));
        when(externalApiService.getCustomer("customer-456"))
                .thenReturn(Mono.just(validCustomer));
        when(externalApiService.getProduct("product-789"))
                .thenReturn(Mono.just(invalidPriceProduct));
        when(retryService.shouldRetry(anyString()))
                .thenReturn(Mono.just(false));
        when(retryService.storeFailedMessage(anyString(), anyString(), anyString()))
                .thenReturn(Mono.empty());

        // When & Then
        StepVerifier.create(orderProcessingService.processOrder(validOrderMessage))
                .expectComplete()
                .verify(Duration.ofSeconds(5));

        verify(retryService).storeFailedMessage(eq("order-123"), anyString(), contains("Invalid product price"));
    }

    @Test
    void processOrder_ShouldRetryWhenApiCallFails() {
        // Given
        when(lockService.withLock(anyString(), anyString(), any(Mono.class)))
                .thenAnswer(invocation -> invocation.getArgument(2));
        when(externalApiService.getCustomer("customer-456"))
                .thenReturn(Mono.error(new RuntimeException("Service unavailable")));
        when(retryService.shouldRetry("order-123"))
                .thenReturn(Mono.just(true));
        when(retryService.incrementRetryCount("order-123"))
                .thenReturn(Mono.empty());

        // When & Then
        StepVerifier.create(orderProcessingService.processOrder(validOrderMessage))
                .expectComplete()
                .verify(Duration.ofSeconds(5));

        verify(retryService).incrementRetryCount("order-123");
        verify(retryService, never()).storeFailedMessage(anyString(), anyString(), anyString());
    }

    @Test
    void processOrder_ShouldHandleMultipleProducts() {
        // Given
        OrderMessage multiProductOrder = new OrderMessage(
                "order-456",
                "customer-456",
                List.of(
                        new OrderMessage.ProductItem("product-789", 1),
                        new OrderMessage.ProductItem("product-001", 3)
                )
        );

        ExternalApiModels.ProductResponse secondProduct = new ExternalApiModels.ProductResponse(
                "product-001",
                "Mouse",
                "Wireless mouse",
                new BigDecimal("29.99"),
                "Electronics",
                true
        );

        when(lockService.withLock(anyString(), anyString(), any(Mono.class)))
                .thenAnswer(invocation -> invocation.getArgument(2));
        when(externalApiService.getCustomer("customer-456"))
                .thenReturn(Mono.just(validCustomer));
        when(externalApiService.getProduct("product-789"))
                .thenReturn(Mono.just(validProduct));
        when(externalApiService.getProduct("product-001"))
                .thenReturn(Mono.just(secondProduct));
        when(orderRepository.save(any(Order.class)))
                .thenReturn(Mono.just(expectedOrder));

        // When & Then
        StepVerifier.create(orderProcessingService.processOrder(multiProductOrder))
                .expectComplete()
                .verify(Duration.ofSeconds(5));

        verify(externalApiService).getProduct("product-789");
        verify(externalApiService).getProduct("product-001");
        verify(orderRepository).save(any(Order.class));
    }

    @Test
    void findByOrderId_ShouldReturnOrder() {
        // Given
        when(orderRepository.findByOrderId("order-123"))
                .thenReturn(Mono.just(expectedOrder));

        // When & Then
        StepVerifier.create(orderProcessingService.findByOrderId("order-123"))
                .expectNext(expectedOrder)
                .expectComplete()
                .verify();

        verify(orderRepository).findByOrderId("order-123");
    }

    @Test
    void findByOrderId_ShouldReturnEmptyWhenNotFound() {
        // Given
        when(orderRepository.findByOrderId("nonexistent"))
                .thenReturn(Mono.empty());

        // When & Then
        StepVerifier.create(orderProcessingService.findByOrderId("nonexistent"))
                .expectComplete()
                .verify();

        verify(orderRepository).findByOrderId("nonexistent");
    }

    @Test
    void findByCustomerId_ShouldReturnCustomerOrders() {
        // Given
        List<Order> customerOrders = List.of(expectedOrder, expectedOrder);
        when(orderRepository.findByCustomerId("customer-456"))
                .thenReturn(Flux.fromIterable(customerOrders));

        // When & Then
        StepVerifier.create(orderProcessingService.findByCustomerId("customer-456"))
                .expectNext(expectedOrder)
                .expectNext(expectedOrder)
                .expectComplete()
                .verify();

        verify(orderRepository).findByCustomerId("customer-456");
    }

    @Test
    void findByCustomerId_ShouldReturnEmptyWhenNoOrders() {
        // Given
        when(orderRepository.findByCustomerId("customer-789"))
                .thenReturn(Flux.empty());

        // When & Then
        StepVerifier.create(orderProcessingService.findByCustomerId("customer-789"))
                .expectComplete()
                .verify();

        verify(orderRepository).findByCustomerId("customer-789");
    }
}