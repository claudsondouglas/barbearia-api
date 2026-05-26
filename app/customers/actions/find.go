package actions

import (
	"barbearia-api/models"

	"gorm.io/gorm"
)

// Find retorna o customer pelo ID dentro da organização identificada por orgSlug.
// Retorna ErrForbidden se requestingRole não for "owner" nem "admin".
// Retorna ErrCustomerNotFound se o ID não existir ou pertencer a outra org.
func Find(db *gorm.DB, orgSlug string, customerID uint, requestingUserID uint, requestingRole string) (*models.Customer, error) {
	if requestingRole != "owner" && requestingRole != "admin" {
		return nil, ErrForbidden
	}

	var org models.Organization
	if result := db.Where("slug = ?", orgSlug).Find(&org); result.Error != nil || org.ID == 0 {
		return nil, ErrOrgNotFound
	}

	var customer models.Customer
	if result := db.Where("id = ? AND organization_id = ?", customerID, org.ID).Find(&customer); result.Error != nil || customer.ID == 0 {
		return nil, ErrCustomerNotFound
	}

	return &customer, nil
}
