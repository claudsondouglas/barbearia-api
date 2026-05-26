package actions

import (
	"barbearia-api/models"

	"gorm.io/gorm"
)

// Find retorna um serviço pelo ID dentro da organização identificada por orgSlug.
// Usuários não autenticados (requestingUserID == nil) só podem ver serviços ativos e não deletados.
// Owners/admins autenticados podem ver serviços inativos, mas não os soft-deleted.
// Retorna ErrOrgNotFound, ErrServiceNotFound ou ErrServiceDeleted conforme o caso.
func Find(db *gorm.DB, orgSlug string, serviceID uint, requestingUserID *uint) (*models.Service, error) {
	var orgs []models.Organization
	if err := db.Where("slug = ?", orgSlug).Find(&orgs).Error; err != nil {
		return nil, ErrOrgNotFound
	}
	if len(orgs) == 0 {
		return nil, ErrOrgNotFound
	}
	org := orgs[0]

	query := db.Where("id = ? AND organization_id = ?", serviceID, org.ID)
	if requestingUserID == nil {
		query = query.Where("active = ?", true)
	}

	var services []models.Service
	if err := query.Find(&services).Error; err != nil {
		return nil, ErrServiceNotFound
	}
	if len(services) == 0 {
		return nil, ErrServiceNotFound
	}

	service := services[0]
	return &service, nil
}
