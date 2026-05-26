// Package models define as estruturas de dados compartilhadas entre as camadas da aplicação.
package models

import (
	"time"

	"gorm.io/gorm"
)

// Organization representa uma organização (barbearia) cadastrada no sistema.
type Organization struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Slug         string         `gorm:"uniqueIndex;not null" json:"slug"`
	OwnerID      uint           `gorm:"not null" json:"owner_id"`
	Name         string         `gorm:"not null" json:"name"`
	Phone        string         `gorm:"not null" json:"phone"`
	Email        string         `gorm:"not null" json:"email"`
	Description  *string        `json:"description"`
	LogoURL      *string        `json:"logo_url"`
	Street       string         `gorm:"not null" json:"street"`
	Number       string         `gorm:"not null" json:"number"`
	Complement   *string        `json:"complement"`
	Neighborhood string         `gorm:"not null" json:"neighborhood"`
	City         string         `gorm:"not null" json:"city"`
	State        string         `gorm:"not null" json:"state"`
	ZipCode      string         `gorm:"not null" json:"zip_code"`
	Latitude     *float64       `json:"latitude"`
	Longitude    *float64       `json:"longitude"`
	Timezone     string         `gorm:"not null;default:'America/Sao_Paulo'" json:"timezone"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-"`
}

// OrgResponse é a representação pública de uma organização (sem owner_id e deleted_at).
type OrgResponse struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Phone        string    `json:"phone"`
	Email        string    `json:"email"`
	Description  *string   `json:"description"`
	LogoURL      *string   `json:"logo_url"`
	Street       string    `json:"street"`
	Number       string    `json:"number"`
	Complement   *string   `json:"complement"`
	Neighborhood string    `json:"neighborhood"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	ZipCode      string    `json:"zip_code"`
	Latitude     *float64  `json:"latitude"`
	Longitude    *float64  `json:"longitude"`
	Timezone     string    `json:"timezone"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ToResponse converte Organization em OrgResponse para exposição via API.
func (o *Organization) ToResponse() OrgResponse {
	return OrgResponse{
		ID:           o.ID,
		Name:         o.Name,
		Slug:         o.Slug,
		Phone:        o.Phone,
		Email:        o.Email,
		Description:  o.Description,
		LogoURL:      o.LogoURL,
		Street:       o.Street,
		Number:       o.Number,
		Complement:   o.Complement,
		Neighborhood: o.Neighborhood,
		City:         o.City,
		State:        o.State,
		ZipCode:      o.ZipCode,
		Latitude:     o.Latitude,
		Longitude:    o.Longitude,
		Timezone:     o.Timezone,
		CreatedAt:    o.CreatedAt,
		UpdatedAt:    o.UpdatedAt,
	}
}

// OrgMember representa o vínculo entre um usuário e uma organização.
type OrgMember struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	OrganizationID uint           `gorm:"not null" json:"organization_id"`
	UserID         uint           `gorm:"not null" json:"user_id"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at"`
}

// MemberBusinessHour representa o horário de funcionamento de um membro em um dia da semana.
type MemberBusinessHour struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	OrgMemberID uint    `gorm:"not null" json:"org_member_id"`
	DayOfWeek   int     `gorm:"not null" json:"day_of_week"` // 0=Sunday ... 6=Saturday
	Closed      bool    `gorm:"not null;default:true" json:"closed"`
	OpenTime    *string `json:"open_time"`
	CloseTime   *string `json:"close_time"`
}

// AddMemberInput contém os campos obrigatórios para adicionar um membro à organização.
type AddMemberInput struct {
	UserID uint `json:"user_id" binding:"required"`
}

// MemberResponse representa os dados de um membro retornados pela API.
type MemberResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
