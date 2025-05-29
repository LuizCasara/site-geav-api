package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/site-geav-api/internal/handlers"
	"github.com/site-geav-api/internal/logger"
	"github.com/site-geav-api/internal/repository"
)

var (
	userHandler   *handlers.UserHandler
	cancaoHandler *handlers.CancaoHandler
	lugarHandler  *handlers.LugarHandler
	log           logger.Logger
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

	// Create repositories
	userRepo := repository.NewPostgresUserRepository(db)
	cancaoRepo := repository.NewPostgresCancaoRepository(db)
	lugarRepo := repository.NewPostgresLugarRepository(db)

	// Create handlers
	userHandler = handlers.NewUserHandler(userRepo, log)
	cancaoHandler = handlers.NewCancaoHandler(cancaoRepo, log)
	lugarHandler = handlers.NewLugarHandler(lugarRepo, log)
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
		// User routes
		if request.Resource == "/users" {
			return userHandler.ListUsers(ctx, request)
		} else if request.Resource == "/users/{id}" {
			return userHandler.GetUser(ctx, request)
		}

		// Cancao routes
		if request.Resource == "/cancoes" {
			return cancaoHandler.ListCancoes(ctx, request)
		} else if request.Resource == "/cancoes/{id}" {
			return cancaoHandler.GetCancao(ctx, request)
		}

		// Lugar routes
		if request.Resource == "/lugares" {
			return lugarHandler.ListLugares(ctx, request)
		} else if request.Resource == "/lugares/{id}" {
			return lugarHandler.GetLugar(ctx, request)
		} else if request.Resource == "/lugares/{id}/ratings" {
			return lugarHandler.GetRatingsForLugar(ctx, request)
		}

	case "POST":
		// User routes
		if request.Resource == "/users" {
			return userHandler.CreateUser(ctx, request)
		}

		// Cancao routes
		if request.Resource == "/cancoes" {
			return cancaoHandler.CreateCancao(ctx, request)
		} else if request.Resource == "/cancoes/{id}/tags" {
			return cancaoHandler.AddTagToCancao(ctx, request)
		} else if request.Resource == "/cancoes/{id}/ramos" {
			return cancaoHandler.AddRamoToCancao(ctx, request)
		}

		// Lugar routes
		if request.Resource == "/lugares" {
			return lugarHandler.CreateLugar(ctx, request)
		} else if request.Resource == "/lugares/{id}/images" {
			return lugarHandler.AddImageToLugar(ctx, request)
		} else if request.Resource == "/lugares/{id}/tags" {
			return lugarHandler.AddTagToLugar(ctx, request)
		} else if request.Resource == "/lugares/{id}/ramos" {
			return lugarHandler.AddRamoToLugar(ctx, request)
		} else if request.Resource == "/lugares/{id}/ratings" {
			return lugarHandler.AddRatingToLugar(ctx, request)
		}

	case "PUT":
		// User routes
		if request.Resource == "/users/{id}" {
			return userHandler.UpdateUser(ctx, request)
		}

		// Cancao routes
		if request.Resource == "/cancoes/{id}" {
			return cancaoHandler.UpdateCancao(ctx, request)
		}

		// Lugar routes
		if request.Resource == "/lugares/{id}" {
			return lugarHandler.UpdateLugar(ctx, request)
		} else if request.Resource == "/lugares/{id}/ratings/{ratingId}" {
			return lugarHandler.UpdateRatingForLugar(ctx, request)
		}

	case "DELETE":
		// User routes
		if request.Resource == "/users/{id}" {
			return userHandler.DeleteUser(ctx, request)
		}

		// Cancao routes
		if request.Resource == "/cancoes/{id}" {
			return cancaoHandler.DeleteCancao(ctx, request)
		} else if request.Resource == "/cancoes/{id}/tags/{tagId}" {
			return cancaoHandler.RemoveTagFromCancao(ctx, request)
		} else if request.Resource == "/cancoes/{id}/ramos/{ramoId}" {
			return cancaoHandler.RemoveRamoFromCancao(ctx, request)
		}

		// Lugar routes
		if request.Resource == "/lugares/{id}" {
			return lugarHandler.DeleteLugar(ctx, request)
		} else if request.Resource == "/lugares/{id}/images/{imageId}" {
			return lugarHandler.DeleteImageFromLugar(ctx, request)
		} else if request.Resource == "/lugares/{id}/tags/{tagId}" {
			return lugarHandler.RemoveTagFromLugar(ctx, request)
		} else if request.Resource == "/lugares/{id}/ramos/{ramoId}" {
			return lugarHandler.RemoveRamoFromLugar(ctx, request)
		} else if request.Resource == "/lugares/{id}/ratings/{ratingId}" {
			return lugarHandler.DeleteRatingFromLugar(ctx, request)
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
