// Package membershttp implementa os handlers HTTP para as rotas de gerenciamento de membros.
package membershttp

import (
	"errors"
	"net/http"
	"strconv"

	"barbearia-api/app"
	membersactions "barbearia-api/app/members/actions"
	"barbearia-api/app/middleware"
	"barbearia-api/models"

	"github.com/gin-gonic/gin"
)

// Handler implementa os endpoints de gerenciamento de membros de uma organização.
type Handler struct {
	*app.Handler
}

// Add POST /organizations/:slug/members
func (h *Handler) Add(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input models.AddMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slug := c.Param("slug")
	err := membersactions.Add(h.DB, slug, authUser.ID, input.UserID)
	if err != nil {
		switch {
		case errors.Is(err, membersactions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, membersactions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, membersactions.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, membersactions.ErrAlreadyActiveMember):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusCreated)
}

// Remove DELETE /organizations/:slug/members/:user_id
func (h *Handler) Remove(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDStr := c.Param("user_id")
	targetUserID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	slug := c.Param("slug")
	if err := membersactions.Remove(h.DB, slug, authUser.ID, uint(targetUserID)); err != nil {
		switch {
		case errors.Is(err, membersactions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, membersactions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, membersactions.ErrMemberNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, membersactions.ErrOwnerCannotBeRemoved):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case errors.Is(err, membersactions.ErrMemberHasActiveSchedules):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// List GET /organizations/:slug/members
func (h *Handler) List(c *gin.Context) {
	slug := c.Param("slug")
	members, err := membersactions.List(h.DB, slug)
	if err != nil {
		switch {
		case errors.Is(err, membersactions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, members)
}
