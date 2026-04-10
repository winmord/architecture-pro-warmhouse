package org.practicum.deviceservice.service;

import lombok.extern.slf4j.Slf4j;
import org.practicum.deviceservice.model.CommandRequest;
import org.practicum.deviceservice.model.DeviceCommand;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;

@Service
@Slf4j
public class CommandService {

    private final CommandPublisher commandPublisher;

    private final Map<String, DeviceCommand> commandStore = new ConcurrentHashMap<>();

    public CommandService(CommandPublisher commandPublisher) {
        this.commandPublisher = commandPublisher;
    }

    public DeviceCommand processCommand(CommandRequest request) {
        log.info("Processing command for device: {}", request.getDeviceId());

        String commandId = UUID.randomUUID().toString();

        DeviceCommand command = DeviceCommand.builder()
                .id(commandId)
                .deviceId(request.getDeviceId())
                .command(request.getCommand())
                .parameters(request.getParameters() != null ? request.getParameters() : new HashMap<>())
                .sourceType(request.getSourceType())
                .sourceId(request.getSourceId())
                .status("SENT")
                .sentAt(LocalDateTime.now())
                .priority(request.getPriority())
                .build();

        commandStore.put(commandId, command);
        commandPublisher.publishCommand(command);

        return command;
    }

    public DeviceCommand getCommand(String commandId) {
        return commandStore.get(commandId);
    }

    public void updateCommandStatus(String commandId, String status, Map<String, Object> result) {
        DeviceCommand command = commandStore.get(commandId);
        if (command != null) {
            command.setStatus(status);
            commandStore.put(commandId, command);
            log.info("Updated command {} to status: {}", commandId, status);
        }
    }
}
