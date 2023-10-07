package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dreamteam/product-service/pkg/utils"
	amqp "github.com/rabbitmq/amqp091-go"
	"net/http"

	"github.com/dreamteam/product-service/pkg/db"
	"github.com/dreamteam/product-service/pkg/models"
	pb "github.com/dreamteam/product-service/pkg/pb"
)

type Server struct {
	H   db.Handler
	Jwt utils.JwtWrapper
}

func (s *Server) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	var product models.Product

	product.Name = req.Name
	product.Stock = req.Stock
	product.Price = req.Price

	if result := s.H.DB.Create(&product); result.Error != nil {
		return &pb.CreateProductResponse{
			Status: http.StatusConflict,
			Error:  result.Error.Error(),
		}, nil
	}

	return &pb.CreateProductResponse{
		Status: http.StatusCreated,
		Id:     product.Id,
	}, nil
}

func (s *Server) FindOne(ctx context.Context, req *pb.FindOneRequest) (*pb.FindOneResponse, error) {
	var product models.Product

	if result := s.H.DB.First(&product, req.Id); result.Error != nil {
		return &pb.FindOneResponse{
			Status: http.StatusNotFound,
			Error:  result.Error.Error(),
		}, nil
	}

	data := &pb.FindOneData{
		Id:    product.Id,
		Name:  product.Name,
		Stock: product.Stock,
		Price: product.Price,
	}

	return &pb.FindOneResponse{
		Status: http.StatusOK,
		Data:   data,
	}, nil
}

//func (s *Server) DecreaseStock(ctx context.Context, req *pb.DecreaseStockRequest) (*pb.DecreaseStockResponse, error) {
//
//	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
//	if err != nil {
//		//todo
//	}
//	defer conn.Close()
//
//	ch, err := conn.Channel()
//	if err != nil {
//		//todo
//	}
//	defer ch.Close()
//	queueName := "order_queue"
//	queue, err := ch.QueueDeclare(
//		queueName, // Queue name
//		false,     // Durable (messages are lost if RabbitMQ restarts)
//		false,     // Auto-delete (queue is deleted when there are no consumers)
//		false,     // Exclusive (only this connection can consume from the queue)
//		false,     // No-wait (wait for the server's response)
//		nil,       // Arguments (optional)
//	)
//	if err != nil {
//		log.Fatalf("Failed to declare a queue: %v", err)
//	}
//
//	msgs, err := ch.Consume(
//		queue.Name, // Queue name
//		"",         // Consumer name (empty string for auto-generated)
//		true,       // Auto-acknowledge messages
//		false,      // Exclusive
//		false,      // No-local
//		false,      // No-wait
//		nil,        // Arguments
//	)
//	if err != nil {
//		//todo
//	}
//
//	go func() {
//		for msg := range msgs {
//			// Process the received message
//			message1 := Order{}
//			json.Unmarshal(msg.Body, &message1)
//			fmt.Println(message1)
//		}
//	}()
//	var product models.Product
//
//	if result := s.H.DB.First(&product, req.Id); result.Error != nil {
//		return &pb.DecreaseStockResponse{
//			Status: http.StatusNotFound,
//			Error:  result.Error.Error(),
//		}, nil
//	}
//
//	if product.Stock <= 0 {
//		return &pb.DecreaseStockResponse{
//			Status: http.StatusConflict,
//			Error:  "Stock too low",
//		}, nil
//	}
//
//	var log models.StockDecreaseLog
//
//	if result := s.H.DB.Where(&models.StockDecreaseLog{OrderId: req.OrderId}).First(&log); result.Error == nil {
//		return &pb.DecreaseStockResponse{
//			Status: http.StatusConflict,
//			Error:  "Stock already decreased",
//		}, nil
//	}
//
//	product.Stock = product.Stock - 1
//
//	s.H.DB.Save(&product)
//
//	log.OrderId = req.OrderId
//	log.ProductRefer = product.Id
//
//	s.H.DB.Create(&log)
//
//	return &pb.DecreaseStockResponse{
//		Status: http.StatusOK,
//	}, nil
//}

func (s *Server) DecreaseStock(ctx context.Context, req *pb.DecreaseStockRequest) (*pb.DecreaseStockResponse, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		// Handle error
		return nil, err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		// Handle error
		return nil, err
	}
	defer ch.Close()

	queueName := "order_queue"
	queue, err := ch.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		// Handle error
		return nil, err
	}

	msgs, err := ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		// Handle error
		return nil, err
	}

	go func() {
		for msg := range msgs {
			// Process the received message
			message1 := Order{}
			json.Unmarshal(msg.Body, &message1)
			fmt.Println(message1)
		}
	}()

	var product models.Product

	if result := s.H.DB.First(&product, req.Id); result.Error != nil {
		return &pb.DecreaseStockResponse{
			Status: http.StatusNotFound,
			Error:  result.Error.Error(),
		}, nil
	}

	if product.Stock <= 0 {
		return &pb.DecreaseStockResponse{
			Status: http.StatusConflict,
			Error:  "Stock too low",
		}, nil
	}

	var log models.StockDecreaseLog

	if result := s.H.DB.Where(&models.StockDecreaseLog{ProductRefer: product.Id}).First(&log); result.Error == nil {
		log.OrderId = req.OrderId
		s.H.DB.Save(&log)
	} else {
		log.OrderId = req.OrderId
		log.ProductRefer = product.Id
		s.H.DB.Create(&log)
	}

	product.Stock = product.Stock - 1

	s.H.DB.Save(&product)

	return &pb.DecreaseStockResponse{
		Status: http.StatusOK,
	}, nil
}

type Order struct {
	Id        int64 `json:"id" gorm:"primaryKey"`
	Price     int64 `json:"price"`
	ProductId int64 `json:"product_id"`
	UserId    int64 `json:"user_id"`
}
