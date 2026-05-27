// Package orgshttp implementa os handlers HTTP para o módulo de organizations.
package orgshttp

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"barbearia-api/app"
	"barbearia-api/app/middleware"
	"barbearia-api/app/organizations/actions"
	"barbearia-api/models"

	"github.com/gin-gonic/gin"
)

func parseUserID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return 0, false
	}
	return uint(id), true
}

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
	slug := c.Param("slug")

	hours, err := actions.GetBusinessHours(h.DB, slug)
	if err != nil {
		if errors.Is(err, actions.ErrOrgNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, hours)
}

// GetMemberBusinessHours trata GET /organizations/:slug/members/:user_id/business-hours (público).
func (h *Handler) GetMemberBusinessHours(c *gin.Context) {
	slug := c.Param("slug")
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	hours, err := actions.GetMemberBusinessHours(h.DB, slug, userID)
	if err != nil {
		switch {
		case errors.Is(err, actions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		case errors.Is(err, actions.ErrMemberNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, hours)
}

// UpdateMemberBusinessHoursBatch trata PUT /organizations/:slug/members/:user_id/business-hours.
func (h *Handler) UpdateMemberBusinessHoursBatch(c *gin.Context) {
	slug := c.Param("slug")
	authUser, _ := middleware.GetAuthUser(c)
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	var updates []actions.UpdateBusinessHourInput
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := actions.UpdateMemberBusinessHoursBatch(h.DB, slug, authUser.ID, userID, updates); err != nil {
		switch {
		case errors.Is(err, actions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		case errors.Is(err, actions.ErrMemberNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		case errors.Is(err, actions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, actions.ErrInvalidBusinessHour):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateMemberBusinessHourDay trata PUT /organizations/:slug/members/:user_id/business-hours/:day.
func (h *Handler) UpdateMemberBusinessHourDay(c *gin.Context) {
	slug := c.Param("slug")
	authUser, _ := middleware.GetAuthUser(c)
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	dayParam := c.Param("day")
	day, err := strconv.Atoi(dayParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid day"})
		return
	}

	var input actions.UpdateBusinessHourInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := actions.UpdateMemberBusinessHourDay(h.DB, slug, authUser.ID, userID, day, input); err != nil {
		switch {
		case errors.Is(err, actions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		case errors.Is(err, actions.ErrMemberNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		case errors.Is(err, actions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, actions.ErrInvalidBusinessHour):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// GetAvailability trata GET /organizations/:slug/availability (público).
func (h *Handler) GetAvailability(c *gin.Context) {
	slug := c.Param("slug")

	professionalIDRaw, err := strconv.ParseUint(c.Query("professional_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid professional_id"})
		return
	}
	serviceIDRaw, err := strconv.ParseUint(c.Query("service_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service_id"})
		return
	}
	date := c.Query("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date is required"})
		return
	}

	result, err := actions.GetAvailability(h.DB, slug, uint(professionalIDRaw), uint(serviceIDRaw), date, time.Now)
	if err != nil {
		switch {
		case errors.Is(err, actions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		case errors.Is(err, actions.ErrProfessionalNotMember):
			c.JSON(http.StatusNotFound, gin.H{"error": "professional not found"})
		case errors.Is(err, actions.ErrServiceNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		case errors.Is(err, actions.ErrDateInPast), errors.Is(err, actions.ErrDateBeyond90Days):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListOrgExceptions trata GET /organizations/:slug/schedule-exceptions.
func (h *Handler) ListOrgExceptions(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	slug := c.Param("slug")
	limit := 20
	offset := 0

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

	var from, to *string
	if v := c.Query("from"); v != "" {
		from = &v
	}
	if v := c.Query("to"); v != "" {
		to = &v
	}

	exceptions, err := actions.ListOrgExceptions(h.DB, slug, authUser.ID, limit, offset, from, to)
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

	c.JSON(http.StatusOK, exceptions)
}

// CreateOrgException trata POST /organizations/:slug/schedule-exceptions.
func (h *Handler) CreateOrgException(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	slug := c.Param("slug")

	var input models.CreateExceptionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	exception, deletedCount, err := actions.CreateOrgException(h.DB, slug, authUser.ID, input)
	if err != nil {
		switch {
		case errors.Is(err, actions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		case errors.Is(err, actions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, actions.ErrDuplicateException):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case errors.Is(err, actions.ErrInvalidBusinessHour):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":                               exception.ID,
		"organization_id":                  exception.OrganizationID,
		"user_id":                          exception.UserID,
		"date":                             exception.Date,
		"closed":                           exception.Closed,
		"open_time":                        exception.OpenTime,
		"close_time":                       exception.CloseTime,
		"reason":                           exception.Reason,
		"created_at":                       exception.CreatedAt,
		"updated_at":                       exception.UpdatedAt,
		"deleted_member_exceptions_count":  deletedCount,
	})
}

// DeleteOrgException trata DELETE /organizations/:slug/schedule-exceptions/:id.
func (h *Handler) DeleteOrgException(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	slug := c.Param("slug")

	idRaw, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := actions.DeleteOrgException(h.DB, slug, authUser.ID, uint(idRaw)); err != nil {
		switch {
		case errors.Is(err, actions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		case errors.Is(err, actions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, actions.ErrExceptionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "exception not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// ListMemberExceptions trata GET /organizations/:slug/members/:user_id/schedule-exceptions.
func (h *Handler) ListMemberExceptions(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	slug := c.Param("slug")
	targetUserID, ok := parseUserID(c)
	if !ok {
		return
	}

	limit := 20
	offset := 0

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

	var from, to *string
	if v := c.Query("from"); v != "" {
		from = &v
	}
	if v := c.Query("to"); v != "" {
		to = &v
	}

	exceptions, err := actions.ListMemberExceptions(h.DB, slug, authUser.ID, targetUserID, limit, offset, from, to)
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

	c.JSON(http.StatusOK, exceptions)
}

// CreateMemberException trata POST /organizations/:slug/members/:user_id/schedule-exceptions.
func (h *Handler) CreateMemberException(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	slug := c.Param("slug")
	targetUserID, ok := parseUserID(c)
	if !ok {
		return
	}

	var input models.CreateExceptionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	exception, err := actions.CreateMemberException(h.DB, slug, authUser.ID, targetUserID, input)
	if err != nil {
		switch {
		case errors.Is(err, actions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		case errors.Is(err, actions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, actions.ErrDuplicateException):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case errors.Is(err, actions.ErrOrgClosedOnDate):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case errors.Is(err, actions.ErrMemberWindowExceedsOrg):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case errors.Is(err, actions.ErrInvalidBusinessHour):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, exception)
}

// DeleteMemberException trata DELETE /organizations/:slug/members/:user_id/schedule-exceptions/:id.
func (h *Handler) DeleteMemberException(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	slug := c.Param("slug")
	targetUserID, ok := parseUserID(c)
	if !ok {
		return
	}

	idRaw, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := actions.DeleteMemberException(h.DB, slug, authUser.ID, targetUserID, uint(idRaw)); err != nil {
		switch {
		case errors.Is(err, actions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		case errors.Is(err, actions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, actions.ErrExceptionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "exception not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
