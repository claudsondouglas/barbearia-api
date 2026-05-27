// Package scheduleshttp implementa os handlers HTTP para as rotas de agendamentos.
package scheduleshttp

import (
	"errors"
	"net/http"
	"strconv"

	"barbearia-api/app"
	"barbearia-api/app/middleware"
	"barbearia-api/app/schedules/actions"
	"barbearia-api/models"

	"github.com/gin-gonic/gin"
)

// Handler implementa os endpoints de agendamentos.
type Handler struct {
	*app.Handler
}

// resolveRole determina o role efetivo do usuário em relação à org: "admin", "owner" ou "user".
func (h *Handler) resolveRole(authUser middleware.AuthUser, orgSlug string) string {
	if authUser.Role == "admin" {
		return "admin"
	}
	var org models.Organization
	if err := h.DB.Where("slug = ? AND deleted_at IS NULL", orgSlug).First(&org).Error; err == nil {
		if authUser.ID == org.OwnerID {
			return "owner"
		}
	}
	return "user"
}

// Create POST /organizations/:slug/schedules
func (h *Handler) Create(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	slug := c.Param("slug")

	var input models.CreateScheduleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := h.resolveRole(authUser, slug)

	result, err := actions.Create(h.DB, slug, authUser.ID, role, input)
	if err != nil {
		switch {
		case errors.Is(err, actions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		case errors.Is(err, actions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, actions.ErrProfessionalNotMember), errors.Is(err, actions.ErrServiceNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, actions.ErrPastScheduledAt), errors.Is(err, actions.ErrUnavailable), errors.Is(err, actions.ErrConflict):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, result)
}

// List GET /organizations/:slug/schedules
func (h *Handler) List(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	slug := c.Param("slug")
	role := h.resolveRole(authUser, slug)

	filter := models.ListSchedulesFilter{}

	filter.Status = c.Query("status")
	filter.Date = c.Query("date")

	if profIDStr := c.Query("professional_id"); profIDStr != "" {
		profID, err := strconv.ParseUint(profIDStr, 10, 64)
		if err == nil {
			id := uint(profID)
			filter.ProfessionalID = &id
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = v
		}
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = v
		}
	}

	result, err := actions.List(h.DB, slug, authUser.ID, role, filter)
	if err != nil {
		switch {
		case errors.Is(err, actions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		case errors.Is(err, actions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// Find GET /organizations/:slug/schedules/:id
func (h *Handler) Find(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	slug := c.Param("slug")
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule id"})
		return
	}

	role := h.resolveRole(authUser, slug)

	result, err := actions.Find(h.DB, slug, uint(id), authUser.ID, role)
	if err != nil {
		switch {
		case errors.Is(err, actions.ErrScheduleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
		case errors.Is(err, actions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// MySchedules GET /my/schedules
func (h *Handler) MySchedules(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	filter := models.ListSchedulesFilter{}

	filter.Status = c.Query("status")
	filter.Date = c.Query("date")
	filter.OrganizationSlug = c.Query("organization_slug")

	if limitStr := c.Query("limit"); limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = v
		}
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = v
		}
	}

	result, err := actions.MySchedules(h.DB, authUser.ID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
