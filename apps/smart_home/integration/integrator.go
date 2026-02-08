package integration

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type DeviceCommandMessage struct {
	CommandID  string                 `json:"commandId"`
	DeviceID   string                 `json:"deviceId"`
	Command    string                 `json:"command"`
	Parameters map[string]interface{} `json:"parameters"`
	SourceType string                 `json:"sourceType"`
	SourceID   string                 `json:"sourceId"`
	Priority   int                    `json:"priority"`
	Timestamp  int64                  `json:"timestamp"`
}

var (
	rabbitConn *amqp.Connection
	rabbitCh   *amqp.Channel
)

func InitializeRabbitMQ() (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial("amqp://user:pass@localhost:5672/")
	if err != nil {
		log.Printf("%s: %v", "Failed to connect to RabbitMQ", err)
		return nil, nil
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("%s: %v", "Failed to open a channel", err)
		conn.Close()
		return nil, nil
	}

	err = ch.ExchangeDeclare(
		"device.exchange",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("%s: %v", "Failed to declare exchange", err)
		ch.Close()
		conn.Close()
		return nil, nil
	}

	commandsQueue, err := ch.QueueDeclare(
		"device.commands.queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("%s: %v", "Failed to declare queue", err)
		ch.Close()
		conn.Close()
		return nil, nil
	}

	err = ch.QueueBind(
		commandsQueue.Name,
		"device.commands",
		"device.exchange",
		false,
		nil,
	)
	if err != nil {
		log.Printf("%s: %v", "Failed to bind queue", err)
		ch.Close()
		conn.Close()
		return nil, nil
	}

	rabbitConn = conn
	rabbitCh = ch

	go startMessageConsumer()

	fmt.Println("Connected to RabbitMQ")
	fmt.Println("Listening for device commands...")

	return conn, ch
}

func startMessageConsumer() {
	if rabbitCh == nil {
		log.Println("RabbitMQ channel is not initialized")
		return
	}

	messages, err := rabbitCh.Consume(
		"device.commands.queue",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("%s: %v", "Failed to register consumer", err)
		return
	}

	for msg := range messages {
		go handleCommand(rabbitCh, msg.Body)
	}
}

func CloseRabbitMQ() {
	if rabbitCh != nil {
		rabbitCh.Close()
	}
	if rabbitConn != nil {
		rabbitConn.Close()
	}
	log.Println("RabbitMQ connection closed")
}

func handleCommand(ch *amqp.Channel, body []byte) {
	log.Printf("Received command: %s", string(body))

	var cmd DeviceCommandMessage
	if err := json.Unmarshal(body, &cmd); err != nil {
		log.Printf("Error unmarshaling command: %v", err)
		return
	}

	log.Printf("Processing command: ID=%s, Device=%s, Command=%s, Priority=%d",
		cmd.CommandID, cmd.DeviceID, cmd.Command, cmd.Priority)

	sendTelemetry(ch, cmd.DeviceID)
}

func sendTelemetry(ch *amqp.Channel, deviceID string) {

	telemetryTypes := []string{"temperature", "humidity", "brightness", "power_consumption"}
	units := map[string]string{
		"temperature":       "C",
		"humidity":          "%",
		"brightness":        "lux",
		"power_consumption": "W",
	}

	readings := []map[string]interface{}{}
	for _, telemetryType := range telemetryTypes {
		value := 20.0 + float64(time.Now().UnixNano()%100)/10
		if telemetryType == "humidity" {
			value = 40.0 + float64(time.Now().UnixNano()%30)
		} else if telemetryType == "brightness" {
			value = 300.0 + float64(time.Now().UnixNano()%700)
		} else if telemetryType == "power_consumption" {
			value = 50.0 + float64(time.Now().UnixNano()%150)
		}

		readings = append(readings, map[string]interface{}{
			"type":  telemetryType,
			"value": value,
			"unit":  units[telemetryType],
		})
	}

	/*c.JSON(http.StatusOK, gin.H{
		"location":    tempData.Location,
		"value":       tempData.Value,
		"unit":        tempData.Unit,
		"status":      tempData.Status,
		"timestamp":   tempData.Timestamp,
		"description": tempData.Description,
	})*/

	telemetry := map[string]interface{}{
		"device_id": deviceID,

		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	body, _ := json.Marshal(telemetry)

	err := ch.Publish(
		"device.exchange",
		"device.telemetry",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		log.Printf("Failed to publish telemetry: %v", err)
	} else {
		fmt.Printf("Sent telemetry for device: %s\n", deviceID)
	}
}

func generateID() string {
	return fmt.Sprintf("id-%d", time.Now().UnixNano())
}
