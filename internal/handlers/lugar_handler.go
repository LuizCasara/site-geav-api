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

// LugarHandler handles place-related requests
type LugarHandler struct {
	lugarRepo repository.LugarRepository
	log       logger.Logger
}

// NewLugarHandler creates a new LugarHandler
func NewLugarHandler(lugarRepo repository.LugarRepository, log logger.Logger) *LugarHandler {
	return &LugarHandler{
		lugarRepo: lugarRepo,
		log:       log,
	}
}

// GetLugar handles GET /lugares/{id} requests
func (h *LugarHandler) GetLugar(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract lugar ID from path parameters
	lugarID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid lugar ID", err, map[string]interface{}{
			"action":   "GetLugar",
			"resource": "lugares",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid lugar ID")
	}

	// Get lugar from repository
	lugar, err := h.lugarRepo.GetByID(ctx, lugarID)
	if err != nil {
		h.log.Error(ctx, "Error getting lugar", err, map[string]interface{}{
			"action":      "GetLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error getting lugar")
	}

	// If lugar not found
	if lugar == nil {
		h.log.Warn(ctx, "Lugar not found", map[string]interface{}{
			"action":      "GetLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
		})
		return createErrorResponse(http.StatusNotFound, "Lugar not found")
	}

	// Log success
	h.log.Info(ctx, "Lugar retrieved successfully", map[string]interface{}{
		"action":      "GetLugar",
		"resource":    "lugares",
		"resource_id": fmt.Sprintf("%d", lugarID),
	})

	// Return lugar as JSON
	return createJSONResponse(http.StatusOK, lugar)
}

// ListLugares handles GET /lugares requests
func (h *LugarHandler) ListLugares(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get lugares from repository
	lugares, err := h.lugarRepo.List(ctx)
	if err != nil {
		h.log.Error(ctx, "Error listing lugares", err, map[string]interface{}{
			"action":   "ListLugares",
			"resource": "lugares",
		})
		return createErrorResponse(http.StatusInternalServerError, "Error listing lugares")
	}

	// Log success
	h.log.Info(ctx, "Lugares listed successfully", map[string]interface{}{
		"action":   "ListLugares",
		"resource": "lugares",
		"count":    len(lugares),
	})

	// Return lugares as JSON
	return createJSONResponse(http.StatusOK, lugares)
}

// CreateLugar handles POST /lugares requests
func (h *LugarHandler) CreateLugar(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse request body
	var lugar models.Lugar
	if err := json.Unmarshal([]byte(request.Body), &lugar); err != nil {
		h.log.Error(ctx, "Invalid request body", err, map[string]interface{}{
			"action":   "CreateLugar",
			"resource": "lugares",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid request body")
	}

	// Validate lugar
	if lugar.NomeLocal == "" {
		h.log.Warn(ctx, "Invalid lugar data: nome_local is required", map[string]interface{}{
			"action":   "CreateLugar",
			"resource": "lugares",
		})
		return createErrorResponse(http.StatusBadRequest, "Nome local is required")
	}

	// Set timestamps
	now := time.Now()
	lugar.CreatedAt = now
	lugar.UpdatedAt = now

	// Create lugar in repository
	lugarID, err := h.lugarRepo.Create(ctx, &lugar)
	if err != nil {
		h.log.Error(ctx, "Error creating lugar", err, map[string]interface{}{
			"action":   "CreateLugar",
			"resource": "lugares",
		})
		return createErrorResponse(http.StatusInternalServerError, "Error creating lugar")
	}

	// Set lugar ID
	lugar.ID = lugarID

	// Process related entities if provided
	if len(lugar.Images) > 0 {
		for i := range lugar.Images {
			lugar.Images[i].LugarID = lugarID
			lugar.Images[i].CreatedAt = now
			imageID, err := h.lugarRepo.AddImage(ctx, lugar.Images[i])
			if err != nil {
				h.log.Error(ctx, "Error adding image to lugar", err, map[string]interface{}{
					"action":      "CreateLugar",
					"resource":    "lugares",
					"resource_id": fmt.Sprintf("%d", lugarID),
				})
				// Continue with other images even if one fails
			} else {
				lugar.Images[i].ID = imageID
			}
		}
	}

	if len(lugar.Tags) > 0 {
		for _, tag := range lugar.Tags {
			if err := h.lugarRepo.AddTag(ctx, lugarID, tag.ID); err != nil {
				h.log.Error(ctx, "Error adding tag to lugar", err, map[string]interface{}{
					"action":      "CreateLugar",
					"resource":    "lugares",
					"resource_id": fmt.Sprintf("%d", lugarID),
					"tag_id":      fmt.Sprintf("%d", tag.ID),
				})
				// Continue with other tags even if one fails
			}
		}
	}

	if len(lugar.Ramos) > 0 {
		for _, ramo := range lugar.Ramos {
			if err := h.lugarRepo.AddRamo(ctx, lugarID, ramo.ID); err != nil {
				h.log.Error(ctx, "Error adding ramo to lugar", err, map[string]interface{}{
					"action":      "CreateLugar",
					"resource":    "lugares",
					"resource_id": fmt.Sprintf("%d", lugarID),
					"ramo_id":     fmt.Sprintf("%d", ramo.ID),
				})
				// Continue with other ramos even if one fails
			}
		}
	}

	// Log success
	h.log.Info(ctx, "Lugar created successfully", map[string]interface{}{
		"action":      "CreateLugar",
		"resource":    "lugares",
		"resource_id": fmt.Sprintf("%d", lugarID),
	})

	// Return created lugar as JSON
	return createJSONResponse(http.StatusCreated, lugar)
}

// UpdateLugar handles PUT /lugares/{id} requests
func (h *LugarHandler) UpdateLugar(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract lugar ID from path parameters
	lugarID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid lugar ID", err, map[string]interface{}{
			"action":   "UpdateLugar",
			"resource": "lugares",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid lugar ID")
	}

	// Get existing lugar
	existingLugar, err := h.lugarRepo.GetByID(ctx, lugarID)
	if err != nil {
		h.log.Error(ctx, "Error getting lugar", err, map[string]interface{}{
			"action":      "UpdateLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error getting lugar")
	}

	// If lugar not found
	if existingLugar == nil {
		h.log.Warn(ctx, "Lugar not found", map[string]interface{}{
			"action":      "UpdateLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
		})
		return createErrorResponse(http.StatusNotFound, "Lugar not found")
	}

	// Parse request body
	var updatedLugar models.Lugar
	if err := json.Unmarshal([]byte(request.Body), &updatedLugar); err != nil {
		h.log.Error(ctx, "Invalid request body", err, map[string]interface{}{
			"action":      "UpdateLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid request body")
	}

	// Validate lugar
	if updatedLugar.NomeLocal == "" {
		h.log.Warn(ctx, "Invalid lugar data: nome_local is required", map[string]interface{}{
			"action":      "UpdateLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
		})
		return createErrorResponse(http.StatusBadRequest, "Nome local is required")
	}

	// Update lugar fields
	existingLugar.NomeLocal = updatedLugar.NomeLocal
	existingLugar.NomeDonoLocal = updatedLugar.NomeDonoLocal
	existingLugar.TelefoneParaContato = updatedLugar.TelefoneParaContato
	existingLugar.LinkGoogleMaps = updatedLugar.LinkGoogleMaps
	existingLugar.LinkSite = updatedLugar.LinkSite
	existingLugar.EnderecoCompleto = updatedLugar.EnderecoCompleto
	existingLugar.LocalPublico = updatedLugar.LocalPublico
	existingLugar.ValorFixo = updatedLugar.ValorFixo
	existingLugar.ValorIndividual = updatedLugar.ValorIndividual
	existingLugar.UserID = updatedLugar.UserID
	existingLugar.UpdatedAt = time.Now()

	// Update lugar in repository
	if err := h.lugarRepo.Update(ctx, existingLugar); err != nil {
		h.log.Error(ctx, "Error updating lugar", err, map[string]interface{}{
			"action":      "UpdateLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error updating lugar")
	}

	// Log success
	h.log.Info(ctx, "Lugar updated successfully", map[string]interface{}{
		"action":      "UpdateLugar",
		"resource":    "lugares",
		"resource_id": fmt.Sprintf("%d", lugarID),
	})

	// Return updated lugar as JSON
	return createJSONResponse(http.StatusOK, existingLugar)
}

// DeleteLugar handles DELETE /lugares/{id} requests
func (h *LugarHandler) DeleteLugar(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract lugar ID from path parameters
	lugarID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid lugar ID", err, map[string]interface{}{
			"action":   "DeleteLugar",
			"resource": "lugares",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid lugar ID")
	}

	// Delete lugar from repository
	if err := h.lugarRepo.Delete(ctx, lugarID); err != nil {
		h.log.Error(ctx, "Error deleting lugar", err, map[string]interface{}{
			"action":      "DeleteLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error deleting lugar")
	}

	// Log success
	h.log.Info(ctx, "Lugar deleted successfully", map[string]interface{}{
		"action":      "DeleteLugar",
		"resource":    "lugares",
		"resource_id": fmt.Sprintf("%d", lugarID),
	})

	// Return success response
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

// AddImageToLugar handles POST /lugares/{id}/images requests
func (h *LugarHandler) AddImageToLugar(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract lugar ID from path parameters
	lugarID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid lugar ID", err, map[string]interface{}{
			"action":   "AddImageToLugar",
			"resource": "lugares",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid lugar ID")
	}

	// Parse request body
	var image models.LugarImage
	if err := json.Unmarshal([]byte(request.Body), &image); err != nil {
		h.log.Error(ctx, "Invalid request body", err, map[string]interface{}{
			"action":      "AddImageToLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid request body")
	}

	// Set lugar ID and created at
	image.LugarID = lugarID
	image.CreatedAt = time.Now()

	// Add image to lugar
	imageID, err := h.lugarRepo.AddImage(ctx, &image)
	if err != nil {
		h.log.Error(ctx, "Error adding image to lugar", err, map[string]interface{}{
			"action":      "AddImageToLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error adding image to lugar")
	}

	// Set image ID
	image.ID = imageID

	// Log success
	h.log.Info(ctx, "Image added to lugar successfully", map[string]interface{}{
		"action":      "AddImageToLugar",
		"resource":    "lugares",
		"resource_id": fmt.Sprintf("%d", lugarID),
		"image_id":    fmt.Sprintf("%d", imageID),
	})

	// Return created image as JSON
	return createJSONResponse(http.StatusCreated, image)
}

// DeleteImageFromLugar handles DELETE /lugares/{id}/images/{imageId} requests
func (h *LugarHandler) DeleteImageFromLugar(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract lugar ID and image ID from path parameters
	_, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid lugar ID", err, map[string]interface{}{
			"action":   "DeleteImageFromLugar",
			"resource": "lugares",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid lugar ID")
	}

	imageID, err := strconv.Atoi(request.PathParameters["imageId"])
	if err != nil {
		h.log.Error(ctx, "Invalid image ID", err, map[string]interface{}{
			"action":   "DeleteImageFromLugar",
			"resource": "lugares",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid image ID")
	}

	// Delete image from lugar
	if err := h.lugarRepo.DeleteImage(ctx, imageID); err != nil {
		h.log.Error(ctx, "Error deleting image from lugar", err, map[string]interface{}{
			"action":   "DeleteImageFromLugar",
			"resource": "lugares",
			"image_id": fmt.Sprintf("%d", imageID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error deleting image from lugar")
	}

	// Log success
	h.log.Info(ctx, "Image deleted from lugar successfully", map[string]interface{}{
		"action":   "DeleteImageFromLugar",
		"resource": "lugares",
		"image_id": fmt.Sprintf("%d", imageID),
	})

	// Return success response
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

// AddTagToLugar handles POST /lugares/{id}/tags requests
func (h *LugarHandler) AddTagToLugar(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract lugar ID from path parameters
	lugarID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid lugar ID", err, map[string]interface{}{
			"action":   "AddTagToLugar",
			"resource": "lugares",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid lugar ID")
	}

	// Parse request body
	var requestBody struct {
		TagID int `json:"tag_id"`
	}
	if err := json.Unmarshal([]byte(request.Body), &requestBody); err != nil {
		h.log.Error(ctx, "Invalid request body", err, map[string]interface{}{
			"action":      "AddTagToLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid request body")
	}

	// Add tag to lugar
	if err := h.lugarRepo.AddTag(ctx, lugarID, requestBody.TagID); err != nil {
		h.log.Error(ctx, "Error adding tag to lugar", err, map[string]interface{}{
			"action":      "AddTagToLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
			"tag_id":      fmt.Sprintf("%d", requestBody.TagID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error adding tag to lugar")
	}

	// Log success
	h.log.Info(ctx, "Tag added to lugar successfully", map[string]interface{}{
		"action":      "AddTagToLugar",
		"resource":    "lugares",
		"resource_id": fmt.Sprintf("%d", lugarID),
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

// RemoveTagFromLugar handles DELETE /lugares/{id}/tags/{tagId} requests
func (h *LugarHandler) RemoveTagFromLugar(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract lugar ID and tag ID from path parameters
	lugarID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid lugar ID", err, map[string]interface{}{
			"action":   "RemoveTagFromLugar",
			"resource": "lugares",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid lugar ID")
	}

	tagID, err := strconv.Atoi(request.PathParameters["tagId"])
	if err != nil {
		h.log.Error(ctx, "Invalid tag ID", err, map[string]interface{}{
			"action":      "RemoveTagFromLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid tag ID")
	}

	// Remove tag from lugar
	if err := h.lugarRepo.RemoveTag(ctx, lugarID, tagID); err != nil {
		h.log.Error(ctx, "Error removing tag from lugar", err, map[string]interface{}{
			"action":      "RemoveTagFromLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
			"tag_id":      fmt.Sprintf("%d", tagID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error removing tag from lugar")
	}

	// Log success
	h.log.Info(ctx, "Tag removed from lugar successfully", map[string]interface{}{
		"action":      "RemoveTagFromLugar",
		"resource":    "lugares",
		"resource_id": fmt.Sprintf("%d", lugarID),
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

// AddRamoToLugar handles POST /lugares/{id}/ramos requests
func (h *LugarHandler) AddRamoToLugar(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract lugar ID from path parameters
	lugarID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid lugar ID", err, map[string]interface{}{
			"action":   "AddRamoToLugar",
			"resource": "lugares",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid lugar ID")
	}

	// Parse request body
	var requestBody struct {
		RamoID int `json:"ramo_id"`
	}
	if err := json.Unmarshal([]byte(request.Body), &requestBody); err != nil {
		h.log.Error(ctx, "Invalid request body", err, map[string]interface{}{
			"action":      "AddRamoToLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid request body")
	}

	// Add ramo to lugar
	if err := h.lugarRepo.AddRamo(ctx, lugarID, requestBody.RamoID); err != nil {
		h.log.Error(ctx, "Error adding ramo to lugar", err, map[string]interface{}{
			"action":      "AddRamoToLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
			"ramo_id":     fmt.Sprintf("%d", requestBody.RamoID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error adding ramo to lugar")
	}

	// Log success
	h.log.Info(ctx, "Ramo added to lugar successfully", map[string]interface{}{
		"action":      "AddRamoToLugar",
		"resource":    "lugares",
		"resource_id": fmt.Sprintf("%d", lugarID),
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

// RemoveRamoFromLugar handles DELETE /lugares/{id}/ramos/{ramoId} requests
func (h *LugarHandler) RemoveRamoFromLugar(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract lugar ID and ramo ID from path parameters
	lugarID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid lugar ID", err, map[string]interface{}{
			"action":   "RemoveRamoFromLugar",
			"resource": "lugares",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid lugar ID")
	}

	ramoID, err := strconv.Atoi(request.PathParameters["ramoId"])
	if err != nil {
		h.log.Error(ctx, "Invalid ramo ID", err, map[string]interface{}{
			"action":      "RemoveRamoFromLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid ramo ID")
	}

	// Remove ramo from lugar
	if err := h.lugarRepo.RemoveRamo(ctx, lugarID, ramoID); err != nil {
		h.log.Error(ctx, "Error removing ramo from lugar", err, map[string]interface{}{
			"action":      "RemoveRamoFromLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
			"ramo_id":     fmt.Sprintf("%d", ramoID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error removing ramo from lugar")
	}

	// Log success
	h.log.Info(ctx, "Ramo removed from lugar successfully", map[string]interface{}{
		"action":      "RemoveRamoFromLugar",
		"resource":    "lugares",
		"resource_id": fmt.Sprintf("%d", lugarID),
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

// AddRatingToLugar handles POST /lugares/{id}/ratings requests
func (h *LugarHandler) AddRatingToLugar(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract lugar ID from path parameters
	lugarID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid lugar ID", err, map[string]interface{}{
			"action":   "AddRatingToLugar",
			"resource": "lugares",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid lugar ID")
	}

	// Parse request body
	var rating models.LugarRating
	if err := json.Unmarshal([]byte(request.Body), &rating); err != nil {
		h.log.Error(ctx, "Invalid request body", err, map[string]interface{}{
			"action":      "AddRatingToLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid request body")
	}

	// Validate rating
	if rating.Rating < 1 || rating.Rating > 5 {
		h.log.Warn(ctx, "Invalid rating value", map[string]interface{}{
			"action":      "AddRatingToLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
			"rating":      rating.Rating,
		})
		return createErrorResponse(http.StatusBadRequest, "Rating must be between 1 and 5")
	}

	// Set lugar ID and date
	rating.LugarID = lugarID
	rating.Date = time.Now()

	// Add rating to lugar
	ratingID, err := h.lugarRepo.AddRating(ctx, &rating)
	if err != nil {
		h.log.Error(ctx, "Error adding rating to lugar", err, map[string]interface{}{
			"action":      "AddRatingToLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error adding rating to lugar")
	}

	// Set rating ID
	rating.ID = ratingID

	// Log success
	h.log.Info(ctx, "Rating added to lugar successfully", map[string]interface{}{
		"action":      "AddRatingToLugar",
		"resource":    "lugares",
		"resource_id": fmt.Sprintf("%d", lugarID),
		"rating_id":   fmt.Sprintf("%d", ratingID),
		"rating":      rating.Rating,
	})

	// Return created rating as JSON
	return createJSONResponse(http.StatusCreated, rating)
}

// UpdateRatingForLugar handles PUT /lugares/{id}/ratings/{ratingId} requests
func (h *LugarHandler) UpdateRatingForLugar(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract lugar ID and rating ID from path parameters
	lugarID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid lugar ID", err, map[string]interface{}{
			"action":   "UpdateRatingForLugar",
			"resource": "lugares",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid lugar ID")
	}

	ratingID, err := strconv.Atoi(request.PathParameters["ratingId"])
	if err != nil {
		h.log.Error(ctx, "Invalid rating ID", err, map[string]interface{}{
			"action":      "UpdateRatingForLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid rating ID")
	}

	// Parse request body
	var rating models.LugarRating
	if err := json.Unmarshal([]byte(request.Body), &rating); err != nil {
		h.log.Error(ctx, "Invalid request body", err, map[string]interface{}{
			"action":      "UpdateRatingForLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
			"rating_id":   fmt.Sprintf("%d", ratingID),
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid request body")
	}

	// Validate rating
	if rating.Rating < 1 || rating.Rating > 5 {
		h.log.Warn(ctx, "Invalid rating value", map[string]interface{}{
			"action":      "UpdateRatingForLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
			"rating_id":   fmt.Sprintf("%d", ratingID),
			"rating":      rating.Rating,
		})
		return createErrorResponse(http.StatusBadRequest, "Rating must be between 1 and 5")
	}

	// Set rating ID, lugar ID, and date
	rating.ID = ratingID
	rating.LugarID = lugarID
	rating.Date = time.Now()

	// Update rating for lugar
	if err := h.lugarRepo.UpdateRating(ctx, &rating); err != nil {
		h.log.Error(ctx, "Error updating rating for lugar", err, map[string]interface{}{
			"action":      "UpdateRatingForLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
			"rating_id":   fmt.Sprintf("%d", ratingID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error updating rating for lugar")
	}

	// Log success
	h.log.Info(ctx, "Rating updated for lugar successfully", map[string]interface{}{
		"action":      "UpdateRatingForLugar",
		"resource":    "lugares",
		"resource_id": fmt.Sprintf("%d", lugarID),
		"rating_id":   fmt.Sprintf("%d", ratingID),
		"rating":      rating.Rating,
	})

	// Return updated rating as JSON
	return createJSONResponse(http.StatusOK, rating)
}

// DeleteRatingFromLugar handles DELETE /lugares/{id}/ratings/{ratingId} requests
func (h *LugarHandler) DeleteRatingFromLugar(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract lugar ID and rating ID from path parameters
	lugarID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid lugar ID", err, map[string]interface{}{
			"action":   "DeleteRatingFromLugar",
			"resource": "lugares",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid lugar ID")
	}

	ratingID, err := strconv.Atoi(request.PathParameters["ratingId"])
	if err != nil {
		h.log.Error(ctx, "Invalid rating ID", err, map[string]interface{}{
			"action":      "DeleteRatingFromLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid rating ID")
	}

	// Delete rating from lugar
	if err := h.lugarRepo.DeleteRating(ctx, ratingID); err != nil {
		h.log.Error(ctx, "Error deleting rating from lugar", err, map[string]interface{}{
			"action":      "DeleteRatingFromLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
			"rating_id":   fmt.Sprintf("%d", ratingID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error deleting rating from lugar")
	}

	// Log success
	h.log.Info(ctx, "Rating deleted from lugar successfully", map[string]interface{}{
		"action":      "DeleteRatingFromLugar",
		"resource":    "lugares",
		"resource_id": fmt.Sprintf("%d", lugarID),
		"rating_id":   fmt.Sprintf("%d", ratingID),
	})

	// Return success response
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

// GetRatingsForLugar handles GET /lugares/{id}/ratings requests
func (h *LugarHandler) GetRatingsForLugar(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract lugar ID from path parameters
	lugarID, err := strconv.Atoi(request.PathParameters["id"])
	if err != nil {
		h.log.Error(ctx, "Invalid lugar ID", err, map[string]interface{}{
			"action":   "GetRatingsForLugar",
			"resource": "lugares",
		})
		return createErrorResponse(http.StatusBadRequest, "Invalid lugar ID")
	}

	// Get ratings for lugar
	ratings, err := h.lugarRepo.GetRatings(ctx, lugarID)
	if err != nil {
		h.log.Error(ctx, "Error getting ratings for lugar", err, map[string]interface{}{
			"action":      "GetRatingsForLugar",
			"resource":    "lugares",
			"resource_id": fmt.Sprintf("%d", lugarID),
		})
		return createErrorResponse(http.StatusInternalServerError, "Error getting ratings for lugar")
	}

	// Log success
	h.log.Info(ctx, "Ratings retrieved for lugar successfully", map[string]interface{}{
		"action":      "GetRatingsForLugar",
		"resource":    "lugares",
		"resource_id": fmt.Sprintf("%d", lugarID),
		"count":       len(ratings),
	})

	// Return ratings as JSON
	return createJSONResponse(http.StatusOK, ratings)
}
