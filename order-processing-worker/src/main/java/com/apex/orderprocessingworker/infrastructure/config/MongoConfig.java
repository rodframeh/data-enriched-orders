package com.apex.orderprocessingworker.infrastructure.config;

import com.mongodb.reactivestreams.client.MongoClient;
import com.mongodb.reactivestreams.client.MongoClients;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Primary;
import org.springframework.context.event.EventListener;
import org.springframework.data.mongodb.config.AbstractReactiveMongoConfiguration;
import org.springframework.data.mongodb.core.ReactiveMongoTemplate;
import org.springframework.data.mongodb.core.convert.MappingMongoConverter;
import org.springframework.data.mongodb.core.index.IndexDefinition;
import org.springframework.data.mongodb.core.index.IndexOperations;
import org.springframework.data.mongodb.core.index.Index;
import org.springframework.data.mongodb.repository.config.EnableReactiveMongoRepositories;

import jakarta.annotation.PostConstruct;

@Configuration
@EnableReactiveMongoRepositories(basePackages = "com.apex.orderprocessingworker.infrastructure.repository")
public class MongoConfig {

    @Value("${spring.data.mongodb.uri}")
    private String mongoUri;

    @Bean
    @Primary
    public MongoClient reactiveMongoClient() {
        return MongoClients.create(mongoUri);
    }

    @Bean
    @Primary
    public ReactiveMongoTemplate reactiveMongoTemplate(MongoClient mongoClient) {
        return new ReactiveMongoTemplate(mongoClient, "orderdb");
    }

    @EventListener(ApplicationReadyEvent.class)
    public void initIndexes() {
        ReactiveMongoTemplate mongoTemplate = reactiveMongoTemplate(reactiveMongoClient());

        // Create indexes for better performance using new API
        var indexOps = mongoTemplate.indexOps("orders");

        // Index on orderId for fast lookups
        IndexDefinition orderIdIndex = new Index()
                .on("orderId", org.springframework.data.domain.Sort.Direction.ASC)
                .unique();

        indexOps.ensureIndex(orderIdIndex)
                .doOnSuccess(index -> System.out.println("✓ Created unique index on orderId"))
                .doOnError(error -> System.err.println("✗ Error creating orderId index: " + error.getMessage()))
                .subscribe();

        // Index on customerId for customer queries
        IndexDefinition customerIdIndex = new Index()
                .on("customerId", org.springframework.data.domain.Sort.Direction.ASC);

        indexOps.ensureIndex(customerIdIndex)
                .doOnSuccess(index -> System.out.println("✓ Created index on customerId"))
                .doOnError(error -> System.err.println("✗ Error creating customerId index: " + error.getMessage()))
                .subscribe();

        // Index on status for status-based queries
        IndexDefinition statusIndex = new Index()
                .on("status", org.springframework.data.domain.Sort.Direction.ASC);

        indexOps.ensureIndex(statusIndex)
                .doOnSuccess(index -> System.out.println("✓ Created index on status"))
                .doOnError(error -> System.err.println("✗ Error creating status index: " + error.getMessage()))
                .subscribe();

        // Compound index on customerId and createdAt for time-based customer queries
        IndexDefinition compoundIndex = new Index()
                .on("customerId", org.springframework.data.domain.Sort.Direction.ASC)
                .on("createdAt", org.springframework.data.domain.Sort.Direction.DESC);

        indexOps.ensureIndex(compoundIndex)
                .doOnSuccess(index -> System.out.println("✓ Created compound index on customerId and createdAt"))
                .doOnError(error -> System.err.println("✗ Error creating compound index: " + error.getMessage()))
                .subscribe();
    }
}