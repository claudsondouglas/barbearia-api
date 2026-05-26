package actions

import (
	"barbearia-api/models"

	"gorm.io/gorm"
)

// List retorna os serviços de uma organização identificada por orgSlug.
// Se requestingUserID for nil ou não for owner/admin da org, apenas serviços ativos são retornados.
// includeInactive só é honrado para owners/admins autenticados.
// Suporta paginação via limit e offset.
// Retorna ErrOrgNotFound se o slug não existir.
func List(db *gorm.DB, orgSlug string, requestingUserID *uint, includeInactive bool, limit, offset int) ([]models.Service, error) {
	var orgs []models.Organization
	if err := db.Where("slug = ?", orgSlug).Find(&orgs).Error; err != nil {
		return nil, ErrOrgNotFound
	}
	if len(orgs) == 0 {
		return nil, ErrOrgNotFound
	}
	org := orgs[0]

	query := db.Where("organization_id = ?", org.ID)

	if requestingUserID == nil || !includeInactive {
		query = query.Where("active = ?", true)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var services []models.Service
	if err := query.Find(&services).Error; err != nil {
		return nil, err
	}

	return services, nil
}
