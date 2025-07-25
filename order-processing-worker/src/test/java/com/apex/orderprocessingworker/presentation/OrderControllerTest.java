package com.apex.orderprocessingworker.presentation;

import com.apex.orderprocessingworker.application.service.OrderProcessingService;
import com.apex.orderprocessingworker.domain.entity.Order;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.ResponseEntity;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;
import reactor.test.StepVerifier;

import java.math.BigDecimal;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class OrderControllerTest {

    @Mock
    private OrderProcessingService orderProcessingService;

    @InjectMocks
    private OrderController orderController;

    private Order testOrder;

    @BeforeEach
    void setUp() {
        testOrder = Order.create(
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
    void getOrder_ShouldReturnOrderWhenFound() {
        // Given
        when(orderProcessingService.findByOrderId("order-123"))
                .thenReturn(Mono.just(testOrder));

        // When & Then
        StepVerifier.create(orderController.getOrder("order-123"))
                .expectNextMatches(response -> {
                    assertTrue(response.getStatusCode().is2xxSuccessful());
                    assertEquals(testOrder, response.getBody());
                    return true;
                })
                .expectComplete()
                .verify();

        verify(orderProcessingService).findByOrderId("order-123");
    }

    @Test
    void getOrder_ShouldReturnNotFoundWhenOrderDoesNotExist() {
        // Given
        when(orderProcessingService.findByOrderId("nonexistent"))
                .thenReturn(Mono.empty());

        // When & Then
        StepVerifier.create(orderController.getOrder("nonexistent"))
                .expectNextMatches(response -> {
                    assertEquals(404, response.getStatusCodeValue());
                    assertNull(response.getBody());
                    return true;
                })
                .expectComplete()
                .verify();

        verify(orderProcessingService).findByOrderId("nonexistent");
    }

    @Test
    void getOrder_ShouldHandleServiceError() {
        // Given
        when(orderProcessingService.findByOrderId("error-order"))
                .thenReturn(Mono.error(new RuntimeException("Database error")));

        // When & Then
        StepVerifier.create(orderController.getOrder("error-order"))
                .expectError(RuntimeException.class)
                .verify();

        verify(orderProcessingService).findByOrderId("error-order");
    }

    @Test
    void getOrdersByCustomer_ShouldReturnCustomerOrders() {
        // Given
        Order secondOrder = Order.create(
                "order-456",
                "customer-456",
                List.of(new Order.OrderProduct(
                        "product-001",
                        "Mouse",
                        "Wireless mouse",
                        new BigDecimal("29.99"),
                        1
                ))
        );

        when(orderProcessingService.findByCustomerId("customer-456"))
                .thenReturn(Flux.just(testOrder, secondOrder));

        // When & Then
        StepVerifier.create(orderController.getOrdersByCustomer("customer-456"))
                .expectNext(testOrder)
                .expectNext(secondOrder)
                .expectComplete()
                .verify();

        verify(orderProcessingService).findByCustomerId("customer-456");
    }

    @Test
    void getOrdersByCustomer_ShouldReturnEmptyWhenNoOrders() {
        // Given
        when(orderProcessingService.findByCustomerId("customer-without-orders"))
                .thenReturn(Flux.empty());

        // When & Then
        StepVerifier.create(orderController.getOrdersByCustomer("customer-without-orders"))
                .expectComplete()
                .verify();

        verify(orderProcessingService).findByCustomerId("customer-without-orders");
    }

    @Test
    void getOrdersByCustomer_ShouldHandleServiceError() {
        // Given
        when(orderProcessingService.findByCustomerId("error-customer"))
                .thenReturn(Flux.error(new RuntimeException("Database connection failed")));

        // When & Then
        StepVerifier.create(orderController.getOrdersByCustomer("error-customer"))
                .expectError(RuntimeException.class)
                .verify();

        verify(orderProcessingService).findByCustomerId("error-customer");
    }

    @Test
    void health_ShouldReturnHealthyStatus() {
        // When & Then
        StepVerifier.create(orderController.health())
                .expectNextMatches(response -> {
                    assertTrue(response.getStatusCode().is2xxSuccessful());
                    assertEquals("Order Processing Worker is running", response.getBody());
                    return true;
                })
                .expectComplete()
                .verify();

        // Health endpoint should not call any service
        verifyNoInteractions(orderProcessingService);
    }

    @Test
    void getOrder_ShouldHandleNullOrderId() {
        // Given
        when(orderProcessingService.findByOrderId(null))
                .thenReturn(Mono.empty());

        // When & Then
        StepVerifier.create(orderController.getOrder(null))
                .expectNextMatches(response -> response.getStatusCode().value() == 404)
                .expectComplete()
                .verify();
        verify(orderProcessingService).findByOrderId(null);
    }
}