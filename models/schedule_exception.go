// Package models define as estruturas de dados compartilhadas entre as camadas da aplicação.
package models

import "time"

// ScheduleException define um comportamento especial para uma data específica,
// podendo ser no nível da organização (user_id = null) ou de um membro específico.
type ScheduleException struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	OrganizationID uint      `gorm:"not null" json:"organization_id"`
	UserID         *uint     `json:"user_id"` // null = nível org
	Date           string    `gorm:"not null;type:date" json:"date"`
	Closed         bool      `gorm:"not null" json:"closed"`
	OpenTime       *string   `json:"open_time"`
	CloseTime      *string   `json:"close_time"`
	Reason         *string   `json:"reason"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateExceptionInput contém os campos para criar uma exceção de horário.
type CreateExceptionInput struct {
	Date      string  `json:"date" binding:"required"`
	Closed    bool    `json:"closed"`
	OpenTime  *string `json:"open_time"`
	CloseTime *string `json:"close_time"`
	Reason    *string `json:"reason"`
}
