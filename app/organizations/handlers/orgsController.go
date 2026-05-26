// Package orgshttp implementa os handlers HTTP para o módulo de organizations.
package orgshttp

import (
	"errors"
	"net/http"
	"strconv"

	"barbearia-api/app"
	"barbearia-api/app/middleware"
	"barbearia-api/app/organizations/actions"

	"github.com/gin-gonic/gin"
)

// Handler é o handler HTTP para o módulo de organizations.
// Embute app.Handler para ter acesso ao DB compartilhado.
type Handler struct {
	*app.Handler
}

// Create trata POST /organizations — cria uma nova organização.
func (h *Handler) Create(c *gin.Context) {
	authUser, _ := middleware.GetAuthUser(c)

	var input actions.CreateOrgInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	org, err := actions.Create(h.DB, authUser.ID, input)
	if err != nil {
		if errors.Is(err, actions.ErrInvalidTimezone) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, org.ToResponse())
}

// Find trata GET /organizations/:slug — retorna uma organização pelo slug.
func (h *Handler) Find(c *gin.Context) {
	slug := c.Param("slug")

	org, err := actions.FindBySlug(h.DB, slug)
	if err != nil {
		if errors.Is(err, actions.ErrOrgNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, org.ToResponse())
}

// Update trata PATCH /organizations/:slug — atualiza campos de uma organização.
func (h *Handler) Update(c *gin.Context) {
	slug := c.Param("slug")
	authUser, _ := middleware.GetAuthUser(c)

	var input actions.UpdateOrgInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	org, err := actions.Update(h.DB, slug, authUser.ID, authUser.Role, input)
	if err != nil {
		switch {
		case errors.Is(err, actions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		case errors.Is(err, actions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, actions.ErrInvalidTimezone):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, org.ToResponse())
}

// Delete trata DELETE /organizations/:slug — soft delete de uma organização.
func (h *Handler) Delete(c *gin.Context) {
	slug := c.Param("slug")
	authUser, _ := middleware.GetAuthUser(c)

	if authUser.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	if err := actions.Delete(h.DB, slug); err != nil {
		if errors.Is(err, actions.ErrOrgNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// List trata GET /organizations — lista orgs com paginação (admin only).
func (h *Handler) List(c *gin.Context) {
	limit := 20
	offset := 0
	deleted := false

	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	if v := c.Query("deleted"); v == "true" {
		deleted = true
	}

	orgs, err := actions.List(h.DB, limit, offset, deleted)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]interface{}, len(orgs))
	for i := range orgs {
		response[i] = orgs[i].ToResponse()
	}

	c.JSON(http.StatusOK, response)
}

// MyOrgs trata GET /my/organizations — lista orgs do usuário autenticado.
func (h *Handler) MyOrgs(c *gin.Context) {
	authUser, _ := middleware.GetAuthUser(c)

	orgs, err := actions.MyOrgs(h.DB, authUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]interface{}, len(orgs))
	for i := range orgs {
		response[i] = orgs[i].ToResponse()
	}

	c.JSON(http.StatusOK, response)
}

// GetBusinessHours trata GET /organizations/:slug/business-hours (público).
func (h *Handler) GetBusinessHours(c *gin.Context) {
	panic("not implemented")
}

// GetMemberBusinessHours trata GET /organizations/:slug/members/:user_id/business-hours (público).
func (h *Handler) GetMemberBusinessHours(c *gin.Context) {
	panic("not implemented")
}

// UpdateMemberBusinessHoursBatch trata PUT /organizations/:slug/members/:user_id/business-hours.
func (h *Handler) UpdateMemberBusinessHoursBatch(c *gin.Context) {
	panic("not implemented")
}

// UpdateMemberBusinessHourDay trata PUT /organizations/:slug/members/:user_id/business-hours/:day.
func (h *Handler) UpdateMemberBusinessHourDay(c *gin.Context) {
	panic("not implemented")
}

// GetAvailability trata GET /organizations/:slug/availability (público).
func (h *Handler) GetAvailability(c *gin.Context) {
	panic("not implemented")
}

// ListOrgExceptions trata GET /organizations/:slug/schedule-exceptions.
func (h *Handler) ListOrgExceptions(c *gin.Context) {
	panic("not implemented")
}

// CreateOrgException trata POST /organizations/:slug/schedule-exceptions.
func (h *Handler) CreateOrgException(c *gin.Context) {
	panic("not implemented")
}

// DeleteOrgException trata DELETE /organizations/:slug/schedule-exceptions/:id.
func (h *Handler) DeleteOrgException(c *gin.Context) {
	panic("not implemented")
}

// ListMemberExceptions trata GET /organizations/:slug/members/:user_id/schedule-exceptions.
func (h *Handler) ListMemberExceptions(c *gin.Context) {
	panic("not implemented")
}

// CreateMemberException trata POST /organizations/:slug/members/:user_id/schedule-exceptions.
func (h *Handler) CreateMemberException(c *gin.Context) {
	panic("not implemented")
}

// DeleteMemberException trata DELETE /organizations/:slug/members/:user_id/schedule-exceptions/:id.
func (h *Handler) DeleteMemberException(c *gin.Context) {
	panic("not implemented")
}
