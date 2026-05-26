// Package models define as estruturas de dados compartilhadas entre as camadas da aplicação.
package models

import "time"

// Schedule representa um agendamento de serviço em uma organização.
type Schedule struct {
	ID                  uint       `gorm:"primaryKey" json:"id"`
	OrganizationID      uint       `gorm:"not null" json:"organization_id"`
	ServiceID           uint       `gorm:"not null" json:"service_id"`
	ProfessionalID      uint       `gorm:"not null" json:"professional_id"`
	ClientID            *uint      `json:"client_id"`
	CustomerID          *uint      `json:"customer_id"`
	Status              string     `gorm:"not null;default:pending" json:"status"`
	ScheduledAt         time.Time  `gorm:"not null" json:"scheduled_at"`
	EndsAt              time.Time  `gorm:"not null" json:"ends_at"`
	PriceSnapshot       float64    `gorm:"not null;type:numeric(10,2)" json:"price_snapshot"`
	DurationMinSnapshot int        `gorm:"not null" json:"duration_min_snapshot"`
	Notes               *string    `json:"notes"`
	OriginalScheduledAt *time.Time `json:"original_scheduled_at"`
	RescheduledAt       *time.Time `json:"rescheduled_at"`
	RescheduledBy       *uint      `json:"rescheduled_by"`
	ConfirmedAt         *time.Time `json:"confirmed_at"`
	ConfirmedBy         *uint      `json:"confirmed_by"`
	CancelledAt         *time.Time `json:"cancelled_at"`
	CancelledBy         *uint      `json:"cancelled_by"`
	CompletedAt         *time.Time `json:"completed_at"`
	CompletedBy         *uint      `json:"completed_by"`
	NoShowAt            *time.Time `json:"no_show_at"`
	NoShowBy            *uint      `json:"no_show_by"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// ScheduleRescheduleHistory registra o histórico de reagendamentos de um agendamento.
type ScheduleRescheduleHistory struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ScheduleID      uint      `gorm:"not null" json:"schedule_id"`
	FromScheduledAt time.Time `gorm:"not null" json:"from_scheduled_at"`
	ToScheduledAt   time.Time `gorm:"not null" json:"to_scheduled_at"`
	RescheduledBy   uint      `gorm:"not null" json:"rescheduled_by"`
	RescheduledAt   time.Time `gorm:"not null" json:"rescheduled_at"`
}

// CreateScheduleInput contém os campos para criar um novo agendamento.
type CreateScheduleInput struct {
	ServiceID      uint    `json:"service_id" binding:"required"`
	ProfessionalID uint    `json:"professional_id" binding:"required"`
	ScheduledAt    string  `json:"scheduled_at" binding:"required"`
	Notes          *string `json:"notes"`
	Status         string  `json:"status"`
}

// ListSchedulesFilter contém os filtros para listar agendamentos.
type ListSchedulesFilter struct {
	Status         string
	Date           string
	ProfessionalID *uint
	Limit          int
	Offset         int
}
