package actions

import (
	"barbearia-api/models"

	"gorm.io/gorm"
)

// Delete realiza o soft delete de um serviço dentro da organização identificada por orgSlug.
// Retorna ErrForbidden se requestingUserID não for owner/admin da org.
// Retorna ErrOrgNotFound se o slug não existir.
// Retorna ErrServiceNotFound se o serviço não existir ou já estiver deletado.
// Retorna ErrServiceHasActiveSchedules se existirem agendamentos com status pending ou confirmed.
func Delete(db *gorm.DB, orgSlug string, serviceID uint, requestingUserID uint) error {
	var orgs []models.Organization
	if err := db.Where("slug = ?", orgSlug).Find(&orgs).Error; err != nil {
		return ErrOrgNotFound
	}
	if len(orgs) == 0 {
		return ErrOrgNotFound
	}
	org := orgs[0]

	// Verifica permissão do solicitante
	var requestingUsers []models.User
	if err := db.Where(`"users"."id" = ?`, requestingUserID).Find(&requestingUsers).Error; err != nil {
		return ErrForbidden
	}
	if len(requestingUsers) == 0 {
		return ErrForbidden
	}
	requestingUser := requestingUsers[0]
	if org.OwnerID != requestingUserID && requestingUser.Role != "admin" {
		return ErrForbidden
	}

	// Busca serviço
	var services []models.Service
	if err := db.Where("id = ? AND organization_id = ?", serviceID, org.ID).Find(&services).Error; err != nil {
		return ErrServiceNotFound
	}
	if len(services) == 0 {
		return ErrServiceNotFound
	}
	service := services[0]

	// Verifica agendamentos ativos
	var activeSchedules []models.Schedule
	if err := db.Where("service_id = ? AND status IN (?)", service.ID, []string{"pending", "confirmed"}).Find(&activeSchedules).Error; err == nil && len(activeSchedules) > 0 {
		return ErrServiceHasActiveSchedules
	}

	// Soft delete
	return db.Delete(&service).Error
}
