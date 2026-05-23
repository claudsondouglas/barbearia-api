// Package actions implementa as operações de negócio para gerenciamento de usuários.
package actions

import (
	"barbearia-api/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Create cria um novo usuário com a senha hasheada via bcrypt.
func Create(db *gorm.DB, input models.CreateUserInput) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hash),
	}

	if result := db.Create(&user); result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}
