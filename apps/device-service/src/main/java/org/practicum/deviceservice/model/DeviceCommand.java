package org.practicum.deviceservice.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;
import java.time.LocalDateTime;
import java.util.Map;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class DeviceCommand {
    private String id;
    private String deviceId;
    private String command;
    private Map<String, Object> parameters;
    private String sourceType;
    private String sourceId;
    private String status;
    private LocalDateTime sentAt;
    private Integer priority;
}
