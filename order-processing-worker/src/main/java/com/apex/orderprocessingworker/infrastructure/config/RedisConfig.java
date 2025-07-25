package com.apex.orderprocessingworker.infrastructure.config;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.data.redis.connection.ReactiveRedisConnectionFactory;
import org.springframework.data.redis.connection.RedisConnectionFactory;
import org.springframework.data.redis.connection.RedisStandaloneConfiguration;
import org.springframework.data.redis.connection.lettuce.LettuceConnectionFactory;
import org.springframework.data.redis.core.ReactiveRedisTemplate;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.data.redis.serializer.GenericJackson2JsonRedisSerializer;
import org.springframework.data.redis.serializer.RedisSerializationContext;
import org.springframework.data.redis.serializer.StringRedisSerializer;

@Configuration
public class RedisConfig {

    @Value("${spring.data.redis.host:localhost}")
    private String redisHost;

    @Value("${spring.data.redis.port:6379}")
    private int redisPort;

    @Value("${spring.data.redis.password:}")
    private String redisPassword;

    @Bean
    public RedisConnectionFactory redisConnectionFactory() {
        RedisStandaloneConfiguration configuration = new RedisStandaloneConfiguration();
        configuration.setHostName(redisHost);
        configuration.setPort(redisPort);

        if (!redisPassword.isEmpty()) {
            configuration.setPassword(redisPassword);
        }

        return new LettuceConnectionFactory(configuration);
    }

    @Bean
    public ReactiveRedisTemplate<String, String> reactiveRedisTemplate(
            RedisConnectionFactory connectionFactory) {

        StringRedisSerializer stringSerializer = StringRedisSerializer.UTF_8;

        RedisSerializationContext<String, String> serializationContext =
                RedisSerializationContext.<String, String>newSerializationContext()
                        .key(stringSerializer)
                        .value(stringSerializer)
                        .hashKey(stringSerializer)    // Esta línea faltaba y causaba el error
                        .hashValue(stringSerializer)  // Esta también es necesaria
                        .build();

        return new ReactiveRedisTemplate<>((ReactiveRedisConnectionFactory) connectionFactory, serializationContext);
    }

    @Bean
    public RedisTemplate<String, Object> redisTemplate(RedisConnectionFactory connectionFactory) {
        RedisTemplate<String, Object> template = new RedisTemplate<>();
        template.setConnectionFactory(connectionFactory);
        template.setKeySerializer(new StringRedisSerializer());
        template.setValueSerializer(new GenericJackson2JsonRedisSerializer());
        template.setHashKeySerializer(new StringRedisSerializer());
        template.setHashValueSerializer(new GenericJackson2JsonRedisSerializer());
        template.afterPropertiesSet(); // Importante para inicializar correctamente
        return template;
    }
}