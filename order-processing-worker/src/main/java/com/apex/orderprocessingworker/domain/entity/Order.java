package com.apex.orderprocessingworker.domain.entity;

import org.springframework.data.annotation.Id;
import org.springframework.data.mongodb.core.mapping.Document;
import org.springframework.data.mongodb.core.mapping.Field;

import java.math.BigDecimal;
import java.util.List;

@Document(collection = "orders")
public record Order(
        @Id
        String id,

        @Field("orderId")
        String orderId,

        @Field("customerId")
        String customerId,

        @Field("products")
        List<OrderProduct> products

) {

    public static Order create(String orderId, String customerId, List<OrderProduct> products) {

        return new Order(
                null,
                orderId,
                customerId,
                products
        );
    }

    public record OrderProduct(
            String productId,
            String name,
            String description,
            BigDecimal price,
            Integer quantity
    ) {}

}