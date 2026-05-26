// Package models define as estruturas de dados compartilhadas entre as camadas da aplicação.
package models

import "time"

// Customer representa um cliente de uma organização.
type Customer struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	OrganizationID uint      `gorm:"not null" json:"organization_id"`
	UserID         *uint     `json:"user_id"`
	Name           string    `gorm:"not null" json:"name"`
	Phone          string    `gorm:"not null" json:"phone"`
	Notes          *string   `json:"notes"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateCustomerInput contém os campos para criar um novo customer manualmente.
type CreateCustomerInput struct {
	Name  string  `json:"name" binding:"required"`
	Phone string  `json:"phone" binding:"required"`
	Notes *string `json:"notes"`
}

// UpdateCustomerInput contém os campos para atualizar um customer (todos opcionais).
type UpdateCustomerInput struct {
	Name  *string `json:"name"`
	Phone *string `json:"phone"`
	Notes *string `json:"notes"`
}

// ListCustomersFilter contém os filtros para listar customers de uma organização.
type ListCustomersFilter struct {
	Query  string
	Limit  int
	Offset int
}
