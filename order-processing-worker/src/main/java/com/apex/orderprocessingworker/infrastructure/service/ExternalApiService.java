package com.apex.orderprocessingworker.infrastructure.service;

import com.apex.orderprocessingworker.infrastructure.model.ExternalApiModels;
import io.github.resilience4j.circuitbreaker.annotation.CircuitBreaker;
import io.github.resilience4j.retry.annotation.Retry;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.web.reactive.function.client.WebClient;
import reactor.core.publisher.Mono;

import java.time.Duration;

@Service
public class ExternalApiService {

    private static final Logger logger = LoggerFactory.getLogger(ExternalApiService.class);

    private final WebClient productServiceClient;
    private final WebClient customerServiceClient;

    public ExternalApiService(
            @Value("${app.external-apis.product-service.base-url}") String productServiceUrl,
            @Value("${app.external-apis.customer-service.base-url}") String customerServiceUrl,
            @Value("${app.external-apis.product-service.timeout}") Duration timeout,
            WebClient.Builder webClientBuilder) {

        this.productServiceClient = webClientBuilder
                .baseUrl(productServiceUrl)
                .codecs(configurer -> configurer.defaultCodecs().maxInMemorySize(1024 * 1024))
                .build();

        this.customerServiceClient = webClientBuilder
                .baseUrl(customerServiceUrl)
                .codecs(configurer -> configurer.defaultCodecs().maxInMemorySize(1024 * 1024))
                .build();
    }

    @Retry(name = "external-api", fallbackMethod = "fallbackGetProduct")
    @CircuitBreaker(name = "external-api", fallbackMethod = "fallbackGetProduct")
    public Mono<ExternalApiModels.ProductResponse> getProduct(String productId) {
        logger.debug("Fetching product details for productId: {}", productId);

        return productServiceClient
                .get()
                .uri("/api/products/{id}", productId)
                .retrieve()
                .bodyToMono(ExternalApiModels.ProductResponse.class)
                .doOnSuccess(product -> logger.debug("Successfully fetched product: {}", productId))
                .doOnError(error -> logger.error("Error fetching product {}: {}", productId, error.getMessage()))
                .timeout(Duration.ofSeconds(5));
    }

    @Retry(name = "external-api", fallbackMethod = "fallbackGetCustomer")
    @CircuitBreaker(name = "external-api", fallbackMethod = "fallbackGetCustomer")
    public Mono<ExternalApiModels.CustomerResponse> getCustomer(String customerId) {
        logger.debug("Fetching customer details for customerId: {}", customerId);

        return customerServiceClient
                .get()
                .uri("/api/customers/{id}", customerId)
                .retrieve()
                .bodyToMono(ExternalApiModels.CustomerResponse.class)
                .doOnSuccess(customer -> logger.debug("Successfully fetched customer: {}", customerId))
                .doOnError(error -> logger.error("Error fetching customer {}: {}", customerId, error.getMessage()))
                .timeout(Duration.ofSeconds(5));
    }

    // Fallback methods
    public Mono<ExternalApiModels.ProductResponse> fallbackGetProduct(String productId, Exception ex) {
        logger.warn("Fallback triggered for product {}: {}", productId, ex.getMessage());
        return Mono.error(new RuntimeException("Product service unavailable for product: " + productId, ex));
    }

    public Mono<ExternalApiModels.CustomerResponse> fallbackGetCustomer(String customerId, Exception ex) {
        logger.warn("Fallback triggered for customer {}: {}", customerId, ex.getMessage());
        return Mono.error(new RuntimeException("Customer service unavailable for customer: " + customerId, ex));
    }
}