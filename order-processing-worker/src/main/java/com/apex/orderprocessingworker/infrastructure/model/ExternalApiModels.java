package com.apex.orderprocessingworker.infrastructure.model;

import java.math.BigDecimal;

public class ExternalApiModels {

    public record ProductResponse(
            String id,
            String name,
            String description,
            BigDecimal price,
            String category,
            Boolean active
    ) {}

    public record CustomerResponse(
            String id,
            String name,
            String email,
            String phone,
            Boolean active,
            CustomerStatus status
    ) {}

    public enum CustomerStatus {
        ACTIVE,
        INACTIVE,
        BLOCKED,
        PENDING
    }

}
