package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/site-geav-api/internal/handlers"
	"github.com/site-geav-api/internal/logger"
	"github.com/site-geav-api/internal/repository"
)

var (
	userHandler *handlers.UserHandler
	log         logger.Logger
)

func init() {
	// Initialize logger
	cwClient, err := createCloudWatchClient()
	if err != nil {
		panic(err)
	}

	// Create loggers
	cloudWatchLogger := logger.NewCloudWatchLogger(cwClient, "site-geav-api", "SiteGeav/API")
	
	// Initialize database connection
	db, err := repository.InitDB()
	if err != nil {
		panic(err)
	}
	
	// Create database logger
	dbLogger := logger.NewDBLogger(db, "site-geav-api", "api_logs")
	
	// Create composite logger
	log = logger.NewCompositeLogger(cloudWatchLogger, dbLogger)
	
	// Create user repository
	userRepo := repository.NewPostgresUserRepository(db)
	
	// Create user handler
	userHandler = handlers.NewUserHandler(userRepo, log)
}

func createCloudWatchClient() (*cloudwatch.Client, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}
	
	// Create CloudWatch client
	return cloudwatch.NewFromConfig(cfg), nil
}

func router(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Add request ID to context
	if requestID, ok := request.Headers["x-request-id"]; ok {
		ctx = context.WithValue(ctx, "requestID", requestID)
	}
	
	// Route request based on HTTP method and path
	switch request.HTTPMethod {
	case "GET":
		if request.Resource == "/users" {
			return userHandler.ListUsers(ctx, request)
		} else if request.Resource == "/users/{id}" {
			return userHandler.GetUser(ctx, request)
		}
	case "POST":
		if request.Resource == "/users" {
			return userHandler.CreateUser(ctx, request)
		}
	case "PUT":
		if request.Resource == "/users/{id}" {
			return userHandler.UpdateUser(ctx, request)
		}
	case "DELETE":
		if request.Resource == "/users/{id}" {
			return userHandler.DeleteUser(ctx, request)
		}
	}
	
	// Return 404 if no route matches
	return events.APIGatewayProxyResponse{
		StatusCode: 404,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"error":"Not Found"}`,
	}, nil
}

func main() {
	// Start Lambda handler
	lambda.Start(router)
}