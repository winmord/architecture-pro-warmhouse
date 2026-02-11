package org.practicum.deviceservice.service;

import lombok.extern.slf4j.Slf4j;
import org.practicum.deviceservice.model.DeviceCommand;
import org.springframework.amqp.rabbit.core.RabbitTemplate;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import tools.jackson.databind.ObjectMapper;

import java.util.HashMap;
import java.util.Map;

@Component
@Slf4j
public class CommandPublisher {
    private final RabbitTemplate rabbitTemplate;
    private final ObjectMapper objectMapper = new ObjectMapper();

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

        try {
            String jsonMessage = objectMapper.writeValueAsString(message);
            rabbitTemplate.convertAndSend(exchange, routingKey, jsonMessage);

            log.info("Command published to RabbitMQ: {} for device: {}", command.getId(), command.getDeviceId());
        } catch (Exception e) {
            log.error("Failed to publish command");
        }
    }
}
