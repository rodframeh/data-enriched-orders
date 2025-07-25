package com.apex.orderprocessingworker.infrastructure.service;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.web.reactive.function.client.WebClient;
import reactor.test.StepVerifier;

import java.time.Duration;

import static org.junit.jupiter.api.Assertions.*;

@ExtendWith(MockitoExtension.class)
class ExternalApiServiceTest {

    private ExternalApiService externalApiService;

    @BeforeEach
    void setUp() {
        // Crear el servicio con URLs de prueba
        WebClient.Builder webClientBuilder = WebClient.builder();

        externalApiService = new ExternalApiService(
                "http://localhost:3001",
                "http://localhost:3002",
                Duration.ofSeconds(5),
                webClientBuilder
        );
    }

    @Test
    void fallbackGetProduct_ShouldReturnErrorWithCustomMessage() {
        // Given
        String productId = "product-789";
        RuntimeException originalException = new RuntimeException("Service unavailable");

        // When & Then
        StepVerifier.create(externalApiService.fallbackGetProduct(productId, originalException))
                .expectErrorMatches(throwable -> {
                    boolean isRuntimeException = throwable instanceof RuntimeException;
                    boolean hasCorrectMessage = throwable.getMessage().contains("Product service unavailable for product: " + productId);
                    boolean hasCause = throwable.getCause() == originalException;

                    return isRuntimeException && hasCorrectMessage && hasCause;
                })
                .verify();
    }

    @Test
    void fallbackGetCustomer_ShouldReturnErrorWithCustomMessage() {
        // Given
        String customerId = "customer-456";
        RuntimeException originalException = new RuntimeException("Service unavailable");

        // When & Then
        StepVerifier.create(externalApiService.fallbackGetCustomer(customerId, originalException))
                .expectErrorMatches(throwable -> {
                    boolean isRuntimeException = throwable instanceof RuntimeException;
                    boolean hasCorrectMessage = throwable.getMessage().contains("Customer service unavailable for customer: " + customerId);
                    boolean hasCause = throwable.getCause() == originalException;

                    return isRuntimeException && hasCorrectMessage && hasCause;
                })
                .verify();
    }

    @Test
    void fallbackGetProduct_ShouldHandleDifferentExceptionTypes() {
        // Given
        String productId = "product-123";

        // Test with different exception types
        Exception networkException = new RuntimeException("Network timeout");
        Exception circuitBreakerException = new RuntimeException("Circuit breaker open");
        Exception unknownException = new RuntimeException("Unknown error");

        // When & Then - Network exception
        StepVerifier.create(externalApiService.fallbackGetProduct(productId, networkException))
                .expectErrorMatches(throwable ->
                        throwable.getMessage().contains("Product service unavailable for product: " + productId) &&
                                throwable.getCause() == networkException)
                .verify();

        // When & Then - Circuit breaker exception
        StepVerifier.create(externalApiService.fallbackGetProduct(productId, circuitBreakerException))
                .expectErrorMatches(throwable ->
                        throwable.getMessage().contains("Product service unavailable for product: " + productId) &&
                                throwable.getCause() == circuitBreakerException)
                .verify();

        // When & Then - Unknown exception
        StepVerifier.create(externalApiService.fallbackGetProduct(productId, unknownException))
                .expectErrorMatches(throwable ->
                        throwable.getMessage().contains("Product service unavailable for product: " + productId) &&
                                throwable.getCause() == unknownException)
                .verify();
    }

    @Test
    void fallbackGetCustomer_ShouldHandleDifferentExceptionTypes() {
        // Given
        String customerId = "customer-789";

        Exception timeoutException = new RuntimeException("Request timeout");
        Exception serviceException = new RuntimeException("Service down");

        // When & Then - Timeout exception
        StepVerifier.create(externalApiService.fallbackGetCustomer(customerId, timeoutException))
                .expectErrorMatches(throwable ->
                        throwable.getMessage().contains("Customer service unavailable for customer: " + customerId) &&
                                throwable.getCause() == timeoutException)
                .verify();

        // When & Then - Service exception
        StepVerifier.create(externalApiService.fallbackGetCustomer(customerId, serviceException))
                .expectErrorMatches(throwable ->
                        throwable.getMessage().contains("Customer service unavailable for customer: " + customerId) &&
                                throwable.getCause() == serviceException)
                .verify();
    }

    @Test
    void fallbackGetProduct_ShouldHandleNullProductId() {
        // Given
        String nullProductId = null;
        Exception exception = new RuntimeException("Service error");

        // When & Then
        StepVerifier.create(externalApiService.fallbackGetProduct(nullProductId, exception))
                .expectErrorMatches(throwable ->
                        throwable.getMessage().contains("Product service unavailable for product: null"))
                .verify();
    }

    @Test
    void fallbackGetCustomer_ShouldHandleEmptyCustomerId() {
        // Given
        String emptyCustomerId = "";
        Exception exception = new RuntimeException("Service error");

        // When & Then
        StepVerifier.create(externalApiService.fallbackGetCustomer(emptyCustomerId, exception))
                .expectErrorMatches(throwable ->
                        throwable.getMessage().contains("Customer service unavailable for customer: "))
                .verify();
    }

    @Test
    void fallbackGetProduct_ShouldPreserveCauseChain() {
        // Given
        String productId = "product-chain";
        RuntimeException rootCause = new RuntimeException("Root cause");
        RuntimeException intermediateCause = new RuntimeException("Intermediate cause", rootCause);

        // When & Then
        StepVerifier.create(externalApiService.fallbackGetProduct(productId, intermediateCause))
                .expectErrorMatches(throwable -> {
                    // Verify the immediate cause
                    boolean hasImmediateCause = throwable.getCause() == intermediateCause;

                    // Verify the root cause is preserved
                    boolean hasRootCause = throwable.getCause().getCause() == rootCause;

                    return hasImmediateCause && hasRootCause;
                })
                .verify();
    }

    @Test
    void fallbackGetCustomer_ShouldPreserveCauseChain() {
        // Given
        String customerId = "customer-chain";
        RuntimeException rootCause = new RuntimeException("Database connection failed");
        RuntimeException intermediateCause = new RuntimeException("Service layer error", rootCause);

        // When & Then
        StepVerifier.create(externalApiService.fallbackGetCustomer(customerId, intermediateCause))
                .expectErrorMatches(throwable -> {
                    boolean hasImmediateCause = throwable.getCause() == intermediateCause;
                    boolean hasRootCause = throwable.getCause().getCause() == rootCause;

                    return hasImmediateCause && hasRootCause;
                })
                .verify();
    }

    @Test
    void fallbackMethods_ShouldHaveConsistentErrorMessageFormat() {
        // Given
        String productId = "test-product";
        String customerId = "test-customer";
        Exception testException = new RuntimeException("Test error");

        // When & Then - Verify product fallback message format
        StepVerifier.create(externalApiService.fallbackGetProduct(productId, testException))
                .expectErrorMatches(throwable -> {
                    String message = throwable.getMessage();
                    return message.startsWith("Product service unavailable for product: ") &&
                            message.endsWith(productId);
                })
                .verify();

        // When & Then - Verify customer fallback message format
        StepVerifier.create(externalApiService.fallbackGetCustomer(customerId, testException))
                .expectErrorMatches(throwable -> {
                    String message = throwable.getMessage();
                    return message.startsWith("Customer service unavailable for customer: ") &&
                            message.endsWith(customerId);
                })
                .verify();
    }

    @Test
    void service_ShouldBeCreatedWithCorrectConfiguration() {
        // Given & When
        WebClient.Builder builder = WebClient.builder();
        ExternalApiService service = new ExternalApiService(
                "http://test-product:8001",
                "http://test-customer:8002",
                Duration.ofSeconds(10),
                builder
        );

        // Then
        assertNotNull(service);
        // Service should be created without throwing exceptions
    }

    @Test
    void service_ShouldHandleNullUrls() {
        // Given & When & Then
        WebClient.Builder builder = WebClient.builder();

        // Should not throw exception during creation, but might fail during actual calls
        assertDoesNotThrow(() -> {
            new ExternalApiService(null, null, Duration.ofSeconds(5), builder);
        });
    }

    @Test
    void fallbackGetProduct_ShouldHandleSpecialCharactersInProductId() {
        // Given
        String specialProductId = "product-123-@#$%^&*()";
        Exception exception = new RuntimeException("Error");

        // When & Then
        StepVerifier.create(externalApiService.fallbackGetProduct(specialProductId, exception))
                .expectErrorMatches(throwable ->
                        throwable.getMessage().contains("Product service unavailable for product: " + specialProductId))
                .verify();
    }

    @Test
    void fallbackGetCustomer_ShouldHandleSpecialCharactersInCustomerId() {
        // Given
        String specialCustomerId = "customer-456-!@#$%";
        Exception exception = new RuntimeException("Error");

        // When & Then
        StepVerifier.create(externalApiService.fallbackGetCustomer(specialCustomerId, exception))
                .expectErrorMatches(throwable ->
                        throwable.getMessage().contains("Customer service unavailable for customer: " + specialCustomerId))
                .verify();
    }
}