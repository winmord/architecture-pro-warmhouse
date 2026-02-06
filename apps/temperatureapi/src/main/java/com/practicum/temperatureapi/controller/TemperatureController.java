package com.practicum.temperatureapi.controller;

import com.practicum.temperatureapi.model.TemperatureResponse;
import com.practicum.temperatureapi.service.TemperatureService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.Optional;

@RestController
public class TemperatureController {
    private final TemperatureService temperatureService;

    public TemperatureController(TemperatureService temperatureService) {
        this.temperatureService = temperatureService;
    }

    @GetMapping("/temperature")
    public ResponseEntity<TemperatureResponse> getTemperatureByLocation(@RequestParam Optional<String> location) {
        if (location.isPresent()) {
            TemperatureResponse temperatureResponse = temperatureService.getTemperatureByLocation(location.get());
            return ResponseEntity.ok(temperatureResponse);
        }
        return ResponseEntity.badRequest().build();
    }

    @GetMapping("/temperature/{sensorId}")
    public ResponseEntity<TemperatureResponse> getTemperatureById(@PathVariable String sensorId) {
        TemperatureResponse temperatureResponse = temperatureService.getTemperatureBySensorId(sensorId);
        return ResponseEntity.ok(temperatureResponse);
    }
}
