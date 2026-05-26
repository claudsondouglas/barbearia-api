// Package actions implementa a lógica de negócio para o módulo de members.
package actions

import (
	"barbearia-api/models"

	"gorm.io/gorm"
)

// List retorna todos os membros ativos (deleted_at IS NULL) da organização.
// Retorna ErrOrgNotFound se o slug não existir.
func List(db *gorm.DB, orgSlug string) ([]models.MemberResponse, error) {
	var orgs []models.Organization
	if err := db.Where("slug = ?", orgSlug).Find(&orgs).Error; err != nil {
		return nil, ErrOrgNotFound
	}
	if len(orgs) == 0 {
		return nil, ErrOrgNotFound
	}
	org := orgs[0]

	var members []models.MemberResponse
	result := db.Table("org_members").
		Select("org_members.id, org_members.user_id, users.name, users.email, org_members.created_at").
		Joins("JOIN users ON users.id = org_members.user_id").
		Where("org_members.organization_id = ? AND org_members.deleted_at IS NULL", org.ID).
		Scan(&members)

	if result.Error != nil {
		return nil, result.Error
	}

	if members == nil {
		members = []models.MemberResponse{}
	}

	return members, nil
}
