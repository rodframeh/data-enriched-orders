package com.apex.orderprocessingworker.infrastructure.messaging;

import com.apex.orderprocessingworker.application.service.OrderProcessingService;
import com.apex.orderprocessingworker.infrastructure.model.OrderMessage;
import com.fasterxml.jackson.databind.ObjectMapper;
import jakarta.validation.ConstraintViolation;
import jakarta.validation.Validator;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.kafka.support.Acknowledgment;
import reactor.core.publisher.Mono;

import java.util.Collections;
import java.util.List;
import java.util.Set;
import java.util.concurrent.CompletableFuture;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class OrderKafkaConsumerTest {

    @Mock
    private OrderProcessingService orderProcessingService;

    @Mock
    private Validator validator;

    @Mock
    private KafkaTemplate<String, String> kafkaTemplate;

    @Mock
    private Acknowledgment acknowledgment;

    private OrderKafkaConsumer orderKafkaConsumer;
    private ObjectMapper objectMapper;

    @BeforeEach
    void setUp() {
        objectMapper = new ObjectMapper();
        orderKafkaConsumer = new OrderKafkaConsumer(
                orderProcessingService,
                objectMapper,
                validator,
                kafkaTemplate
        );
    }

    @Test
    void handleOrderMessage_ShouldProcessValidMessage() throws Exception {
        // Given
        OrderMessage validOrderMessage = new OrderMessage(
                "order-123",
                "customer-456",
                List.of(new OrderMessage.ProductItem("product-789", 2))
        );

        String messageJson = objectMapper.writeValueAsString(validOrderMessage);

        when(validator.validate(any(OrderMessage.class)))
                .thenReturn(Collections.emptySet());
        when(orderProcessingService.processOrder(any(OrderMessage.class)))
                .thenReturn(Mono.empty());

        // When
        orderKafkaConsumer.handleOrderMessage(
                messageJson,
                "orders",
                0,
                1L,
                "order-123",
                acknowledgment
        );

        // Then
        // Wait a bit for async processing
        Thread.sleep(100);

        verify(validator).validate(any(OrderMessage.class));
        verify(orderProcessingService).processOrder(any(OrderMessage.class));
        verify(acknowledgment).acknowledge();
    }

    @Test
    void handleOrderMessage_ShouldHandleInvalidJson() throws Exception {
        // Given
        String invalidJson = "{ invalid json }";

        when(kafkaTemplate.send(anyString(), anyString(), anyString()))
                .thenReturn(CompletableFuture.completedFuture(null));

        // When
        orderKafkaConsumer.handleOrderMessage(
                invalidJson,
                "orders",
                0,
                1L,
                "unknown",
                acknowledgment
        );

        // Then
        // Wait a bit for processing
        Thread.sleep(100);

        verify(acknowledgment).acknowledge(); // Should acknowledge malformed messages
        verify(kafkaTemplate).send(eq("order-errors"), eq("unknown"), anyString());
        verify(orderProcessingService, never()).processOrder(any(OrderMessage.class));
    }

    @Test
    void handleOrderMessage_ShouldHandleValidationErrors() throws Exception {
        // Given
        OrderMessage invalidOrderMessage = new OrderMessage(
                "", // blank order ID
                "customer-456",
                List.of(new OrderMessage.ProductItem("product-789", 2))
        );

        String messageJson = objectMapper.writeValueAsString(invalidOrderMessage);

        // Mock validation failure
        ConstraintViolation<OrderMessage> violation = mock(ConstraintViolation.class);
        when(violation.getPropertyPath()).thenReturn(mock(jakarta.validation.Path.class));
        when(violation.getMessage()).thenReturn("Order ID cannot be blank");

        Set<ConstraintViolation<OrderMessage>> violations = Set.of(violation);
        when(validator.validate(any(OrderMessage.class)))
                .thenReturn(violations);

        when(kafkaTemplate.send(anyString(), anyString(), anyString()))
                .thenReturn(CompletableFuture.completedFuture(null));

        // When
        orderKafkaConsumer.handleOrderMessage(
                messageJson,
                "orders",
                0,
                1L,
                "order-123",
                acknowledgment
        );

        // Then
        Thread.sleep(100);

        verify(acknowledgment).acknowledge(); // Should acknowledge invalid messages
        verify(kafkaTemplate).send(eq("order-errors"), anyString(), anyString());
        verify(orderProcessingService, never()).processOrder(any(OrderMessage.class));
    }

    @Test
    void handleOrderMessage_ShouldHandleProcessingFailure() throws Exception {
        // Given
        OrderMessage validOrderMessage = new OrderMessage(
                "order-123",
                "customer-456",
                List.of(new OrderMessage.ProductItem("product-789", 2))
        );

        String messageJson = objectMapper.writeValueAsString(validOrderMessage);

        when(validator.validate(any(OrderMessage.class)))
                .thenReturn(Collections.emptySet());
        when(orderProcessingService.processOrder(any(OrderMessage.class)))
                .thenReturn(Mono.error(new RuntimeException("Processing failed")));

        when(kafkaTemplate.send(anyString(), anyString(), anyString()))
                .thenReturn(CompletableFuture.completedFuture(null));

        // When
        orderKafkaConsumer.handleOrderMessage(
                messageJson,
                "orders",
                0,
                1L,
                "order-123",
                acknowledgment
        );

        // Then
        Thread.sleep(100);

        verify(validator).validate(any(OrderMessage.class));
        verify(orderProcessingService).processOrder(any(OrderMessage.class));
        verify(kafkaTemplate).send(eq("order-errors"), eq("order-123"), anyString());
        verify(acknowledgment, never()).acknowledge(); // Should not acknowledge on processing failure
    }

    @Test
    void handleOrderMessage_ShouldSendErrorMessageToKafka() throws Exception {
        // Given
        String invalidJson = "{ invalid }";

        ArgumentCaptor<String> topicCaptor = ArgumentCaptor.forClass(String.class);
        ArgumentCaptor<String> keyCaptor = ArgumentCaptor.forClass(String.class);
        ArgumentCaptor<String> valueCaptor = ArgumentCaptor.forClass(String.class);

        when(kafkaTemplate.send(topicCaptor.capture(), keyCaptor.capture(), valueCaptor.capture()))
                .thenReturn(CompletableFuture.completedFuture(null));

        // When
        orderKafkaConsumer.handleOrderMessage(
                invalidJson,
                "orders",
                0,
                1L,
                "test-key",
                acknowledgment
        );

        // Then
        Thread.sleep(100);

        assertEquals("order-errors", topicCaptor.getValue());
        assertEquals("unknown", keyCaptor.getValue());

        // Verify error message content
        String errorMessageJson = valueCaptor.getValue();
        assertTrue(errorMessageJson.contains("\"orderId\":\"unknown\""));
        assertTrue(errorMessageJson.contains("\"originalMessage\":\"{ invalid }\""));
        assertTrue(errorMessageJson.contains("\"errorMessage\""));
    }

    @Test
    void handleOrderMessage_ShouldHandleKafkaErrorTopicFailure() throws Exception {
        // Given
        String invalidJson = "{ invalid }";

        CompletableFuture<org.springframework.kafka.support.SendResult<String, String>> failedFuture = new CompletableFuture<>();
        failedFuture.completeExceptionally(new RuntimeException("Kafka send failed"));

        when(kafkaTemplate.send(anyString(), anyString(), anyString()))
                .thenReturn(failedFuture);

        // When - Should not throw exception even if error topic send fails
        assertDoesNotThrow(() -> {
            orderKafkaConsumer.handleOrderMessage(
                    invalidJson,
                    "orders",
                    0,
                    1L,
                    "test-key",
                    acknowledgment
            );
        });

        // Then
        Thread.sleep(100);
        verify(acknowledgment).acknowledge();
    }

    @Test
    void handleOrderMessage_ShouldHandleSuccessfulProcessingWithComplexOrder() throws Exception {
        // Given
        OrderMessage complexOrder = new OrderMessage(
                "order-complex-123",
                "customer-premium-456",
                List.of(
                        new OrderMessage.ProductItem("product-laptop-789", 1),
                        new OrderMessage.ProductItem("product-mouse-001", 2),
                        new OrderMessage.ProductItem("product-keyboard-002", 1)
                )
        );

        String messageJson = objectMapper.writeValueAsString(complexOrder);

        when(validator.validate(any(OrderMessage.class)))
                .thenReturn(Collections.emptySet());
        when(orderProcessingService.processOrder(any(OrderMessage.class)))
                .thenReturn(Mono.empty());

        // When
        orderKafkaConsumer.handleOrderMessage(
                messageJson,
                "orders",
                0,
                1L,
                "order-complex-123",
                acknowledgment
        );

        // Then
        Thread.sleep(100);

        ArgumentCaptor<OrderMessage> orderCaptor = ArgumentCaptor.forClass(OrderMessage.class);
        verify(orderProcessingService).processOrder(orderCaptor.capture());

        OrderMessage capturedOrder = orderCaptor.getValue();
        assertEquals("order-complex-123", capturedOrder.orderId());
        assertEquals("customer-premium-456", capturedOrder.customerId());
        assertEquals(3, capturedOrder.products().size());

        verify(acknowledgment).acknowledge();
    }
}