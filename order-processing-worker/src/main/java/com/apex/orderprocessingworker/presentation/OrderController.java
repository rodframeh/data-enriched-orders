package com.apex.orderprocessingworker.presentation;

import com.apex.orderprocessingworker.application.service.OrderProcessingService;
import com.apex.orderprocessingworker.domain.entity.Order;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import reactor.core.publisher.Flux;
import reactor.core.publisher.Mono;

@RestController
@RequestMapping("/api/orders")
public class OrderController {

    private final OrderProcessingService orderProcessingService;

    public OrderController(OrderProcessingService orderProcessingService) {
        this.orderProcessingService = orderProcessingService;
    }

    @GetMapping("/{orderId}")
    public Mono<ResponseEntity<Order>> getOrder(@PathVariable String orderId) {
        return orderProcessingService.findByOrderId(orderId)
                .map(ResponseEntity::ok)
                .defaultIfEmpty(ResponseEntity.notFound().build());
    }

    @GetMapping("/customer/{customerId}")
    public Flux<Order> getOrdersByCustomer(@PathVariable String customerId) {
        return orderProcessingService.findByCustomerId(customerId);
    }

    @GetMapping("/health")
    public Mono<ResponseEntity<String>> health() {
        return Mono.just(ResponseEntity.ok("Order Processing Worker is running"));
    }
}
