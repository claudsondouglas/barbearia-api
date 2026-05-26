// Package actions implementa a lógica de negócio para o módulo de members.
package actions

import (
	"errors"

	"barbearia-api/models"

	"gorm.io/gorm"
)

var ErrOwnerCannotBeRemoved = errors.New("owner cannot be removed")
var ErrMemberHasActiveSchedules = errors.New("member has active schedules")
var ErrMemberNotFound = errors.New("member not found")

// Remove realiza soft-delete do membro da organização.
// Deleta os 7 member_business_hours e schedule_exceptions do membro.
// Retorna ErrOwnerCannotBeRemoved se targetUserID for owner da org.
// Retorna ErrMemberHasActiveSchedules se houver schedules com status pending/confirmed.
func Remove(db *gorm.DB, orgSlug string, requestingUserID, targetUserID uint) error {
	var orgs []models.Organization
	if err := db.Where("slug = ?", orgSlug).Find(&orgs).Error; err != nil {
		return ErrOrgNotFound
	}
	if len(orgs) == 0 {
		return ErrOrgNotFound
	}
	org := orgs[0]

	// Owner não pode ser removido
	if targetUserID == org.OwnerID {
		return ErrOwnerCannotBeRemoved
	}

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

	// Busca o membro ativo
	var members []models.OrgMember
	if err := db.Where("organization_id = ? AND user_id = ?", org.ID, targetUserID).Find(&members).Error; err != nil {
		return ErrMemberNotFound
	}
	if len(members) == 0 {
		return ErrMemberNotFound
	}
	member := members[0]

	// Verifica schedules ativos
	var activeSchedules []models.Schedule
	if err := db.Where("professional_id = ? AND status IN (?)", member.ID, []string{"pending", "confirmed"}).Find(&activeSchedules).Error; err == nil && len(activeSchedules) > 0 {
		return ErrMemberHasActiveSchedules
	}

	// Remove business hours
	if err := db.Where("org_member_id = ?", member.ID).Delete(&models.MemberBusinessHour{}).Error; err != nil {
		return err
	}

	// Remove schedule exceptions do membro
	if err := db.Where("organization_id = ? AND user_id = ?", org.ID, targetUserID).Delete(&models.ScheduleException{}).Error; err != nil {
		return err
	}

	// Soft delete do membro
	return db.Delete(&member).Error
}
