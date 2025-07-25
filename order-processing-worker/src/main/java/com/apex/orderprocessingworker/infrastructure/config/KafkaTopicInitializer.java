package com.apex.orderprocessingworker.infrastructure.config;

import org.apache.kafka.clients.admin.AdminClient;
import org.apache.kafka.clients.admin.CreateTopicsResult;
import org.apache.kafka.clients.admin.NewTopic;
import org.apache.kafka.common.errors.TopicExistsException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.event.EventListener;
import org.springframework.kafka.core.KafkaAdmin;
import org.springframework.stereotype.Component;

import java.util.Collections;
import java.util.concurrent.ExecutionException;

@Component
public class KafkaTopicInitializer {

    private static final Logger logger = LoggerFactory.getLogger(KafkaTopicInitializer.class);

    private final KafkaAdmin kafkaAdmin;

    @Value("${app.kafka.topics.orders}")
    private String ordersTopic;

    public KafkaTopicInitializer(KafkaAdmin kafkaAdmin) {
        this.kafkaAdmin = kafkaAdmin;
    }

    @EventListener(ApplicationReadyEvent.class)
    public void createTopics() {
        logger.info("Creating Kafka topics...");

        try (AdminClient adminClient = AdminClient.create(kafkaAdmin.getConfigurationProperties())) {

            // Create orders topic
            NewTopic orderEventsTopic = new NewTopic(ordersTopic, 3, (short) 1);

            CreateTopicsResult result = adminClient.createTopics(Collections.singletonList(orderEventsTopic));

            // Wait for topic creation
            result.all().get();

            logger.info("✅ Successfully created topic: {}", ordersTopic);

        } catch (ExecutionException e) {
            if (e.getCause() instanceof TopicExistsException) {
                logger.info("✅ Topic already exists: {}", ordersTopic);
            } else {
                logger.error("❌ Error creating topic {}: {}", ordersTopic, e.getMessage());
            }
        } catch (Exception e) {
            logger.error("❌ Unexpected error creating topics: {}", e.getMessage(), e);
        }
    }
}