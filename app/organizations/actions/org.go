// Package actions implementa a lógica de negócio para o módulo de organizations.
package actions

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"barbearia-api/models"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"gorm.io/gorm"
)

var ErrOrgNotFound = errors.New("organization not found")
var ErrSlugTaken = errors.New("slug already taken")
var ErrInvalidTimezone = errors.New("invalid timezone")
var ErrForbidden = errors.New("forbidden")

// CreateOrgInput contém os campos para criar uma organização.
type CreateOrgInput struct {
	Name         string   `json:"name" binding:"required"`
	Phone        string   `json:"phone" binding:"required"`
	Email        string   `json:"email" binding:"required,email"`
	Street       string   `json:"street" binding:"required"`
	Number       string   `json:"number" binding:"required"`
	Neighborhood string   `json:"neighborhood" binding:"required"`
	City         string   `json:"city" binding:"required"`
	State        string   `json:"state" binding:"required"`
	ZipCode      string   `json:"zip_code" binding:"required"`
	Timezone     string   `json:"timezone"`
	Description  *string  `json:"description"`
	LogoURL      *string  `json:"logo_url"`
	Complement   *string  `json:"complement"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
}

// UpdateOrgInput contém os campos opcionais para atualizar uma organização.
type UpdateOrgInput struct {
	Name         *string  `json:"name"`
	Phone        *string  `json:"phone"`
	Email        *string  `json:"email"`
	Street       *string  `json:"street"`
	Number       *string  `json:"number"`
	Neighborhood *string  `json:"neighborhood"`
	City         *string  `json:"city"`
	State        *string  `json:"state"`
	ZipCode      *string  `json:"zip_code"`
	Timezone     *string  `json:"timezone"`
	Description  *string  `json:"description"`
	LogoURL      *string  `json:"logo_url"`
	Complement   *string  `json:"complement"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
}

// isMn reports whether r is a non-spacing mark (used to strip accents).
func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}

// GenerateSlug gera um slug URL-friendly a partir de um nome:
// converte para minúsculas, remove acentos, substitui espaços e caracteres não alfanuméricos por hífens.
func GenerateSlug(name string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	result, _, _ := transform.String(t, name)

	result = strings.ToLower(result)

	re := regexp.MustCompile(`[^a-z0-9]+`)
	result = re.ReplaceAllString(result, "-")

	result = strings.Trim(result, "-")

	return result
}

// generateUniqueSlug gera um slug único com sufixo incremental em caso de colisão.
func generateUniqueSlug(db *gorm.DB, name string) (string, error) {
	base := GenerateSlug(name)
	slug := base
	for i := 2; ; i++ {
		var count int64
		if err := db.Model(&models.Organization{}).Unscoped().Where("slug = ?", slug).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
}

// Create cria uma nova organização com o owner indicado.
// Executa em transação: cria org + OrgMember + 7 MemberBusinessHours.
// Valida timezone IANA; gera slug único a partir do nome.
func Create(db *gorm.DB, ownerID uint, input CreateOrgInput) (*models.Organization, error) {
	tz := input.Timezone
	if tz == "" {
		tz = "America/Sao_Paulo"
	}
	if _, err := time.LoadLocation(tz); err != nil {
		return nil, ErrInvalidTimezone
	}

	slug, err := generateUniqueSlug(db, input.Name)
	if err != nil {
		return nil, err
	}

	var org models.Organization
	err = db.Transaction(func(tx *gorm.DB) error {
		org = models.Organization{
			OwnerID:      ownerID,
			Name:         input.Name,
			Slug:         slug,
			Phone:        input.Phone,
			Email:        input.Email,
			Description:  input.Description,
			LogoURL:      input.LogoURL,
			Street:       input.Street,
			Number:       input.Number,
			Complement:   input.Complement,
			Neighborhood: input.Neighborhood,
			City:         input.City,
			State:        input.State,
			ZipCode:      input.ZipCode,
			Latitude:     input.Latitude,
			Longitude:    input.Longitude,
			Timezone:     tz,
		}
		if err := tx.Create(&org).Error; err != nil {
			return err
		}

		member := models.OrgMember{
			OrganizationID: org.ID,
			UserID:         ownerID,
		}
		if err := tx.Create(&member).Error; err != nil {
			return err
		}

		for day := 0; day < 7; day++ {
			hour := models.MemberBusinessHour{
				OrgMemberID: member.ID,
				DayOfWeek:   day,
				Closed:      true,
			}
			if err := tx.Create(&hour).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &org, nil
}

// FindBySlug busca uma organização ativa pelo slug.
// Retorna ErrOrgNotFound se não existir ou estiver soft-deleted.
func FindBySlug(db *gorm.DB, slug string) (*models.Organization, error) {
	var org models.Organization
	if err := db.Where("slug = ?", slug).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return &org, nil
}

// Update aplica campos opcionais de UpdateOrgInput na organização identificada por slug.
// Apenas owner ou admin (requestingRole) pode atualizar.
// Se o nome mudar, regenera o slug.
// Retorna ErrForbidden, ErrOrgNotFound ou ErrInvalidTimezone conforme o caso.
func Update(db *gorm.DB, slug string, requestingUserID uint, requestingRole string, input UpdateOrgInput) (*models.Organization, error) {
	org, err := FindBySlug(db, slug)
	if err != nil {
		return nil, err
	}

	if org.OwnerID != requestingUserID && requestingRole != "admin" {
		return nil, ErrForbidden
	}

	if input.Timezone != nil {
		if _, err := time.LoadLocation(*input.Timezone); err != nil {
			return nil, ErrInvalidTimezone
		}
	}

	updates := map[string]interface{}{}

	if input.Name != nil {
		newSlug, err := generateUniqueSlug(db, *input.Name)
		if err != nil {
			return nil, err
		}
		org.Name = *input.Name
		org.Slug = newSlug
		updates["name"] = *input.Name
		updates["slug"] = newSlug
	}
	if input.Phone != nil {
		org.Phone = *input.Phone
		updates["phone"] = *input.Phone
	}
	if input.Email != nil {
		org.Email = *input.Email
		updates["email"] = *input.Email
	}
	if input.Street != nil {
		org.Street = *input.Street
		updates["street"] = *input.Street
	}
	if input.Number != nil {
		org.Number = *input.Number
		updates["number"] = *input.Number
	}
	if input.Neighborhood != nil {
		org.Neighborhood = *input.Neighborhood
		updates["neighborhood"] = *input.Neighborhood
	}
	if input.City != nil {
		org.City = *input.City
		updates["city"] = *input.City
	}
	if input.State != nil {
		org.State = *input.State
		updates["state"] = *input.State
	}
	if input.ZipCode != nil {
		org.ZipCode = *input.ZipCode
		updates["zip_code"] = *input.ZipCode
	}
	if input.Timezone != nil {
		org.Timezone = *input.Timezone
		updates["timezone"] = *input.Timezone
	}
	if input.Description != nil {
		org.Description = input.Description
		updates["description"] = input.Description
	}
	if input.LogoURL != nil {
		org.LogoURL = input.LogoURL
		updates["logo_url"] = input.LogoURL
	}
	if input.Complement != nil {
		org.Complement = input.Complement
		updates["complement"] = input.Complement
	}
	if input.Latitude != nil {
		org.Latitude = input.Latitude
		updates["latitude"] = input.Latitude
	}
	if input.Longitude != nil {
		org.Longitude = input.Longitude
		updates["longitude"] = input.Longitude
	}

	if len(updates) == 0 {
		return org, nil
	}

	if err := db.Model(org).Updates(updates).Error; err != nil {
		return nil, err
	}

	return org, nil
}

// Delete realiza soft delete da organização identificada por slug.
// Retorna ErrOrgNotFound se não existir.
func Delete(db *gorm.DB, slug string) error {
	org, err := FindBySlug(db, slug)
	if err != nil {
		return err
	}
	return db.Delete(org).Error
}

// List retorna orgs paginadas. Se deleted=true, retorna apenas soft-deleted; caso contrário, apenas ativas.
func List(db *gorm.DB, limit, offset int, deleted bool) ([]models.Organization, error) {
	var orgs []models.Organization
	q := db.Limit(limit).Offset(offset)
	if deleted {
		q = q.Unscoped().Where("deleted_at IS NOT NULL")
	}
	if err := q.Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

// MyOrgs retorna todas as orgs ativas nas quais o usuário é owner ou membro ativo.
func MyOrgs(db *gorm.DB, userID uint) ([]models.Organization, error) {
	var orgs []models.Organization
	err := db.
		Distinct("organizations.*").
		Joins("LEFT JOIN org_members ON org_members.organization_id = organizations.id AND org_members.deleted_at IS NULL").
		Where("organizations.owner_id = ? OR org_members.user_id = ?", userID, userID).
		Find(&orgs).Error
	if err != nil {
		return nil, err
	}
	return orgs, nil
}
