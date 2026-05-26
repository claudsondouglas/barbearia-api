// Package serviceshttp implementa os handlers HTTP para as rotas de gerenciamento de serviços.
package serviceshttp

import (
	"errors"
	"net/http"
	"strconv"

	"barbearia-api/app"
	servicesactions "barbearia-api/app/services/actions"
	"barbearia-api/app/middleware"
	"barbearia-api/models"

	"github.com/gin-gonic/gin"
)

// Handler implementa os endpoints de gerenciamento de serviços de uma organização.
type Handler struct {
	*app.Handler
}

// List GET /organizations/:slug/services
func (h *Handler) List(c *gin.Context) {
	slug := c.Param("slug")

	var requestingUserID *uint
	if authUser, ok := middleware.GetAuthUser(c); ok {
		id := authUser.ID
		requestingUserID = &id
	}

	includeInactive := c.Query("include_inactive") == "true"

	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil {
			offset = v
		}
	}

	services, err := servicesactions.List(h.DB, slug, requestingUserID, includeInactive, limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, servicesactions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, services)
}

// Find GET /organizations/:slug/services/:id
func (h *Handler) Find(c *gin.Context) {
	slug := c.Param("slug")

	serviceIDStr := c.Param("id")
	serviceID, err := strconv.ParseUint(serviceIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service id"})
		return
	}

	var requestingUserID *uint
	if authUser, ok := middleware.GetAuthUser(c); ok {
		id := authUser.ID
		requestingUserID = &id
	}

	service, err := servicesactions.Find(h.DB, slug, uint(serviceID), requestingUserID)
	if err != nil {
		switch {
		case errors.Is(err, servicesactions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, servicesactions.ErrServiceNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, service)
}

// Create POST /organizations/:slug/services
func (h *Handler) Create(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input models.CreateServiceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slug := c.Param("slug")
	service, err := servicesactions.Create(h.DB, slug, authUser.ID, input)
	if err != nil {
		switch {
		case errors.Is(err, servicesactions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, servicesactions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, servicesactions.ErrInvalidPrice):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case errors.Is(err, servicesactions.ErrInvalidDuration):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, service)
}

// Update PATCH /organizations/:slug/services/:id
func (h *Handler) Update(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	serviceIDStr := c.Param("id")
	serviceID, err := strconv.ParseUint(serviceIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service id"})
		return
	}

	var input models.UpdateServiceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slug := c.Param("slug")
	service, err := servicesactions.Update(h.DB, slug, uint(serviceID), authUser.ID, input)
	if err != nil {
		switch {
		case errors.Is(err, servicesactions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, servicesactions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, servicesactions.ErrServiceNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, servicesactions.ErrServiceDeleted):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case errors.Is(err, servicesactions.ErrInvalidPrice):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case errors.Is(err, servicesactions.ErrInvalidDuration):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, service)
}

// Delete DELETE /organizations/:slug/services/:id
func (h *Handler) Delete(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	serviceIDStr := c.Param("id")
	serviceID, err := strconv.ParseUint(serviceIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service id"})
		return
	}

	slug := c.Param("slug")
	if err := servicesactions.Delete(h.DB, slug, uint(serviceID), authUser.ID); err != nil {
		switch {
		case errors.Is(err, servicesactions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, servicesactions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, servicesactions.ErrServiceNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, servicesactions.ErrServiceHasActiveSchedules):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
