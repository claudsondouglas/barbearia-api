package actions

import (
	"barbearia-api/models"

	"gorm.io/gorm"
)

// List retorna todos os usuários cadastrados.
func List(db *gorm.DB) ([]models.User, error) {
	var users []models.User
	if result := db.Find(&users); result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}
