// Package scheduleshttp implementa os handlers HTTP para as rotas de agendamentos.
package scheduleshttp

import (
	"net/http"

	"barbearia-api/app"
	"barbearia-api/app/middleware"

	"github.com/gin-gonic/gin"
)

// Handler implementa os endpoints de agendamentos.
type Handler struct {
	*app.Handler
}

// Create POST /organizations/:slug/schedules
func (h *Handler) Create(c *gin.Context) {
	_, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	panic("not implemented")
}

// List GET /organizations/:slug/schedules
func (h *Handler) List(c *gin.Context) {
	_, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	panic("not implemented")
}

// Find GET /organizations/:slug/schedules/:id
func (h *Handler) Find(c *gin.Context) { panic("not implemented") }

// MySchedules GET /my/schedules
func (h *Handler) MySchedules(c *gin.Context) {
	_, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	panic("not implemented")
}
