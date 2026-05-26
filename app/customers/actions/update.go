package actions

import (
	"errors"

	"barbearia-api/models"

	"gorm.io/gorm"
)

// Update atualiza parcialmente um customer identificado por customerID dentro da org.
// Retorna ErrForbidden se requestingUserID não for owner da org nem tiver role "admin".
// Retorna ErrCustomerNotFound se o ID não existir ou pertencer a outra org.
// Retorna ErrDuplicatePhone se o novo phone já estiver em uso por outro customer da mesma org.
func Update(db *gorm.DB, orgSlug string, customerID uint, requestingUserID uint, requestingRole string, input models.UpdateCustomerInput) (*models.Customer, error) {
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

	if input.Name == nil && input.Phone == nil && input.Notes == nil {
		return nil, errors.New("no fields to update")
	}

	if input.Phone != nil {
		var existing models.Customer
		if result := db.Where("organization_id = ? AND phone = ? AND id != ?", org.ID, *input.Phone, customerID).Find(&existing); result.Error == nil && existing.ID != 0 {
			return nil, ErrDuplicatePhone
		}
	}

	updates := map[string]interface{}{}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Phone != nil {
		updates["phone"] = *input.Phone
	}
	if input.Notes != nil {
		updates["notes"] = *input.Notes
	}

	if result := db.Model(&customer).Updates(updates); result.Error != nil {
		return nil, result.Error
	}

	return &customer, nil
}
