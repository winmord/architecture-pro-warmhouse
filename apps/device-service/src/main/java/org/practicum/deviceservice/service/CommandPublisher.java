package org.practicum.deviceservice.service;

import org.practicum.deviceservice.model.DeviceCommand;
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.HashMap;
import java.util.Map;

@Component
public class CommandPublisher {
    private final RabbitTemplate rabbitTemplate;

    @Value("${app.rabbitmq.exchange}")
    private String exchange;

    @Value("${app.rabbitmq.routing-key}")
    private String routingKey;

    public CommandPublisher(RabbitTemplate rabbitTemplate) {
        this.rabbitTemplate = rabbitTemplate;
    }

    public void publishCommand(DeviceCommand command) {
        Map<String, Object> message = new HashMap<>();
        message.put("commandId", command.getId());
        message.put("deviceId", command.getDeviceId());
        message.put("command", command.getCommand());
        message.put("parameters", command.getParameters());
        message.put("sourceType", command.getSourceType());
        message.put("sourceId", command.getSourceId());
        message.put("priority", command.getPriority());
        message.put("timestamp", System.currentTimeMillis());

        rabbitTemplate.convertAndSend(exchange, routingKey, message);

        System.out.println("Command published to RabbitMQ: " + command.getId() +
                " for device: " + command.getDeviceId());
    }
}
