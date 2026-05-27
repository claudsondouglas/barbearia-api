package scheduleshttp

import (
	"errors"
	"net/http"
	"strconv"

	"barbearia-api/app/middleware"
	"barbearia-api/app/schedules/actions"

	"github.com/gin-gonic/gin"
)

// Reschedule PATCH /organizations/:slug/schedules/:id/reschedule
func (h *Handler) Reschedule(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	slug := c.Param("slug")

	idStr := c.Param("id")
	scheduleID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule id"})
		return
	}

	var input actions.RescheduleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := h.resolveRole(authUser, slug)

	result, err := actions.Reschedule(h.DB, slug, uint(scheduleID), authUser.ID, role, input)
	if err != nil {
		switch {
		case errors.Is(err, actions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		case errors.Is(err, actions.ErrScheduleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
		case errors.Is(err, actions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, actions.ErrInvalidTransition):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case errors.Is(err, actions.ErrPastScheduledAt):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case errors.Is(err, actions.ErrConflict):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case errors.Is(err, actions.ErrUnavailable):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// RescheduleHistory GET /organizations/:slug/schedules/:id/reschedule-history
func (h *Handler) RescheduleHistory(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	slug := c.Param("slug")

	idStr := c.Param("id")
	scheduleID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule id"})
		return
	}

	role := h.resolveRole(authUser, slug)

	history, err := actions.GetRescheduleHistory(h.DB, slug, uint(scheduleID), authUser.ID, role)
	if err != nil {
		switch {
		case errors.Is(err, actions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		case errors.Is(err, actions.ErrScheduleNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
		case errors.Is(err, actions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, history)
}
