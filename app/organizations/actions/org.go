// Package actions implementa a lógica de negócio para o módulo de organizations.
package actions

import (
	"errors"
	"regexp"
	"strings"
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
	CNPJ         *string  `json:"cnpj"`
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
	CNPJ         *string  `json:"cnpj"`
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
	// Normaliza para NFD para separar letras de seus diacríticos.
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	result, _, _ := transform.String(t, name)

	result = strings.ToLower(result)

	// Substitui qualquer sequência de caracteres não alfanuméricos por um hífen.
	re := regexp.MustCompile(`[^a-z0-9]+`)
	result = re.ReplaceAllString(result, "-")

	// Remove hífens no início e no fim.
	result = strings.Trim(result, "-")

	return result
}

// Create cria uma nova organização com o owner indicado.
// Executa em transação: cria org + OrgMember + 7 MemberBusinessHours.
// Valida timezone IANA; gera slug único a partir do nome.
func Create(db *gorm.DB, ownerID uint, input CreateOrgInput) (*models.Organization, error) {
	panic("not implemented")
}

// FindBySlug busca uma organização ativa pelo slug.
// Retorna ErrOrgNotFound se não existir ou estiver soft-deleted.
func FindBySlug(db *gorm.DB, slug string) (*models.Organization, error) {
	panic("not implemented")
}

// Update aplica campos opcionais de UpdateOrgInput na organização identificada por slug.
// Apenas owner ou admin (requestingUserID) pode atualizar.
// Se o nome mudar, regenera o slug.
// Retorna ErrForbidden, ErrOrgNotFound ou ErrInvalidTimezone conforme o caso.
func Update(db *gorm.DB, slug string, requestingUserID uint, input UpdateOrgInput) (*models.Organization, error) {
	panic("not implemented")
}

// Delete realiza soft delete da organização identificada por slug.
// Retorna ErrOrgNotFound se não existir.
func Delete(db *gorm.DB, slug string) error {
	panic("not implemented")
}

// List retorna orgs paginadas. Se deleted=true, retorna apenas soft-deleted; caso contrário, apenas ativas.
func List(db *gorm.DB, limit, offset int, deleted bool) ([]models.Organization, error) {
	panic("not implemented")
}

// MyOrgs retorna todas as orgs ativas nas quais o usuário é owner ou membro ativo.
func MyOrgs(db *gorm.DB, userID uint) ([]models.Organization, error) {
	panic("not implemented")
}
