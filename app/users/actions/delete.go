package actions

import (
	"barbearia-api/models"

	"gorm.io/gorm"
)

// Delete remove permanentemente o usuário do banco de dados.
func Delete(db *gorm.DB, user *models.User) error {
	if result := db.Delete(user); result.Error != nil {
		return result.Error
	}
	return nil
}
