package actions

import (
	"barbearia-api/models"

	"gorm.io/gorm"
)

// Update aplica uma atualização parcial em um serviço existente.
// Apenas os campos não-nulos do input são alterados.
// Retorna ErrForbidden se requestingUserID não for owner/admin da org.
// Retorna ErrOrgNotFound se o slug não existir.
// Retorna ErrServiceNotFound se o serviço não existir na org.
// Retorna ErrServiceDeleted se o serviço estiver soft-deleted.
// Retorna ErrInvalidPrice se price <= 0.
// Retorna ErrInvalidDuration se duration_min < 1 ou > 240.
func Update(db *gorm.DB, orgSlug string, serviceID uint, requestingUserID uint, input models.UpdateServiceInput) (*models.Service, error) {
	var orgs []models.Organization
	if err := db.Where("slug = ?", orgSlug).Find(&orgs).Error; err != nil {
		return nil, ErrOrgNotFound
	}
	if len(orgs) == 0 {
		return nil, ErrOrgNotFound
	}
	org := orgs[0]

	// Verifica permissão do solicitante
	var requestingUsers []models.User
	if err := db.Where(`"users"."id" = ?`, requestingUserID).Find(&requestingUsers).Error; err != nil {
		return nil, ErrForbidden
	}
	if len(requestingUsers) == 0 {
		return nil, ErrForbidden
	}
	requestingUser := requestingUsers[0]
	if org.OwnerID != requestingUserID && requestingUser.Role != "admin" {
		return nil, ErrForbidden
	}

	// Busca serviço (incluindo soft-deleted para detectar ErrServiceDeleted)
	var services []models.Service
	if err := db.Unscoped().Where("id = ? AND organization_id = ?", serviceID, org.ID).Find(&services).Error; err != nil {
		return nil, ErrServiceNotFound
	}
	if len(services) == 0 {
		return nil, ErrServiceNotFound
	}
	service := services[0]

	if service.DeletedAt.Valid {
		return nil, ErrServiceDeleted
	}

	// Validações
	if input.Price != nil && *input.Price <= 0 {
		return nil, ErrInvalidPrice
	}
	if input.DurationMin != nil && (*input.DurationMin < 1 || *input.DurationMin > 240) {
		return nil, ErrInvalidDuration
	}

	// Aplica atualizações
	updates := map[string]interface{}{}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Price != nil {
		updates["price"] = *input.Price
	}
	if input.DurationMin != nil {
		updates["duration_min"] = *input.DurationMin
	}
	if input.Active != nil {
		updates["active"] = *input.Active
	}

	if len(updates) > 0 {
		if err := db.Model(&service).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return &service, nil
}
