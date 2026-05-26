// Package actions implementa a lógica de negócio para o módulo de members.
package actions

import (
	"errors"

	"barbearia-api/models"

	"gorm.io/gorm"
)

var ErrAlreadyActiveMember = errors.New("user is already an active member")
var ErrUserNotFound = errors.New("user not found")
var ErrOrgNotFound = errors.New("organization not found")
var ErrForbidden = errors.New("forbidden")

// Add adiciona um usuário como membro da organização identificada por slug.
// Reativa soft-deleted members se já existirem. Cria 7 member_business_hours com closed=true.
// Retorna ErrForbidden se requestingUserID não for owner/admin da org.
// Retorna ErrOrgNotFound, ErrUserNotFound, ErrAlreadyActiveMember conforme o caso.
func Add(db *gorm.DB, orgSlug string, requestingUserID, targetUserID uint) error {
	var orgs []models.Organization
	if err := db.Where("slug = ?", orgSlug).Find(&orgs).Error; err != nil {
		return ErrOrgNotFound
	}
	if len(orgs) == 0 {
		return ErrOrgNotFound
	}
	org := orgs[0]

	// Verifica se o solicitante tem permissão (owner ou admin do sistema)
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

	// Busca o usuário alvo
	var targetUsers []models.User
	if err := db.Where(`"users"."id" = ?`, targetUserID).Find(&targetUsers).Error; err != nil {
		return ErrUserNotFound
	}
	if len(targetUsers) == 0 {
		return ErrUserNotFound
	}

	// Verifica se já existe membro (incluindo soft-deleted)
	var existingMembers []models.OrgMember
	if err := db.Unscoped().Where("organization_id = ? AND user_id = ?", org.ID, targetUserID).Find(&existingMembers).Error; err != nil {
		return err
	}

	var existing models.OrgMember
	if len(existingMembers) > 0 {
		existing = existingMembers[0]
		if existing.DeletedAt.Valid {
			// Reativa membro soft-deleted
			if err := db.Unscoped().Model(&existing).Update("deleted_at", nil).Error; err != nil {
				return err
			}
			// Remove business hours antigos
			if err := db.Where("org_member_id = ?", existing.ID).Delete(&models.MemberBusinessHour{}).Error; err != nil {
				return err
			}
		} else {
			return ErrAlreadyActiveMember
		}
	} else {
		// Cria novo membro
		member := models.OrgMember{
			OrganizationID: org.ID,
			UserID:         targetUserID,
		}
		if err := db.Create(&member).Error; err != nil {
			return err
		}
		existing = member
	}

	// Cria 7 business hours com closed=true
	for i := 0; i < 7; i++ {
		hour := models.MemberBusinessHour{
			OrgMemberID: existing.ID,
			DayOfWeek:   i,
			Closed:      true,
		}
		if err := db.Create(&hour).Error; err != nil {
			return err
		}
	}

	return nil
}
