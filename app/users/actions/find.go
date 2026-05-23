package actions

import (
	"barbearia-api/models"

	"gorm.io/gorm"
)

// Find busca um usuário pelo ID.
// Retorna gorm.ErrRecordNotFound se o usuário não existir.
func Find(db *gorm.DB, id string) (*models.User, error) {
	var user models.User
	if result := db.First(&user, id); result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}
