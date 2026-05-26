// Package orgshttp implementa os handlers HTTP para o módulo de organizations.
package orgshttp

import (
	"barbearia-api/app"

	"github.com/gin-gonic/gin"
)

// Handler é o handler HTTP para o módulo de organizations.
// Embute app.Handler para ter acesso ao DB compartilhado.
type Handler struct {
	*app.Handler
}

// Create trata POST /organizations — cria uma nova organização.
func (h *Handler) Create(c *gin.Context) {
	panic("not implemented")
}

// Find trata GET /organizations/:slug — retorna uma organização pelo slug.
func (h *Handler) Find(c *gin.Context) {
	panic("not implemented")
}

// Update trata PATCH /organizations/:slug — atualiza campos de uma organização.
func (h *Handler) Update(c *gin.Context) {
	panic("not implemented")
}

// Delete trata DELETE /organizations/:slug — soft delete de uma organização.
func (h *Handler) Delete(c *gin.Context) {
	panic("not implemented")
}

// List trata GET /organizations — lista orgs com paginação (admin only).
func (h *Handler) List(c *gin.Context) {
	panic("not implemented")
}

// MyOrgs trata GET /my/organizations — lista orgs do usuário autenticado.
func (h *Handler) MyOrgs(c *gin.Context) {
	panic("not implemented")
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
