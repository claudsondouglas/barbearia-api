package actions

import (
	"barbearia-api/models"

	"gorm.io/gorm"
)

// Create cria um novo serviço na organização identificada por orgSlug.
// Retorna ErrForbidden se requestingUserID não for owner/admin da org.
// Retorna ErrOrgNotFound se o slug não existir.
// Retorna ErrInvalidPrice se price <= 0.
// Retorna ErrInvalidDuration se duration_min < 1 ou > 240.
// Se Active não for informado no input, o serviço é criado com active=true.
func Create(db *gorm.DB, orgSlug string, requestingUserID uint, input models.CreateServiceInput) (*models.Service, error) {
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

	// Validações
	if input.Price <= 0 {
		return nil, ErrInvalidPrice
	}
	if input.DurationMin < 1 || input.DurationMin > 240 {
		return nil, ErrInvalidDuration
	}

	active := true
	if input.Active != nil {
		active = *input.Active
	}

	service := models.Service{
		OrganizationID: org.ID,
		Name:           input.Name,
		Description:    input.Description,
		Price:          input.Price,
		DurationMin:    input.DurationMin,
		Active:         active,
	}

	if err := db.Create(&service).Error; err != nil {
		return nil, err
	}

	return &service, nil
}
