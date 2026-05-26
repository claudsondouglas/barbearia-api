package scheduleshttp

import (
	"net/http"

	"barbearia-api/app/middleware"

	"github.com/gin-gonic/gin"
)

// Reschedule PATCH /organizations/:slug/schedules/:id/reschedule
func (h *Handler) Reschedule(c *gin.Context) {
	_, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	panic("not implemented")
}

// RescheduleHistory GET /organizations/:slug/schedules/:id/reschedule-history
func (h *Handler) RescheduleHistory(c *gin.Context) {
	_, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	panic("not implemented")
}
