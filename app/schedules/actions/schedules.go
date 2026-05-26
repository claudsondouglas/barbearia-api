// Package actions implementa a lógica de negócio para o módulo de schedules.
package actions

import (
	"errors"

	"barbearia-api/models"

	"gorm.io/gorm"
)

// Erros sentinela do módulo de schedules.
var (
	ErrScheduleNotFound      = errors.New("schedule not found")
	ErrConflict              = errors.New("schedule conflict")
	ErrInvalidTransition     = errors.New("invalid status transition")
	ErrForbidden             = errors.New("forbidden")
	ErrOrgNotFound           = errors.New("organization not found")
	ErrProfessionalNotMember = errors.New("professional is not a member")
	ErrServiceNotFound       = errors.New("service not found")
	ErrPastScheduledAt       = errors.New("scheduled_at is in the past")
	ErrUnavailable           = errors.New("professional is unavailable at this time")
)

// Create cria um novo agendamento na organização identificada por orgSlug.
// Valida disponibilidade do profissional, conflito de horário e snapshot de preço/duração.
// Retorna ErrOrgNotFound, ErrForbidden, ErrProfessionalNotMember, ErrServiceNotFound,
// ErrPastScheduledAt, ErrConflict ou ErrUnavailable conforme o caso.
func Create(db *gorm.DB, orgSlug string, requestingUserID uint, requestingRole string, input models.CreateScheduleInput) (*models.Schedule, error) {
	panic("not implemented")
}

// List retorna os agendamentos da organização identificada por orgSlug, aplicando os filtros fornecidos.
// Owner/admin veem todos; profissional vê apenas os próprios; cliente recebe ErrForbidden.
func List(db *gorm.DB, orgSlug string, requestingUserID uint, requestingRole string, filter models.ListSchedulesFilter) ([]models.Schedule, error) {
	panic("not implemented")
}

// Find retorna o detalhe de um agendamento pelo ID.
// Controla acesso: owner/admin veem qualquer; profissional vê apenas os próprios;
// cliente vê apenas os seus. Retorna ErrScheduleNotFound ou ErrForbidden conforme o caso.
func Find(db *gorm.DB, orgSlug string, scheduleID uint, requestingUserID uint, requestingRole string) (*models.Schedule, error) {
	panic("not implemented")
}

// MySchedules retorna todos os agendamentos do usuário autenticado em todas as organizações.
func MySchedules(db *gorm.DB, userID uint, filter models.ListSchedulesFilter) ([]models.Schedule, error) {
	panic("not implemented")
}
