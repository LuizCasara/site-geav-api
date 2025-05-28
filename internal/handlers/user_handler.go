package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/site-geav-api/internal/logger"
	"github.com/site-geav-api/internal/models"
	"github.com/site-geav-api/internal/repository"
)

// UserHandler handles user-related requests
type UserHandler struct {
	userRepo repository.UserRepository
	log      logger.Logger
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userRepo repository.UserRepository, log logger.Logger) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
		log:      log,
	}
}

// GetUser handles GET /users/{id} requests
func (h *UserHandler) GetUser(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract user ID from path parameters
	userID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid user ID", err, map[string]interface{}{
			"action":   "GetUser",
			"resource": "users",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid user ID")
	}

	// Get user from repository
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		h.log.Error(ctx, "Error getting user", err, map[string]interface{}{
			"action":      "GetUser",
			"resource":    "users",
			"resource_id": fmt.Sprintf("%d", userID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error getting user")
	}

	// If user not found
	if user == nil {
		h.log.Warn(ctx, "User not found", map[string]interface{}{
			"action":      "GetUser",
			"resource":    "users",
			"resource_id": fmt.Sprintf("%d", userID),
		})
		return createErrorResponse(http.StatusNotFound, "User not found")
	}

	// Log success
	h.log.Info(ctx, "User retrieved successfully", map[string]interface{}{
		"action":      "GetUser",
		"resource":    "users",
		"resource_id": fmt.Sprintf("%d", userID),
	})

	// Return user as JSON
	return createJSONResponse(http.StatusOK, user)
}

// ListUsers handles GET /users requests
func (h *UserHandler) ListUsers(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get users from repository
	users, err := h.userRepo.List(ctx)
	if err != nil {
		h.log.Error(ctx, "Error listing users", err, map[string]interface{}{
			"action":   "ListUsers",
			"resource": "users",
		})
		return createErrorResponse(http.StatusInternalServerError, "Error listing users")
	}

	// Log success
	h.log.Info(ctx, "Users listed successfully", map[string]interface{}{
		"action":   "ListUsers",
		"resource": "users",
		"count":    len(users),
	})

	// Return users as JSON
	return createJSONResponse(http.StatusOK, users)
}

// CreateUser handles POST /users requests
func (h *UserHandler) CreateUser(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse request body
	var user models.User
	if err := json.Unmarshal([]byte(request.Body), &user); err != nil {
		h.log.Error(ctx, "Invalid request body", err, map[string]interface{}{
			"action":   "CreateUser",
			"resource": "users",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid request body")
	}

	// Validate user
	if user.Username == "" || user.Password == "" || !models.IsValidRole(user.Role) {
		h.log.Warn(ctx, "Invalid user data", map[string]interface{}{
			"action":   "CreateUser",
			"resource": "users",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid user data")
	}

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Create user in repository
	userID, err := h.userRepo.Create(ctx, &user)
	if err != nil {
		h.log.Error(ctx, "Error creating user", err, map[string]interface{}{
			"action":   "CreateUser",
			"resource": "users",
		})
		return createErrorResponse(http.StatusInternalServerError, "Error creating user")
	}

	// Set user ID
	user.ID = userID

	// Log success
	h.log.Info(ctx, "User created successfully", map[string]interface{}{
		"action":      "CreateUser",
		"resource":    "users",
		"resource_id": fmt.Sprintf("%d", userID),
	})

	// Return created user as JSON
	return createJSONResponse(http.StatusCreated, user)
}

// UpdateUser handles PUT /users/{id} requests
func (h *UserHandler) UpdateUser(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract user ID from path parameters
	userID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid user ID", err, map[string]interface{}{
			"action":   "UpdateUser",
			"resource": "users",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid user ID")
	}

	// Get existing user
	existingUser, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		h.log.Error(ctx, "Error getting user", err, map[string]interface{}{
			"action":      "UpdateUser",
			"resource":    "users",
			"resource_id": fmt.Sprintf("%d", userID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error getting user")
	}

	// If user not found
	if existingUser == nil {
		h.log.Warn(ctx, "User not found", map[string]interface{}{
			"action":      "UpdateUser",
			"resource":    "users",
			"resource_id": fmt.Sprintf("%d", userID),
		})
		return createErrorResponse(http.StatusNotFound, "User not found")
	}

	// Parse request body
	var updatedUser models.User
	if err := json.Unmarshal([]byte(request.Body), &updatedUser); err != nil {
		h.log.Error(ctx, "Invalid request body", err, map[string]interface{}{
			"action":      "UpdateUser",
			"resource":    "users",
			"resource_id": fmt.Sprintf("%d", userID),
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid request body")
	}

	// Validate user
	if updatedUser.Username == "" || updatedUser.Password == "" || !models.IsValidRole(updatedUser.Role) {
		h.log.Warn(ctx, "Invalid user data", map[string]interface{}{
			"action":      "UpdateUser",
			"resource":    "users",
			"resource_id": fmt.Sprintf("%d", userID),
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid user data")
	}

	// Update user fields
	existingUser.Username = updatedUser.Username
	existingUser.Password = updatedUser.Password
	existingUser.Role = updatedUser.Role
	existingUser.UpdatedAt = time.Now()

	// Update user in repository
	if err := h.userRepo.Update(ctx, existingUser); err != nil {
		h.log.Error(ctx, "Error updating user", err, map[string]interface{}{
			"action":      "UpdateUser",
			"resource":    "users",
			"resource_id": fmt.Sprintf("%d", userID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error updating user")
	}

	// Log success
	h.log.Info(ctx, "User updated successfully", map[string]interface{}{
		"action":      "UpdateUser",
		"resource":    "users",
		"resource_id": fmt.Sprintf("%d", userID),
	})

	// Return updated user as JSON
	return createJSONResponse(http.StatusOK, existingUser)
}

// DeleteUser handles DELETE /users/{id} requests
func (h *UserHandler) DeleteUser(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract user ID from path parameters
	userID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid user ID", err, map[string]interface{}{
			"action":   "DeleteUser",
			"resource": "users",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid user ID")
	}

	// Delete user from repository
	if err := h.userRepo.Delete(ctx, userID); err != nil {
		h.log.Error(ctx, "Error deleting user", err, map[string]interface{}{
			"action":      "DeleteUser",
			"resource":    "users",
			"resource_id": fmt.Sprintf("%d", userID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error deleting user")
	}

	// Log success
	h.log.Info(ctx, "User deleted successfully", map[string]interface{}{
		"action":      "DeleteUser",
		"resource":    "users",
		"resource_id": fmt.Sprintf("%d", userID),
	})

	// Return success response
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

// Helper functions

// createJSONResponse creates a JSON response
func createJSONResponse(statusCode int, body interface{}) (events.APIGatewayProxyResponse, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return createErrorResponse(http.StatusInternalServerError, "Error creating response")
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(jsonBody),
	}, nil
}

// createErrorResponse creates an error response
func createErrorResponse(statusCode int, message string) (events.APIGatewayProxyResponse, error) {
	return createJSONResponse(statusCode, map[string]string{
		"error": message,
	})
}