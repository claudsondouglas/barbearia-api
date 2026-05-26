package actions

import (
	"barbearia-api/models"

	"gorm.io/gorm"
)

// Confirm confirma um agendamento com status "pending".
// Apenas owner, admin ou o profissional responsável podem confirmar.
// Retorna ErrForbidden, ErrScheduleNotFound, ErrOrgNotFound ou ErrInvalidTransition.
func Confirm(db *gorm.DB, orgSlug string, scheduleID uint, requestingUserID uint, requestingRole string) (*models.Schedule, error) {
	panic("not implemented")
}

// Cancel cancela um agendamento com status "pending" ou "confirmed".
// Cliente pode cancelar apenas o próprio; profissional pode cancelar apenas os seus;
// owner/admin cancelam qualquer agendamento da org.
// Retorna ErrForbidden, ErrScheduleNotFound, ErrOrgNotFound ou ErrInvalidTransition.
func Cancel(db *gorm.DB, orgSlug string, scheduleID uint, requestingUserID uint, requestingRole string) (*models.Schedule, error) {
	panic("not implemented")
}

// Complete conclui um agendamento com status "confirmed".
// Apenas owner, admin ou o profissional responsável podem concluir.
// Retorna ErrForbidden, ErrScheduleNotFound, ErrOrgNotFound ou ErrInvalidTransition.
func Complete(db *gorm.DB, orgSlug string, scheduleID uint, requestingUserID uint, requestingRole string) (*models.Schedule, error) {
	panic("not implemented")
}

// NoShow marca um agendamento confirmado como "no_show".
// Apenas owner, admin ou o profissional responsável podem marcar no-show.
// Retorna ErrForbidden, ErrScheduleNotFound, ErrOrgNotFound ou ErrInvalidTransition.
func NoShow(db *gorm.DB, orgSlug string, scheduleID uint, requestingUserID uint, requestingRole string) (*models.Schedule, error) {
	panic("not implemented")
}
