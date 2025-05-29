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

// CancaoHandler handles song-related requests
type CancaoHandler struct {
	cancaoRepo repository.CancaoRepository
	log        logger.Logger
}

// NewCancaoHandler creates a new CancaoHandler
func NewCancaoHandler(cancaoRepo repository.CancaoRepository, log logger.Logger) *CancaoHandler {
	return &CancaoHandler{
		cancaoRepo: cancaoRepo,
		log:        log,
	}
}

// GetCancao handles GET /cancoes/{id} requests
func (h *CancaoHandler) GetCancao(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract cancao ID from path parameters
	cancaoID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid cancao ID", err, map[string]interface{}{
			"action":   "GetCancao",
			"resource": "cancoes",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid cancao ID")
	}

	// Get cancao from repository
	cancao, err := h.cancaoRepo.GetByID(ctx, cancaoID)
	if err != nil {
		h.log.Error(ctx, "Error getting cancao", err, map[string]interface{}{
			"action":      "GetCancao",
			"resource":    "cancoes",
			"resource_id": fmt.Sprintf("%d", cancaoID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error getting cancao")
	}

	// If cancao not found
	if cancao == nil {
		h.log.Warn(ctx, "Cancao not found", map[string]interface{}{
			"action":      "GetCancao",
			"resource":    "cancoes",
			"resource_id": fmt.Sprintf("%d", cancaoID),
		})
		return createErrorResponse(http.StatusNotFound, "Cancao not found")
	}

	// Log success
	h.log.Info(ctx, "Cancao retrieved successfully", map[string]interface{}{
		"action":      "GetCancao",
		"resource":    "cancoes",
		"resource_id": fmt.Sprintf("%d", cancaoID),
	})

	// Return cancao as JSON
	return createJSONResponse(http.StatusOK, cancao)
}

// ListCancoes handles GET /cancoes requests
func (h *CancaoHandler) ListCancoes(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get cancoes from repository
	cancoes, err := h.cancaoRepo.List(ctx)
	if err != nil {
		h.log.Error(ctx, "Error listing cancoes", err, map[string]interface{}{
			"action":   "ListCancoes",
			"resource": "cancoes",
		})
		return createErrorResponse(http.StatusInternalServerError, "Error listing cancoes")
	}

	// Log success
	h.log.Info(ctx, "Cancoes listed successfully", map[string]interface{}{
		"action":   "ListCancoes",
		"resource": "cancoes",
		"count":    len(cancoes),
	})

	// Return cancoes as JSON
	return createJSONResponse(http.StatusOK, cancoes)
}

// CreateCancao handles POST /cancoes requests
func (h *CancaoHandler) CreateCancao(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse request body
	var cancao models.Cancao
	if err := json.Unmarshal([]byte(request.Body), &cancao); err != nil {
		h.log.Error(ctx, "Invalid request body", err, map[string]interface{}{
			"action":   "CreateCancao",
			"resource": "cancoes",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid request body")
	}

	// Validate cancao
	if cancao.Nome == "" {
		h.log.Warn(ctx, "Invalid cancao data: nome is required", map[string]interface{}{
			"action":   "CreateCancao",
			"resource": "cancoes",
		})
		return createErrorResponse(http.StatusBadRequest, "Nome is required")
	}

	// Set timestamps
	now := time.Now()
	cancao.CreatedAt = now
	cancao.UpdatedAt = now

	// Create cancao in repository
	cancaoID, err := h.cancaoRepo.Create(ctx, &cancao)
	if err != nil {
		h.log.Error(ctx, "Error creating cancao", err, map[string]interface{}{
			"action":   "CreateCancao",
			"resource": "cancoes",
		})
		return createErrorResponse(http.StatusInternalServerError, "Error creating cancao")
	}

	// Set cancao ID
	cancao.ID = cancaoID

	// Process related entities if provided
	if len(cancao.Tags) > 0 {
		for _, tag := range cancao.Tags {
			if err := h.cancaoRepo.AddTag(ctx, cancaoID, tag.ID); err != nil {
				h.log.Error(ctx, "Error adding tag to cancao", err, map[string]interface{}{
					"action":      "CreateCancao",
					"resource":    "cancoes",
					"resource_id": fmt.Sprintf("%d", cancaoID),
					"tag_id":      fmt.Sprintf("%d", tag.ID),
				})
				// Continue with other tags even if one fails
			}
		}
	}

	if len(cancao.Ramos) > 0 {
		for _, ramo := range cancao.Ramos {
			if err := h.cancaoRepo.AddRamo(ctx, cancaoID, ramo.ID); err != nil {
				h.log.Error(ctx, "Error adding ramo to cancao", err, map[string]interface{}{
					"action":      "CreateCancao",
					"resource":    "cancoes",
					"resource_id": fmt.Sprintf("%d", cancaoID),
					"ramo_id":     fmt.Sprintf("%d", ramo.ID),
				})
				// Continue with other ramos even if one fails
			}
		}
	}

	// Log success
	h.log.Info(ctx, "Cancao created successfully", map[string]interface{}{
		"action":      "CreateCancao",
		"resource":    "cancoes",
		"resource_id": fmt.Sprintf("%d", cancaoID),
	})

	// Return created cancao as JSON
	return createJSONResponse(http.StatusCreated, cancao)
}

// UpdateCancao handles PUT /cancoes/{id} requests
func (h *CancaoHandler) UpdateCancao(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract cancao ID from path parameters
	cancaoID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid cancao ID", err, map[string]interface{}{
			"action":   "UpdateCancao",
			"resource": "cancoes",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid cancao ID")
	}

	// Get existing cancao
	existingCancao, err := h.cancaoRepo.GetByID(ctx, cancaoID)
	if err != nil {
		h.log.Error(ctx, "Error getting cancao", err, map[string]interface{}{
			"action":      "UpdateCancao",
			"resource":    "cancoes",
			"resource_id": fmt.Sprintf("%d", cancaoID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error getting cancao")
	}

	// If cancao not found
	if existingCancao == nil {
		h.log.Warn(ctx, "Cancao not found", map[string]interface{}{
			"action":      "UpdateCancao",
			"resource":    "cancoes",
			"resource_id": fmt.Sprintf("%d", cancaoID),
		})
		return createErrorResponse(http.StatusNotFound, "Cancao not found")
	}

	// Parse request body
	var updatedCancao models.Cancao
	if err := json.Unmarshal([]byte(request.Body), &updatedCancao); err != nil {
		h.log.Error(ctx, "Invalid request body", err, map[string]interface{}{
			"action":      "UpdateCancao",
			"resource":    "cancoes",
			"resource_id": fmt.Sprintf("%d", cancaoID),
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid request body")
	}

	// Validate cancao
	if updatedCancao.Nome == "" {
		h.log.Warn(ctx, "Invalid cancao data: nome is required", map[string]interface{}{
			"action":      "UpdateCancao",
			"resource":    "cancoes",
			"resource_id": fmt.Sprintf("%d", cancaoID),
		})
		return createErrorResponse(http.StatusBadRequest, "Nome is required")
	}

	// Update cancao fields
	existingCancao.Nome = updatedCancao.Nome
	existingCancao.LinkYoutube = updatedCancao.LinkYoutube
	existingCancao.Letra = updatedCancao.Letra
	existingCancao.UserID = updatedCancao.UserID
	existingCancao.UpdatedAt = time.Now()

	// Update cancao in repository
	if err := h.cancaoRepo.Update(ctx, existingCancao); err != nil {
		h.log.Error(ctx, "Error updating cancao", err, map[string]interface{}{
			"action":      "UpdateCancao",
			"resource":    "cancoes",
			"resource_id": fmt.Sprintf("%d", cancaoID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error updating cancao")
	}

	// Log success
	h.log.Info(ctx, "Cancao updated successfully", map[string]interface{}{
		"action":      "UpdateCancao",
		"resource":    "cancoes",
		"resource_id": fmt.Sprintf("%d", cancaoID),
	})

	// Return updated cancao as JSON
	return createJSONResponse(http.StatusOK, existingCancao)
}

// DeleteCancao handles DELETE /cancoes/{id} requests
func (h *CancaoHandler) DeleteCancao(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract cancao ID from path parameters
	cancaoID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid cancao ID", err, map[string]interface{}{
			"action":   "DeleteCancao",
			"resource": "cancoes",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid cancao ID")
	}

	// Delete cancao from repository
	if err := h.cancaoRepo.Delete(ctx, cancaoID); err != nil {
		h.log.Error(ctx, "Error deleting cancao", err, map[string]interface{}{
			"action":      "DeleteCancao",
			"resource":    "cancoes",
			"resource_id": fmt.Sprintf("%d", cancaoID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error deleting cancao")
	}

	// Log success
	h.log.Info(ctx, "Cancao deleted successfully", map[string]interface{}{
		"action":      "DeleteCancao",
		"resource":    "cancoes",
		"resource_id": fmt.Sprintf("%d", cancaoID),
	})

	// Return success response
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

// AddTagToCancao handles POST /cancoes/{id}/tags requests
func (h *CancaoHandler) AddTagToCancao(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract cancao ID from path parameters
	cancaoID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid cancao ID", err, map[string]interface{}{
			"action":   "AddTagToCancao",
			"resource": "cancoes",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid cancao ID")
	}

	// Parse request body
	var requestBody struct {
		TagID int `json:"tag_id"`
	}
	if err := json.Unmarshal([]byte(request.Body), &requestBody); err != nil {
		h.log.Error(ctx, "Invalid request body", err, map[string]interface{}{
			"action":      "AddTagToCancao",
			"resource":    "cancoes",
			"resource_id": fmt.Sprintf("%d", cancaoID),
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid request body")
	}

	// Add tag to cancao
	if err := h.cancaoRepo.AddTag(ctx, cancaoID, requestBody.TagID); err != nil {
		h.log.Error(ctx, "Error adding tag to cancao", err, map[string]interface{}{
			"action":      "AddTagToCancao",
			"resource":    "cancoes",
			"resource_id": fmt.Sprintf("%d", cancaoID),
			"tag_id":      fmt.Sprintf("%d", requestBody.TagID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error adding tag to cancao")
	}

	// Log success
	h.log.Info(ctx, "Tag added to cancao successfully", map[string]interface{}{
		"action":      "AddTagToCancao",
		"resource":    "cancoes",
		"resource_id": fmt.Sprintf("%d", cancaoID),
		"tag_id":      fmt.Sprintf("%d", requestBody.TagID),
	})

	// Return success response
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

// RemoveTagFromCancao handles DELETE /cancoes/{id}/tags/{tagId} requests
func (h *CancaoHandler) RemoveTagFromCancao(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract cancao ID and tag ID from path parameters
	cancaoID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid cancao ID", err, map[string]interface{}{
			"action":   "RemoveTagFromCancao",
			"resource": "cancoes",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid cancao ID")
	}

	tagID, err := strconv.Atoi(request.PathParameters["tagId"])
	if err != nil {
		h.log.Error(ctx, "Invalid tag ID", err, map[string]interface{}{
			"action":      "RemoveTagFromCancao",
			"resource":    "cancoes",
			"resource_id": fmt.Sprintf("%d", cancaoID),
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid tag ID")
	}

	// Remove tag from cancao
	if err := h.cancaoRepo.RemoveTag(ctx, cancaoID, tagID); err != nil {
		h.log.Error(ctx, "Error removing tag from cancao", err, map[string]interface{}{
			"action":      "RemoveTagFromCancao",
			"resource":    "cancoes",
			"resource_id": fmt.Sprintf("%d", cancaoID),
			"tag_id":      fmt.Sprintf("%d", tagID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error removing tag from cancao")
	}

	// Log success
	h.log.Info(ctx, "Tag removed from cancao successfully", map[string]interface{}{
		"action":      "RemoveTagFromCancao",
		"resource":    "cancoes",
		"resource_id": fmt.Sprintf("%d", cancaoID),
		"tag_id":      fmt.Sprintf("%d", tagID),
	})

	// Return success response
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

// AddRamoToCancao handles POST /cancoes/{id}/ramos requests
func (h *CancaoHandler) AddRamoToCancao(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract cancao ID from path parameters
	cancaoID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid cancao ID", err, map[string]interface{}{
			"action":   "AddRamoToCancao",
			"resource": "cancoes",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid cancao ID")
	}

	// Parse request body
	var requestBody struct {
		RamoID int `json:"ramo_id"`
	}
	if err := json.Unmarshal([]byte(request.Body), &requestBody); err != nil {
		h.log.Error(ctx, "Invalid request body", err, map[string]interface{}{
			"action":      "AddRamoToCancao",
			"resource":    "cancoes",
			"resource_id": fmt.Sprintf("%d", cancaoID),
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid request body")
	}

	// Add ramo to cancao
	if err := h.cancaoRepo.AddRamo(ctx, cancaoID, requestBody.RamoID); err != nil {
		h.log.Error(ctx, "Error adding ramo to cancao", err, map[string]interface{}{
			"action":      "AddRamoToCancao",
			"resource":    "cancoes",
			"resource_id": fmt.Sprintf("%d", cancaoID),
			"ramo_id":     fmt.Sprintf("%d", requestBody.RamoID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error adding ramo to cancao")
	}

	// Log success
	h.log.Info(ctx, "Ramo added to cancao successfully", map[string]interface{}{
		"action":      "AddRamoToCancao",
		"resource":    "cancoes",
		"resource_id": fmt.Sprintf("%d", cancaoID),
		"ramo_id":     fmt.Sprintf("%d", requestBody.RamoID),
	})

	// Return success response
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

// RemoveRamoFromCancao handles DELETE /cancoes/{id}/ramos/{ramoId} requests
func (h *CancaoHandler) RemoveRamoFromCancao(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract cancao ID and ramo ID from path parameters
	cancaoID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid cancao ID", err, map[string]interface{}{
			"action":   "RemoveRamoFromCancao",
			"resource": "cancoes",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid cancao ID")
	}

	ramoID, err := strconv.Atoi(request.PathParameters["ramoId"])
	if err != nil {
		h.log.Error(ctx, "Invalid ramo ID", err, map[string]interface{}{
			"action":      "RemoveRamoFromCancao",
			"resource":    "cancoes",
			"resource_id": fmt.Sprintf("%d", cancaoID),
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid ramo ID")
	}

	// Remove ramo from cancao
	if err := h.cancaoRepo.RemoveRamo(ctx, cancaoID, ramoID); err != nil {
		h.log.Error(ctx, "Error removing ramo from cancao", err, map[string]interface{}{
			"action":      "RemoveRamoFromCancao",
			"resource":    "cancoes",
			"resource_id": fmt.Sprintf("%d", cancaoID),
			"ramo_id":     fmt.Sprintf("%d", ramoID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error removing ramo from cancao")
	}

	// Log success
	h.log.Info(ctx, "Ramo removed from cancao successfully", map[string]interface{}{
		"action":      "RemoveRamoFromCancao",
		"resource":    "cancoes",
		"resource_id": fmt.Sprintf("%d", cancaoID),
		"ramo_id":     fmt.Sprintf("%d", ramoID),
	})

	// Return success response
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}
