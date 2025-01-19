package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"CodingStandards/CodingPractice/internal/handler"
	"CodingStandards/CodingPractice/internal/repository"
	"CodingStandards/CodingPractice/internal/service"
	"CodingStandards/CodingPractice/internal/validator"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Echo instance
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Validator setup
	e.Validator = validator.NewValidator()

	// MongoDB connection setup
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Initialize dependencies
	repo := repository.NewMongoUserRepository(client, "echoapp", "users")
	service := service.NewUserService(*repo)
	handler := handler.NewUserHandler(service)

	// Routes setup
	e.POST("/signup", handler.SignUp)
	e.POST("/login", handler.Login)

	// Graceful shutdown setup
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	go func() {
		if err := e.Start(":8080"); err != nil {
			log.Printf("Shutting down the server: %v", err)
		}
	}()

	<-shutdown
	log.Println("Received shutdown signal")

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to gracefully shut down: %v", err)
	}
	log.Println("Server shut down gracefully")
}
