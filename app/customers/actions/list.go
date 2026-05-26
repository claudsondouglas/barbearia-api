package actions

import (
	"barbearia-api/models"

	"gorm.io/gorm"
)

// List retorna os customers da organização identificada por orgSlug.
// Suporta busca por texto (name ou phone) via filter.Query e paginação via Limit/Offset.
// Retorna ErrForbidden se requestingUserID não for owner da org nem tiver role "admin".
// Retorna ErrOrgNotFound se o slug não existir.
func List(db *gorm.DB, orgSlug string, requestingUserID uint, requestingRole string, filter models.ListCustomersFilter) ([]models.Customer, error) {
	var org models.Organization
	if result := db.Where("slug = ?", orgSlug).Find(&org); result.Error != nil || org.ID == 0 {
		return nil, ErrOrgNotFound
	}

	if org.OwnerID != requestingUserID && requestingRole != "admin" {
		return nil, ErrForbidden
	}

	query := db.Where("organization_id = ?", org.ID)

	if filter.Query != "" {
		search := "%" + filter.Query + "%"
		query = query.Where("name ILIKE ? OR phone ILIKE ?", search, search)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var customers []models.Customer
	if result := query.Find(&customers); result.Error != nil {
		return nil, result.Error
	}

	return customers, nil
}
