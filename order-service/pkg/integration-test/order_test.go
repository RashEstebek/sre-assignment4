package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dreamteam/order-service/pkg/client"
	"github.com/dreamteam/order-service/pkg/db"
	"github.com/dreamteam/order-service/pkg/models"
	"github.com/dreamteam/order-service/pkg/pb"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"testing"

	"github.com/dreamteam/order-service/pkg/service"
)

func TestCreateOrderIntegration(t *testing.T) {
	testDB, err := createTestDatabase()
	assert.NoError(t, err)

	productClient := client.InitProductServiceClient("localhost:50052")
	assert.NoError(t, err)

	server := &service.Server{
		H:          db.Handler{DB: testDB},
		ProductSvc: productClient,
	}

	req := &pb.CreateOrderRequest{
		UserId:    1,
		ProductId: 1,
		Quantity:  1,
	}

	res, err := server.CreateOrder(context.Background(), req)

	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, http.StatusCreated, int(res.Status), "Expected status to be http.StatusCreated")

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	queueName := "order_queue"

	msgs, err := ch.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to set up consumer: %v", err)
	}

	go func() {
		for msg := range msgs {
			var order models.Order

			err = json.Unmarshal(msg.Body, &order)
			if err != nil {
				log.Fatalf("Failed to unmarshal message body: %v", err)
			}

			fmt.Println(string(msg.Body))
			fmt.Println(order)

			assert.Equal(t, req.ProductId, order.ProductId, "Expected the order's product ID to match the request's product ID")
			assert.Equal(t, req.UserId, order.UserId, "Expected the order's user ID to match the request's user ID")

			err = msg.Ack(false)
			if err != nil {
				log.Fatalf("Failed to acknowledge message: %v", err)
			}

			log.Fatalf("Timeout: No message received from the queue")
		}
	}()
}

func createTestDatabase() (*gorm.DB, error) {
	dsn := "host=horton.db.elephantsql.com user=ierrxilr  password=m0tFfwPVDiaNM6kV3kqpYGxlU0oqDD-Z dbname=ierrxilr   port=5432 sslmode=disable TimeZone=Asia/Kolkata"
	testDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %v", err)
	}

	// Migrate the test database
	err = testDB.AutoMigrate(&models.Order{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate test database: %v", err)
	}

	return testDB, nil
}
