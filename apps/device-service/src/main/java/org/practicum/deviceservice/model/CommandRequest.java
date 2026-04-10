package org.practicum.deviceservice.model;

import lombok.Data;

import java.util.Map;

@Data
public class CommandRequest {
    private String deviceId;
    private String command;
    private Map<String, Object> parameters;
    private String sourceType;
    private String sourceId;
    private Integer priority = 0;
}
