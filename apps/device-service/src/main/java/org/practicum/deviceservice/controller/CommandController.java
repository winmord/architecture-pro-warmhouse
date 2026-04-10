package org.practicum.deviceservice.controller;

import org.practicum.deviceservice.model.CommandRequest;
import org.practicum.deviceservice.model.DeviceCommand;
import org.practicum.deviceservice.service.CommandService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/api/v1/commands")
public class CommandController {
    private final CommandService commandService;

    public CommandController(CommandService commandService) {
        this.commandService = commandService;
    }

    @PostMapping
    public ResponseEntity<Map<String, String>> sendCommand(@RequestBody CommandRequest request) {
        DeviceCommand command = commandService.processCommand(request);

        return ResponseEntity.ok(Map.of(
                "message", "Command accepted",
                "commandId", command.getId(),
                "deviceId", command.getDeviceId(),
                "status", command.getStatus()
        ));
    }

    @GetMapping("/{commandId}")
    public ResponseEntity<DeviceCommand> getCommand(@PathVariable String commandId) {
        DeviceCommand command = commandService.getCommand(commandId);
        if (command == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(command);
    }
}
