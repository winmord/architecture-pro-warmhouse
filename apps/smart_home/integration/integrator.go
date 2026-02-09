package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"smarthome/db"
	"smarthome/models"
	"smarthome/services"

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

func InitializeRabbitMQ(db *db.DB, temperatureService *services.TemperatureService) (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial("amqp://user:pass@rabbitmq:5672/")
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

	go startMessageConsumer(db, temperatureService)

	fmt.Println("Connected to RabbitMQ")
	fmt.Println("Listening for device commands...")

	return conn, ch
}

func startMessageConsumer(db *db.DB, temperatureService *services.TemperatureService) {
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
		go handleCommand(db, temperatureService, msg.Body)
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

func handleCommand(db *db.DB, temperatureService *services.TemperatureService, body []byte) {
	log.Printf("Received command: %s", string(body))

	var cmd DeviceCommandMessage
	if err := json.Unmarshal(body, &cmd); err != nil {
		log.Printf("Error unmarshaling command: %v", err)
		return
	}

	if cmd.Command == "create_sensor" {
		CreateSensor(db, cmd.DeviceID, cmd.Parameters)
	} else if cmd.Command == "get_all_sensors" {
		GetSensors(db, temperatureService, cmd.DeviceID)
	} else {
		log.Println("Not implemented yet")
	}

	log.Printf("Processing command: ID=%s, Device=%s, Command=%s, Priority=%d",
		cmd.CommandID, cmd.DeviceID, cmd.Command, cmd.Priority)
}

func CreateSensor(db *db.DB, deviceID string, parameters map[string]interface{}) {
	var sensorCreate models.SensorCreate
	sensorCreate.Name = parameters["name"].(string)
	sensorCreate.Type = models.SensorType(parameters["type"].(string))
	sensorCreate.Location = parameters["location"].(string)
	sensorCreate.Unit = parameters["unit"].(string)

	sensor, err := db.CreateSensor(context.Background(), sensorCreate)
	if err == nil {
		sendTelemetry(deviceID, sensor)
	}
}

func GetSensors(db *db.DB, temperatureService *services.TemperatureService, deviceID string) {
	sensors, err := db.GetSensors(context.Background())
	if err == nil {
		for i, sensor := range sensors {
			if sensor.Type == models.Temperature {
				tempData, err := temperatureService.GetTemperatureByID(fmt.Sprintf("%d", sensor.ID))
				if err == nil {
					sensors[i].Value = tempData.Value
					sensors[i].Status = tempData.Status
					sensors[i].LastUpdated = tempData.Timestamp
					log.Printf("Updated temperature data for sensor %d from external API", sensor.ID)
					sendTelemetry(deviceID, sensors[i])
				} else {
					log.Printf("Failed to fetch temperature data for sensor %d: %v", sensor.ID, err)
				}
			}
		}
	}
}

func sendTelemetry(deviceID string, sensor models.Sensor) {
	telemetry := map[string]interface{}{
		"device_id": deviceID,
		"type":      sensor.Type,
		"value":     sensor.Value,
		"unit":      sensor.Unit,
	}

	body, _ := json.Marshal(telemetry)

	err := rabbitCh.Publish(
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
