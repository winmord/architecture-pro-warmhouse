package org.practicum.deviceservice.configuration;

import org.springframework.amqp.core.Binding;
import org.springframework.amqp.core.BindingBuilder;
import org.springframework.amqp.core.Queue;
import org.springframework.amqp.core.TopicExchange;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class RabbitMQConfig {

    @Value("${app.rabbitmq.exchange}")
    private String exchange;

    @Value("${app.rabbitmq.routing-key}")
    private String routingKey;

    private static final String QUEUE_NAME = "device.commands.queue";

    @Bean
    public Queue commandsQueue() {
        return new Queue(QUEUE_NAME, true);
    }

    @Bean
    public TopicExchange deviceExchange() {
        return new TopicExchange(exchange);
    }

    @Bean
    public Binding commandsBinding() {
        return BindingBuilder.bind(commandsQueue())
                .to(deviceExchange())
                .with(routingKey);
    }
}
