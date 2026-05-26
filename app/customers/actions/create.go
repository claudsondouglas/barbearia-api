// Package actions implementa a lógica de negócio para o módulo de customers.
package actions

import (
	"errors"

	"barbearia-api/models"

	"gorm.io/gorm"
)

var ErrCustomerNotFound = errors.New("customer not found")
var ErrDuplicatePhone = errors.New("phone already registered for this org")
var ErrForbidden = errors.New("forbidden")
var ErrOrgNotFound = errors.New("organization not found")
var ErrActiveMemberAsCustomer = errors.New("active members cannot be customers")
var ErrCustomerHasActiveSchedules = errors.New("customer has active schedules")

// Create cria um novo customer manualmente para a organização identificada por orgSlug.
// Retorna ErrForbidden se requestingRole não for "owner" nem "admin".
// Retorna ErrOrgNotFound se o slug não existir.
// Retorna ErrDuplicatePhone se o phone já estiver registrado na org.
// Retorna ErrActiveMemberAsCustomer se o usuário vinculado for membro ativo da org.
// O campo user_id nunca é definido pelo chamador — é ignorado do input.
func Create(db *gorm.DB, orgSlug string, requestingUserID uint, requestingRole string, input models.CreateCustomerInput) (*models.Customer, error) {
	if requestingRole != "owner" && requestingRole != "admin" {
		return nil, ErrForbidden
	}

	var org models.Organization
	if result := db.Where("slug = ?", orgSlug).Find(&org); result.Error != nil || org.ID == 0 {
		return nil, ErrOrgNotFound
	}

	var existingCustomer models.Customer
	if result := db.Where("organization_id = ? AND phone = ?", org.ID, input.Phone).Find(&existingCustomer); result.Error == nil && existingCustomer.ID != 0 {
		return nil, ErrDuplicatePhone
	}

	var member models.OrgMember
	if result := db.Where("organization_id = ? AND deleted_at IS NULL", org.ID).Find(&member); result.Error == nil && member.ID != 0 {
		return nil, ErrActiveMemberAsCustomer
	}

	customer := models.Customer{
		OrganizationID: org.ID,
		Name:           input.Name,
		Phone:          input.Phone,
		Notes:          input.Notes,
	}

	if result := db.Create(&customer); result.Error != nil {
		return nil, result.Error
	}

	return &customer, nil
}
