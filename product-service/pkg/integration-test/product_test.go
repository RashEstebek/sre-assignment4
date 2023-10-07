package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dreamteam/product-service/pkg/models"
	"github.com/dreamteam/product-service/pkg/pb"
	"github.com/dreamteam/product-service/pkg/utils"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"testing"

	"github.com/dreamteam/product-service/pkg/db"
	"github.com/dreamteam/product-service/pkg/services"
	"github.com/stretchr/testify/assert"
)

func TestCreateProductIntegration(t *testing.T) {
	testDB, err := createTestDatabase()
	assert.NoError(t, err)

	server := services.Server{
		H: db.Handler{DB: testDB},
	}

	req := &pb.CreateProductRequest{
		Name:  "Test Product",
		Stock: 20,
		Price: 100,
	}

	res, err := server.CreateProduct(context.Background(), req)
	assert.NoError(t, err, "CreateProduct returned an error")
	assert.Equal(t, http.StatusCreated, int(res.Status), "Incorrect response status")
	assert.NotZero(t, res.Id, "Product ID should not be zero")
}

func TestFindOneIntegration(t *testing.T) {
	testDB, err := createTestDatabase()
	assert.NoError(t, err)

	server := services.Server{
		H: db.Handler{DB: testDB},
		Jwt: utils.JwtWrapper{
			SecretKey:       "your-secret-key",
			ExpirationHours: 3600,
		},
	}

	product := models.Product{
		Name:  "Test Product",
		Stock: 10,
		Price: 100,
	}
	testDB.Create(&product)

	req := &pb.FindOneRequest{
		Id: product.Id,
	}

	resp, err := server.FindOne(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, int64(http.StatusOK), resp.Status)
	assert.NotNil(t, resp.Data)
	assert.Equal(t, req.Id, resp.Data.Id)
	assert.Equal(t, product.Name, resp.Data.Name)
	assert.Equal(t, product.Stock, resp.Data.Stock)
	assert.Equal(t, product.Price, resp.Data.Price)
}

func TestDecreaseStockIntegration(t *testing.T) {
	testDB, err := createTestDatabase()
	assert.NoError(t, err)

	server := services.Server{
		H: db.Handler{DB: testDB},
		Jwt: utils.JwtWrapper{
			SecretKey:       "your-secret-key",
			ExpirationHours: 3600,
		},
	}

	product := models.Product{
		Name:  "Test Product",
		Stock: 10,
		Price: 100,
	}
	testDB.Create(&product)

	order := models.Order{
		Id:        1,
		Price:     product.Price,
		ProductId: product.Id,
		UserId:    1,
	}
	orderJSON, _ := json.Marshal(order)

	conn, _ := amqp.Dial("amqp://guest:guest@localhost:5672/")
	defer conn.Close()

	ch, _ := conn.Channel()
	defer ch.Close()
	queueName := "order_queue"
	queue, _ := ch.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)

	ch.Publish(
		"",
		queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        orderJSON,
		},
	)

	req := &pb.DecreaseStockRequest{
		Id:      product.Id,
		OrderId: order.Id,
	}

	resp, err := server.DecreaseStock(context.Background(), req)
	assert.NoError(t, err)
	fmt.Println(resp)
	updatedProduct := models.Product{}
	testDB.First(&updatedProduct, product.Id)
	assert.Equal(t, product.Stock-1, updatedProduct.Stock)
}

func createTestDatabase() (*gorm.DB, error) {
	dsn := "host=horton.db.elephantsql.com user=eonvrlzx  password=DkoRZdNcXel3ZDsDjjDcySzDIkQSYc4k dbname=eonvrlzx  port=5432 sslmode=disable TimeZone=Asia/Kolkata"
	testDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %v", err)
	}

	err = testDB.AutoMigrate(&models.Product{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate test database: %v", err)
	}

	return testDB, nil
}
