// Package usershttp implementa os handlers HTTP para as rotas de gerenciamento de usuários.
package usershttp

import (
	"errors"
	"net/http"

	"barbearia-api/app"
	"barbearia-api/app/users/actions"
	"barbearia-api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler implementa os endpoints de CRUD de usuários.
type Handler struct {
	*app.Handler
}

// List retorna todos os usuários cadastrados.
func (h *Handler) List(c *gin.Context) {
	users, err := actions.List(h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]gin.H, len(users))
	for i, u := range users {
		response[i] = gin.H{
			"id":         u.ID,
			"name":       u.Name,
			"email":      u.Email,
			"created_at": u.CreatedAt,
			"updated_at": u.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, response)
}

// Find retorna um usuário pelo ID. Responde 404 se não encontrado.
func (h *Handler) Find(c *gin.Context) {
	id := c.Param("id")

	user, err := actions.Find(h.DB, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Create cria um novo usuário com a senha hasheada. Responde 201 em caso de sucesso.
func (h *Handler) Create(c *gin.Context) {
	var input models.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := actions.Create(h.DB, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Update aplica as alterações parciais ao usuário identificado pelo ID.
// Responde 404 se não encontrado ou 400 se nenhum campo for informado.
func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")

	var input models.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := actions.Update(h.DB, id, input)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		case errors.Is(err, actions.ErrNoFieldsToUpdate):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// Delete remove o usuário identificado pelo ID. Responde 204 em caso de sucesso.
func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")

	user, err := actions.Find(h.DB, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := actions.Delete(h.DB, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
