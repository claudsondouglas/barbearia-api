package actions

import (
	"barbearia-api/models"

	"gorm.io/gorm"
)

// RescheduleInput contém os campos para reagendar um agendamento.
type RescheduleInput struct {
	ScheduledAt string `json:"scheduled_at" binding:"required"`
}

// Reschedule altera o horário de um agendamento existente.
// Se o solicitante for cliente, o status volta para "pending"; caso contrário o status é mantido.
// O campo original_scheduled_at é preenchido apenas no primeiro reagendamento e não é alterado depois.
// Registra entrada em schedule_reschedule_history na mesma transação.
// Retorna ErrForbidden, ErrScheduleNotFound, ErrOrgNotFound, ErrInvalidTransition,
// ErrPastScheduledAt, ErrConflict ou ErrUnavailable conforme o caso.
func Reschedule(db *gorm.DB, orgSlug string, scheduleID uint, requestingUserID uint, requestingRole string, input RescheduleInput) (*models.Schedule, error) {
	panic("not implemented")
}

// GetRescheduleHistory retorna o histórico de reagendamentos de um agendamento.
// Owner/admin e o cliente dono do agendamento podem consultar.
// Retorna ErrForbidden, ErrScheduleNotFound ou ErrOrgNotFound conforme o caso.
func GetRescheduleHistory(db *gorm.DB, orgSlug string, scheduleID uint, requestingUserID uint, requestingRole string) ([]models.ScheduleRescheduleHistory, error) {
	panic("not implemented")
}
