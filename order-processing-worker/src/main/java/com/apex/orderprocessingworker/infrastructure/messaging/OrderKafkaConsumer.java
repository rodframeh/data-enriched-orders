package com.apex.orderprocessingworker.infrastructure.messaging;

import com.apex.orderprocessingworker.application.service.OrderProcessingService;
import com.apex.orderprocessingworker.infrastructure.model.OrderMessage;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import jakarta.validation.ConstraintViolation;
import jakarta.validation.Validator;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.kafka.support.Acknowledgment;
import org.springframework.kafka.support.KafkaHeaders;
import org.springframework.messaging.handler.annotation.Header;
import org.springframework.messaging.handler.annotation.Payload;
import org.springframework.stereotype.Component;

import java.util.Set;

@Component
public class OrderKafkaConsumer {

    private static final Logger logger = LoggerFactory.getLogger(OrderKafkaConsumer.class);

    private final OrderProcessingService orderProcessingService;
    private final ObjectMapper objectMapper;
    private final Validator validator;
    private final KafkaTemplate<String, String> kafkaTemplate;

    public OrderKafkaConsumer(
            OrderProcessingService orderProcessingService,
            ObjectMapper objectMapper,
            Validator validator,
            KafkaTemplate<String, String> kafkaTemplate) {
        this.orderProcessingService = orderProcessingService;
        this.objectMapper = objectMapper;
        this.validator = validator;
        this.kafkaTemplate = kafkaTemplate;
    }

    @KafkaListener(
            topics = "${app.kafka.topics.orders}",
            groupId = "${spring.kafka.consumer.group-id}",
            containerFactory = "kafkaListenerContainerFactory"
    )
    public void handleOrderMessage(
            @Payload String message,
            @Header(KafkaHeaders.RECEIVED_TOPIC) String topic,
            @Header(KafkaHeaders.RECEIVED_PARTITION) int partition,
            @Header(KafkaHeaders.OFFSET) long offset,
            @Header(KafkaHeaders.RECEIVED_KEY) String key,
            Acknowledgment acknowledgment) {

        logger.info("Received message from topic: {}, partition: {}, offset: {}, key: {}",
                topic, partition, offset, key);
        logger.debug("Message content: {}", message);

        try {
            // Parse and validate message
            OrderMessage orderMessage = parseAndValidateMessage(message);

            // Process order asynchronously
            orderProcessingService.processOrder(orderMessage)
                    .doOnSuccess(v -> {
                        logger.info("Order processed successfully: {}", orderMessage.orderId());
                        acknowledgment.acknowledge();
                    })
                    .doOnError(error -> {
                        logger.error("Error processing order {}: {}", orderMessage.orderId(), error.getMessage());
                        // Don't acknowledge - let Kafka retry or send to DLQ
                        handleProcessingError(orderMessage, error, message);
                    })
                    .subscribe();

        } catch (Exception e) {
            logger.error("Error parsing message from topic {}: {}", topic, e.getMessage());
            handleParsingError(message, e);
            acknowledgment.acknowledge(); // Acknowledge to avoid infinite retries for malformed messages
        }
    }

    private OrderMessage parseAndValidateMessage(String message) throws JsonProcessingException {
        // Parse JSON
        OrderMessage orderMessage = objectMapper.readValue(message, OrderMessage.class);

        // Validate using Bean Validation
        Set<ConstraintViolation<OrderMessage>> violations = validator.validate(orderMessage);
        if (!violations.isEmpty()) {
            StringBuilder sb = new StringBuilder("Validation errors: ");
            violations.forEach(violation ->
                    sb.append(violation.getPropertyPath()).append(" ").append(violation.getMessage()).append("; ")
            );
            throw new IllegalArgumentException(sb.toString());
        }

        logger.debug("Message parsed and validated successfully for order: {}", orderMessage.orderId());
        return orderMessage;
    }

    private void handleProcessingError(OrderMessage orderMessage, Throwable error, String originalMessage) {
        logger.error("Processing failed for order: {}", orderMessage.orderId(), error);

        // Could send to dead letter queue or error topic
        sendToErrorTopic(orderMessage.orderId(), originalMessage, error.getMessage());
    }

    private void handleParsingError(String message, Throwable error) {
        logger.error("Message parsing failed", error);

        // Send malformed message to error topic
        sendToErrorTopic("unknown", message, "Parsing error: " + error.getMessage());
    }

    private void sendToErrorTopic(String orderId, String originalMessage, String errorMessage) {
        try {
            ErrorMessage errorMsg = new ErrorMessage(orderId, originalMessage, errorMessage, System.currentTimeMillis());
            String errorJson = objectMapper.writeValueAsString(errorMsg);

            kafkaTemplate.send("order-errors", orderId, errorJson)
                    .whenComplete((result, throwable) -> {
                        if (throwable != null) {
                            logger.error("Failed to send error message to topic for order: {}", orderId, throwable);
                        } else {
                            logger.info("Error message sent to topic for order: {}", orderId);
                        }
                    });
        } catch (Exception e) {
            logger.error("Failed to serialize error message for order: {}", orderId, e);
        }
    }

    private record ErrorMessage(
            String orderId,
            String originalMessage,
            String errorMessage,
            long timestamp
    ) {}
}