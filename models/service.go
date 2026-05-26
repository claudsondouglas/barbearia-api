// Package models define as estruturas de dados compartilhadas entre as camadas da aplicação.
package models

import (
	"time"

	"gorm.io/gorm"
)

// Service representa um serviço oferecido por uma organização (ex: corte de cabelo, barba).
type Service struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	OrganizationID uint           `gorm:"not null" json:"organization_id"`
	Name           string         `gorm:"not null" json:"name"`
	Description    *string        `json:"description"`
	Price          float64        `gorm:"not null;type:numeric(10,2)" json:"price"`
	DurationMin    int            `gorm:"not null" json:"duration_min"`
	Active         bool           `gorm:"not null;default:true" json:"active"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at"`
}

// CreateServiceInput contém os campos para criar um novo serviço.
type CreateServiceInput struct {
	Name        string   `json:"name" binding:"required"`
	Description *string  `json:"description"`
	Price       float64  `json:"price"`
	DurationMin int      `json:"duration_min"`
	Active      *bool    `json:"active"`
}

// UpdateServiceInput contém os campos opcionais para atualizar um serviço existente.
// Apenas os campos não-nulos são aplicados.
type UpdateServiceInput struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	DurationMin *int     `json:"duration_min"`
	Active      *bool    `json:"active"`
}
