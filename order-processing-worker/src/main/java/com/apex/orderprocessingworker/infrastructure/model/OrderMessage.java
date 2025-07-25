package com.apex.orderprocessingworker.infrastructure.model;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotEmpty;
import jakarta.validation.constraints.NotNull;

import java.util.List;

public record OrderMessage(
        @NotBlank(message = "Order ID cannot be blank")
        String orderId,

        @NotBlank(message = "Customer ID cannot be blank")
        String customerId,

        @NotEmpty(message = "Products list cannot be empty")
        List<ProductItem> products
) {

    public record ProductItem(
            @NotBlank(message = "Product ID cannot be blank")
            String productId,

            @NotNull(message = "Quantity cannot be null")
            Integer quantity
    ) {}
}