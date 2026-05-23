package actions

import (
	"errors"

	"barbearia-api/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ErrNoFieldsToUpdate é retornado quando a requisição de update não contém nenhum campo a alterar.
var ErrNoFieldsToUpdate = errors.New("no fields to update")

// Update aplica as alterações informadas ao usuário identificado pelo ID.
// Retorna ErrNoFieldsToUpdate se o input não contiver nenhum campo,
// ou gorm.ErrRecordNotFound se o usuário não existir.
func Update(db *gorm.DB, id string, input models.UpdateUserInput) (*models.User, error) {
	user, err := Find(db, id)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}

	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Email != nil {
		updates["email"] = *input.Email
	}
	if input.Password != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		updates["password"] = string(hash)
	}

	if len(updates) == 0 {
		return nil, ErrNoFieldsToUpdate
	}

	if result := db.Model(user).Updates(updates); result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}
