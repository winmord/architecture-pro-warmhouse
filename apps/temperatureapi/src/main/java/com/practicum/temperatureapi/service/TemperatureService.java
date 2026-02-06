package com.practicum.temperatureapi.service;

import com.practicum.temperatureapi.model.TemperatureResponse;
import org.springframework.stereotype.Service;

import java.sql.Timestamp;
import java.time.Instant;
import java.util.Random;

@Service
public class TemperatureService {
    public TemperatureService() {
    }

    public TemperatureResponse getTemperatureByLocation(String location) {
        String sensorId = switch (location) {
            case "Living Room" -> "1";
            case "Bedroom" -> "2";
            case "Kitchen" -> "3";
            default -> "0";
        };

        float value = getRandomTemperature(10f, 30f);

        return TemperatureResponse.builder()
                .value(value)
                .unit("°C")
                .timestamp(Timestamp.from(Instant.now()))
                .location(location)
                .status("active")
                .sensorId(sensorId)
                .sensorType("temperature")
                .description("Temperature sensor in " + location)
                .build();
    }

    public TemperatureResponse getTemperatureBySensorId(String sensorId) {
        String location = switch (sensorId) {
            case "1" -> "Living Room";
            case "2" -> "Bedroom";
            case "3" -> "Kitchen";
            default -> "Unknown location";
        };

        float value = getRandomTemperature(12f, 28f);

        return TemperatureResponse.builder()
                .value(value)
                .unit("°C")
                .timestamp(Timestamp.from(Instant.now()))
                .location(location)
                .status("active")
                .sensorId(sensorId)
                .sensorType("temperature")
                .description("Temperature sensor in " + location)
                .build();
    }

    private float getRandomTemperature(float min, float max) {
        Random rand = new Random();
        return rand.nextFloat() * (max - min) + min;
    }
}
