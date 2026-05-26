package scheduleshttp

import (
	"net/http"

	"barbearia-api/app/middleware"

	"github.com/gin-gonic/gin"
)

// Confirm PATCH /organizations/:slug/schedules/:id/confirm
func (h *Handler) Confirm(c *gin.Context) {
	_, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	panic("not implemented")
}

// Cancel PATCH /organizations/:slug/schedules/:id/cancel
func (h *Handler) Cancel(c *gin.Context) {
	_, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	panic("not implemented")
}

// Complete PATCH /organizations/:slug/schedules/:id/complete
func (h *Handler) Complete(c *gin.Context) { panic("not implemented") }

// NoShow PATCH /organizations/:slug/schedules/:id/no-show
func (h *Handler) NoShow(c *gin.Context) { panic("not implemented") }
