package scheduleshttp

import (
	"errors"
	"net/http"
	"strconv"

	"barbearia-api/app/middleware"
	"barbearia-api/app/schedules/actions"

	"github.com/gin-gonic/gin"
)

func parseScheduleID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule id"})
		return 0, false
	}
	return uint(id), true
}

func handleTransitionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, actions.ErrOrgNotFound), errors.Is(err, actions.ErrScheduleNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, actions.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, actions.ErrInvalidTransition):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// Confirm PATCH /organizations/:slug/schedules/:id/confirm
func (h *Handler) Confirm(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	slug := c.Param("slug")

	scheduleID, ok := parseScheduleID(c)
	if !ok {
		return
	}

	role := h.resolveRole(authUser, slug)

	result, err := actions.Confirm(h.DB, slug, scheduleID, authUser.ID, role)
	if err != nil {
		handleTransitionError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// Cancel PATCH /organizations/:slug/schedules/:id/cancel
func (h *Handler) Cancel(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	slug := c.Param("slug")

	scheduleID, ok := parseScheduleID(c)
	if !ok {
		return
	}

	role := h.resolveRole(authUser, slug)

	result, err := actions.Cancel(h.DB, slug, scheduleID, authUser.ID, role)
	if err != nil {
		handleTransitionError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// Complete PATCH /organizations/:slug/schedules/:id/complete
func (h *Handler) Complete(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	slug := c.Param("slug")

	scheduleID, ok := parseScheduleID(c)
	if !ok {
		return
	}

	role := h.resolveRole(authUser, slug)

	result, err := actions.Complete(h.DB, slug, scheduleID, authUser.ID, role)
	if err != nil {
		handleTransitionError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// NoShow PATCH /organizations/:slug/schedules/:id/no-show
func (h *Handler) NoShow(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	slug := c.Param("slug")

	scheduleID, ok := parseScheduleID(c)
	if !ok {
		return
	}

	role := h.resolveRole(authUser, slug)

	result, err := actions.NoShow(h.DB, slug, scheduleID, authUser.ID, role)
	if err != nil {
		handleTransitionError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
