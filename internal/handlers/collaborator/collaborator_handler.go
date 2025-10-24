package collaborator

import (
	"strconv"
	"strings"

	"kisanlink-ecom/entities/models/collaborator"
	collaboratorRequests "kisanlink-ecom/entities/requests/collaborator"
	"kisanlink-ecom/internal/common"
	"kisanlink-ecom/internal/middleware"
	collaboratorService "kisanlink-ecom/internal/services/collaborator"

	"github.com/gin-gonic/gin"
)

// CollaboratorHandler handles HTTP requests for collaborator operations
type CollaboratorHandler struct {
	collaboratorService collaboratorService.CollaboratorServiceInterface
}

// NewCollaboratorHandler creates a new collaborator handler
func NewCollaboratorHandler(collaboratorService collaboratorService.CollaboratorServiceInterface) *CollaboratorHandler {
	return &CollaboratorHandler{
		collaboratorService: collaboratorService,
	}
}

// CreateCollaborator godoc
// @Summary Create a new collaborator
// @Description Create a new collaborator in the platform
// @Tags collaborators
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param collaborator body object true "Collaborator data"
// @Success 201 {object} object
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 409 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/collaborators [post]
func (h *CollaboratorHandler) CreateCollaborator(c *gin.Context) {
	var req collaboratorRequests.CreateCollaboratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	collaborator, err := h.collaboratorService.CreateCollaborator(c.Request.Context(), &req, userID.(string))
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			common.Error(c, 409, "COLLABORATOR_EXISTS", err.Error(), nil)
			return
		}
		common.InternalServerError(c, "CREATE_FAILED", "Failed to create collaborator", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Created(c, collaborator, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// GetCollaboratorByID godoc
// @Summary Get collaborator by ID
// @Description Retrieve a collaborator by their ID
// @Tags collaborators
// @Accept json
// @Produce json
// @Param id path string true "Collaborator ID"
// @Success 200 {object} object
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/collaborators/{id} [get]
func (h *CollaboratorHandler) GetCollaboratorByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Collaborator ID is required", nil)
		return
	}

	collaborator, err := h.collaboratorService.GetCollaboratorByID(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, "COLLABORATOR_NOT_FOUND", "Collaborator not found", map[string]interface{}{
				"collaborator_id": id,
			})
			return
		}
		common.InternalServerError(c, "GET_FAILED", "Failed to get collaborator", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, collaborator, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// GetCollaboratorByUserID godoc
// @Summary Get collaborator by user ID
// @Description Retrieve a collaborator by their user ID from AAA service
// @Tags collaborators
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} common.Response{data=collaborator.CollaboratorResponse}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/collaborators/user/{user_id} [get]
func (h *CollaboratorHandler) GetCollaboratorByUserID(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		common.BadRequest(c, "MISSING_USER_ID", "User ID is required", nil)
		return
	}

	collaborator, err := h.collaboratorService.GetCollaboratorByUserID(c.Request.Context(), userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, "COLLABORATOR_NOT_FOUND", "Collaborator not found", map[string]interface{}{
				"user_id": userID,
			})
			return
		}
		common.InternalServerError(c, "GET_FAILED", "Failed to get collaborator", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, collaborator, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// GetCollaboratorProfile godoc
// @Summary Get detailed collaborator profile
// @Description Retrieve detailed profile information for a collaborator
// @Tags collaborators
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Collaborator ID"
// @Success 200 {object} common.Response{data=collaborator.CollaboratorProfileResponse}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/collaborators/{id}/profile [get]
func (h *CollaboratorHandler) GetCollaboratorProfile(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Collaborator ID is required", nil)
		return
	}

	// Get requestor ID from context
	requestorID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	// Check if requestor is admin (this would need proper role checking)
	isAdmin := false
	if roles, ok := c.Get("roles"); ok {
		if roleSlice, ok := roles.([]string); ok {
			for _, role := range roleSlice {
				if role == "admin" || role == "super_admin" {
					isAdmin = true
					break
				}
			}
		}
	}

	profile, err := h.collaboratorService.GetCollaboratorProfile(c.Request.Context(), id, requestorID.(string), isAdmin)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, "COLLABORATOR_NOT_FOUND", "Collaborator not found", map[string]interface{}{
				"collaborator_id": id,
			})
			return
		}
		common.InternalServerError(c, "GET_FAILED", "Failed to get collaborator profile", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, profile, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// UpdateCollaborator godoc
// @Summary Update collaborator
// @Description Update collaborator information
// @Tags collaborators
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Collaborator ID"
// @Param collaborator body collaborator.UpdateCollaboratorRequest true "Update data"
// @Success 200 {object} common.Response{data=collaborator.CollaboratorResponse}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/collaborators/{id} [put]
func (h *CollaboratorHandler) UpdateCollaborator(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Collaborator ID is required", nil)
		return
	}

	var req collaboratorRequests.UpdateCollaboratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	collaborator, err := h.collaboratorService.UpdateCollaborator(c.Request.Context(), id, &req, userID.(string))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, "COLLABORATOR_NOT_FOUND", "Collaborator not found", map[string]interface{}{
				"collaborator_id": id,
			})
			return
		}
		common.InternalServerError(c, "UPDATE_FAILED", "Failed to update collaborator", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, collaborator, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// DeleteCollaborator godoc
// @Summary Delete collaborator
// @Description Soft delete a collaborator
// @Tags collaborators
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Collaborator ID"
// @Success 204 "Collaborator deleted successfully"
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/collaborators/{id} [delete]
func (h *CollaboratorHandler) DeleteCollaborator(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Collaborator ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	err := h.collaboratorService.DeleteCollaborator(c.Request.Context(), id, userID.(string))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, "COLLABORATOR_NOT_FOUND", "Collaborator not found", map[string]interface{}{
				"collaborator_id": id,
			})
			return
		}
		common.InternalServerError(c, "DELETE_FAILED", "Failed to delete collaborator", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.Status(204)
}

// ListCollaborators godoc
// @Summary List collaborators
// @Description Retrieve a list of collaborators with filtering and pagination
// @Tags collaborators
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param collaborator_type query string false "Filter by collaborator type (FARMER, SUPPLIER, BUYER, AGENT, ADMIN)"
// @Param status query string false "Filter by status (ACTIVE, INACTIVE, SUSPENDED, PENDING)"
// @Param organization_id query string false "Filter by organization ID"
// @Param is_verified query bool false "Filter by verification status"
// @Param onboarding_completed query bool false "Filter by onboarding completion"
// @Param location query string false "Filter by location"
// @Param business_type query string false "Filter by business type"
// @Param min_trust_score query number false "Minimum trust score filter"
// @Param max_trust_score query number false "Maximum trust score filter"
// @Param search query string false "Search term"
// @Param tags query []string false "Filter by tags"
// @Param created_after query string false "Filter by creation date (ISO format)"
// @Param created_before query string false "Filter by creation date (ISO format)"
// @Param include_deleted query bool false "Include soft-deleted items (admin only)" default(false)
// @Success 200 {object} object
// @Router /api/v1/collaborators [get]
func (h *CollaboratorHandler) ListCollaborators(c *gin.Context) {
	// Extract query options (includes deleted items if user is admin and include_deleted=true)
	middleware.ExtractQueryOptions(c)

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Parse filter parameters
	filter := &collaboratorRequests.CollaboratorFilter{}

	// Collaborator type filter
	if collaboratorType := c.Query("collaborator_type"); collaboratorType != "" {
		collaboratorTypeEnum := collaborator.CollaboratorType(strings.ToUpper(collaboratorType))
		filter.CollaboratorType = &collaboratorTypeEnum
	}

	// Status filter
	if status := c.Query("status"); status != "" {
		statusEnum := collaborator.CollaboratorStatus(strings.ToUpper(status))
		filter.Status = &statusEnum
	}

	// Organization filter (default to context org if present and not specified)
	if orgID := c.Query("organization_id"); orgID != "" {
		filter.OrganizationID = &orgID
	} else if ctxOrg, ok := middleware.GetOrgID(c); ok {
		filter.OrganizationID = &ctxOrg
	}

	// Verification status filter
	if isVerifiedStr := c.Query("is_verified"); isVerifiedStr != "" {
		if isVerified, err := strconv.ParseBool(isVerifiedStr); err == nil {
			filter.IsVerified = &isVerified
		}
	}

	// Onboarding completion filter
	if onboardingCompletedStr := c.Query("onboarding_completed"); onboardingCompletedStr != "" {
		if onboardingCompleted, err := strconv.ParseBool(onboardingCompletedStr); err == nil {
			filter.OnboardingCompleted = &onboardingCompleted
		}
	}

	// Location filter
	if location := c.Query("location"); location != "" {
		filter.Location = &location
	}

	// Business type filter
	if businessType := c.Query("business_type"); businessType != "" {
		filter.BusinessType = &businessType
	}

	// Trust score range filters
	if minTrustScoreStr := c.Query("min_trust_score"); minTrustScoreStr != "" {
		if minTrustScore, err := strconv.ParseFloat(minTrustScoreStr, 64); err == nil {
			filter.MinTrustScore = &minTrustScore
		}
	}
	if maxTrustScoreStr := c.Query("max_trust_score"); maxTrustScoreStr != "" {
		if maxTrustScore, err := strconv.ParseFloat(maxTrustScoreStr, 64); err == nil {
			filter.MaxTrustScore = &maxTrustScore
		}
	}

	// Search filter
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	// Tags filter
	if tagsStr := c.Query("tags"); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		filter.Tags = tags
	}

	// Date range filters
	if createdAfter := c.Query("created_after"); createdAfter != "" {
		filter.CreatedAfter = &createdAfter
	}
	if createdBefore := c.Query("created_before"); createdBefore != "" {
		filter.CreatedBefore = &createdBefore
	}

	// Get collaborators
	collaborators, total, err := h.collaboratorService.ListCollaborators(c.Request.Context(), filter, offset, limit)
	if err != nil {
		common.InternalServerError(c, "LIST_FAILED", "Failed to list collaborators", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, collaborators, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
		Pagination: &common.PaginationMeta{
			Page:    page,
			Limit:   limit,
			Total:   total,
			HasNext: len(collaborators) == limit,
		},
	})
}

// SearchCollaborators godoc
// @Summary Search collaborators
// @Description Search collaborators with a query string
// @Tags collaborators
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param include_deleted query bool false "Include soft-deleted items (admin only)" default(false)
// @Success 200 {object} common.Response{data=[]collaborator.CollaboratorSummaryResponse,meta=common.ResponseMeta{pagination=common.PaginationMeta}}
// @Router /api/v1/collaborators/search [get]
func (h *CollaboratorHandler) SearchCollaborators(c *gin.Context) {
	// Extract query options (includes deleted items if user is admin and include_deleted=true)
	middleware.ExtractQueryOptions(c)

	// Get search query
	query := c.Query("q")
	if query == "" {
		common.BadRequest(c, "MISSING_QUERY", "Search query is required", nil)
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Search collaborators
	collaborators, total, err := h.collaboratorService.SearchCollaborators(c.Request.Context(), query, offset, limit)
	if err != nil {
		common.InternalServerError(c, "SEARCH_FAILED", "Failed to search collaborators", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, collaborators, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
		Pagination: &common.PaginationMeta{
			Page:    page,
			Limit:   limit,
			Total:   total,
			HasNext: len(collaborators) == limit,
		},
	})
}

// UpdateCollaboratorStatus godoc
// @Summary Update collaborator status
// @Description Update the status of a collaborator
// @Tags collaborators
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Collaborator ID"
// @Param status body object true "Status update data"
// @Success 200 {object} common.Response{data=collaborator.CollaboratorResponse}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/collaborators/{id}/status [patch]
func (h *CollaboratorHandler) UpdateCollaboratorStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Collaborator ID is required", nil)
		return
	}

	var req collaboratorRequests.UpdateCollaboratorStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	collaborator, err := h.collaboratorService.UpdateCollaboratorStatus(c.Request.Context(), id, &req, userID.(string))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, "COLLABORATOR_NOT_FOUND", "Collaborator not found", map[string]interface{}{
				"collaborator_id": id,
			})
			return
		}
		common.InternalServerError(c, "UPDATE_FAILED", "Failed to update collaborator status", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, collaborator, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// VerifyCollaborator godoc
// @Summary Verify collaborator
// @Description Mark a collaborator as verified
// @Tags collaborators
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Collaborator ID"
// @Param verification body object true "Verification data"
// @Success 200 {object} common.Response{data=collaborator.CollaboratorVerificationResponse}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/collaborators/{id}/verify [post]
func (h *CollaboratorHandler) VerifyCollaborator(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Collaborator ID is required", nil)
		return
	}

	var req collaboratorRequests.VerifyCollaboratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	verification, err := h.collaboratorService.VerifyCollaborator(c.Request.Context(), id, &req, userID.(string))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, "COLLABORATOR_NOT_FOUND", "Collaborator not found", map[string]interface{}{
				"collaborator_id": id,
			})
			return
		}
		common.InternalServerError(c, "VERIFY_FAILED", "Failed to verify collaborator", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, verification, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// UpdateOnboardingStep godoc
// @Summary Update onboarding step
// @Description Update the onboarding step for a collaborator
// @Tags collaborators
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Collaborator ID"
// @Param step body collaborator.UpdateOnboardingStepRequest true "Onboarding step data"
// @Success 200 {object} common.Response{data=collaborator.OnboardingStepResponse}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/collaborators/{id}/onboarding [patch]
func (h *CollaboratorHandler) UpdateOnboardingStep(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Collaborator ID is required", nil)
		return
	}

	var req collaboratorRequests.UpdateOnboardingStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	response, err := h.collaboratorService.UpdateOnboardingStep(c.Request.Context(), id, &req, userID.(string))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, "COLLABORATOR_NOT_FOUND", "Collaborator not found", map[string]interface{}{
				"collaborator_id": id,
			})
			return
		}
		common.InternalServerError(c, "UPDATE_FAILED", "Failed to update onboarding step", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, response, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// CompleteOnboarding godoc
// @Summary Complete onboarding
// @Description Mark onboarding as completed for a collaborator
// @Tags collaborators
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Collaborator ID"
// @Success 200 {object} common.Response{data=collaborator.OnboardingStepResponse}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Failure 404 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/collaborators/{id}/onboarding/complete [post]
func (h *CollaboratorHandler) CompleteOnboarding(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		common.BadRequest(c, "MISSING_ID", "Collaborator ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	response, err := h.collaboratorService.CompleteOnboarding(c.Request.Context(), id, userID.(string))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			common.NotFound(c, "COLLABORATOR_NOT_FOUND", "Collaborator not found", map[string]interface{}{
				"collaborator_id": id,
			})
			return
		}
		common.InternalServerError(c, "COMPLETE_FAILED", "Failed to complete onboarding", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, response, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// BulkUpdateCollaborators godoc
// @Summary Bulk update collaborators
// @Description Perform bulk updates on multiple collaborators
// @Tags collaborators
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param bulk body collaborator.BulkUpdateCollaboratorsRequest true "Bulk update data"
// @Success 200 {object} common.Response{data=collaborator.BulkOperationResponse}
// @Failure 400 {object} common.Response{error=common.ResponseError}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/collaborators/bulk [patch]
func (h *CollaboratorHandler) BulkUpdateCollaborators(c *gin.Context) {
	var req collaboratorRequests.BulkUpdateCollaboratorsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "INVALID_REQUEST", "Invalid request body", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("subjectID")
	if !exists {
		common.Unauthorized(c, "MISSING_USER", "User ID not found in context", nil)
		return
	}

	response, err := h.collaboratorService.BulkUpdateCollaborators(c.Request.Context(), &req, userID.(string))
	if err != nil {
		common.InternalServerError(c, "BULK_UPDATE_FAILED", "Failed to perform bulk update", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, response, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}

// GetCollaboratorStats godoc
// @Summary Get collaborator statistics
// @Description Retrieve statistics about collaborators
// @Tags collaborators
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param organization_id query string false "Filter by organization ID"
// @Success 200 {object} common.Response{data=collaborator.CollaboratorStatsResponse}
// @Failure 401 {object} common.Response{error=common.ResponseError}
// @Router /api/v1/collaborators/stats [get]
func (h *CollaboratorHandler) GetCollaboratorStats(c *gin.Context) {
	var orgID *string
	if orgIDParam := c.Query("organization_id"); orgIDParam != "" {
		orgID = &orgIDParam
	} else if ctxOrg, ok := middleware.GetOrgID(c); ok {
		orgID = &ctxOrg
	}

	stats, err := h.collaboratorService.GetCollaboratorStats(c.Request.Context(), orgID)
	if err != nil {
		common.InternalServerError(c, "STATS_FAILED", "Failed to get collaborator statistics", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	common.Success(c, stats, &common.ResponseMeta{
		TraceID: common.GetTraceID(c),
	})
}
