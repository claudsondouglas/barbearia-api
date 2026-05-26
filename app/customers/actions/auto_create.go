package actions

import (
	"barbearia-api/models"

	"gorm.io/gorm"
)

// AutoCreate cria automaticamente um customer vinculado a um user durante o agendamento.
// Se já existir um customer com user_id=userID na org, retorna o existente sem criar novo.
// Se phone for vazio (""), retorna nil, nil (sem erro, sem customer criado).
// Deve ser chamado dentro de uma transação já aberta pelo módulo de schedules.
func AutoCreate(db *gorm.DB, orgID, userID uint, name, phone string) (*models.Customer, error) {
	if phone == "" {
		return nil, nil
	}

	var existing models.Customer
	if result := db.Where("organization_id = ? AND user_id = ?", orgID, userID).First(&existing); result.Error == nil {
		return &existing, nil
	}

	customer := models.Customer{
		OrganizationID: orgID,
		UserID:         &userID,
		Name:           name,
		Phone:          phone,
	}

	if result := db.Create(&customer); result.Error != nil {
		return nil, result.Error
	}

	return &customer, nil
}
