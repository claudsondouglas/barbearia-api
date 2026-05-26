// Package customershttp implementa os handlers HTTP para as rotas de gerenciamento de customers.
package customershttp

import (
	"errors"
	"net/http"
	"strconv"

	"barbearia-api/app"
	customersactions "barbearia-api/app/customers/actions"
	"barbearia-api/app/middleware"
	"barbearia-api/models"

	"github.com/gin-gonic/gin"
)

// Handler implementa os endpoints de gerenciamento de customers de uma organização.
type Handler struct {
	*app.Handler
}

// Create POST /organizations/:slug/customers
func (h *Handler) Create(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input models.CreateCustomerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slug := c.Param("slug")
	customer, err := customersactions.Create(h.DB, slug, authUser.ID, authUser.Role, input)
	if err != nil {
		switch {
		case errors.Is(err, customersactions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, customersactions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, customersactions.ErrDuplicatePhone):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, customersactions.ErrActiveMemberAsCustomer):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, customer)
}

// List GET /organizations/:slug/customers
func (h *Handler) List(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	slug := c.Param("slug")
	filter := models.ListCustomersFilter{
		Query: c.Query("q"),
	}

	customers, err := customersactions.List(h.DB, slug, authUser.ID, authUser.Role, filter)
	if err != nil {
		switch {
		case errors.Is(err, customersactions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, customersactions.ErrOrgNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, customers)
}

// Find GET /organizations/:slug/customers/:id
func (h *Handler) Find(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	slug := c.Param("slug")
	customer, err := customersactions.Find(h.DB, slug, uint(id), authUser.ID, authUser.Role)
	if err != nil {
		switch {
		case errors.Is(err, customersactions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, customersactions.ErrCustomerNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, customer)
}

// Update PATCH /organizations/:slug/customers/:id
func (h *Handler) Update(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var input models.UpdateCustomerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slug := c.Param("slug")
	customer, err := customersactions.Update(h.DB, slug, uint(id), authUser.ID, authUser.Role, input)
	if err != nil {
		switch {
		case errors.Is(err, customersactions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, customersactions.ErrCustomerNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, customersactions.ErrDuplicatePhone):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, customer)
}

// Delete DELETE /organizations/:slug/customers/:id
func (h *Handler) Delete(c *gin.Context) {
	authUser, ok := middleware.GetAuthUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	slug := c.Param("slug")
	if err := customersactions.Delete(h.DB, slug, uint(id), authUser.ID, authUser.Role); err != nil {
		switch {
		case errors.Is(err, customersactions.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, customersactions.ErrCustomerNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, customersactions.ErrCustomerHasActiveSchedules):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
