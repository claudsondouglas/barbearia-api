package actions

import (
	"barbearia-api/models"

	"gorm.io/gorm"
)

// Find retorna o customer pelo ID dentro da organização identificada por orgSlug.
// Retorna ErrForbidden se requestingUserID não for owner da org nem tiver role "admin".
// Retorna ErrCustomerNotFound se o ID não existir ou pertencer a outra org.
func Find(db *gorm.DB, orgSlug string, customerID uint, requestingUserID uint, requestingRole string) (*models.Customer, error) {
	var org models.Organization
	if result := db.Where("slug = ?", orgSlug).Find(&org); result.Error != nil || org.ID == 0 {
		return nil, ErrOrgNotFound
	}

	if org.OwnerID != requestingUserID && requestingRole != "admin" {
		return nil, ErrForbidden
	}

	var customer models.Customer
	if result := db.Where("id = ? AND organization_id = ?", customerID, org.ID).Find(&customer); result.Error != nil || customer.ID == 0 {
		return nil, ErrCustomerNotFound
	}

	return &customer, nil
}
