package actions

import (
	"errors"
	"time"

	"barbearia-api/models"

	"gorm.io/gorm"
)

func fetchOrgAndSchedule(db *gorm.DB, orgSlug string, scheduleID uint) (*models.Organization, *models.Schedule, error) {
	var org models.Organization
	if err := db.Where("slug = ? AND deleted_at IS NULL", orgSlug).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrOrgNotFound
		}
		return nil, nil, err
	}

	var schedule models.Schedule
	if err := db.Where("id = ? AND organization_id = ?", scheduleID, org.ID).First(&schedule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrScheduleNotFound
		}
		return nil, nil, err
	}

	return &org, &schedule, nil
}

// Confirm confirma um agendamento com status "pending".
// Apenas owner, admin ou o profissional responsável podem confirmar.
// Retorna ErrForbidden, ErrScheduleNotFound, ErrOrgNotFound ou ErrInvalidTransition.
func Confirm(db *gorm.DB, orgSlug string, scheduleID uint, requestingUserID uint, requestingRole string) (*models.Schedule, error) {
	_, schedule, err := fetchOrgAndSchedule(db, orgSlug, scheduleID)
	if err != nil {
		return nil, err
	}

	isPrivileged := requestingRole == "owner" || requestingRole == "admin"
	isProfessional := schedule.ProfessionalID == requestingUserID

	if !isPrivileged && !isProfessional {
		return nil, ErrForbidden
	}

	if schedule.Status != "pending" {
		return nil, ErrInvalidTransition
	}

	now := time.Now().UTC()
	schedule.Status = "confirmed"
	schedule.ConfirmedBy = &requestingUserID
	schedule.ConfirmedAt = &now

	if err := db.Save(schedule).Error; err != nil {
		return nil, err
	}

	return schedule, nil
}

// Cancel cancela um agendamento com status "pending" ou "confirmed".
// Cliente pode cancelar apenas o próprio; profissional pode cancelar apenas os seus;
// owner/admin cancelam qualquer agendamento da org.
// Retorna ErrForbidden, ErrScheduleNotFound, ErrOrgNotFound ou ErrInvalidTransition.
func Cancel(db *gorm.DB, orgSlug string, scheduleID uint, requestingUserID uint, requestingRole string) (*models.Schedule, error) {
	_, schedule, err := fetchOrgAndSchedule(db, orgSlug, scheduleID)
	if err != nil {
		return nil, err
	}

	isPrivileged := requestingRole == "owner" || requestingRole == "admin"
	isProfessional := schedule.ProfessionalID == requestingUserID
	isClient := schedule.ClientID != nil && *schedule.ClientID == requestingUserID

	if !isPrivileged && !isProfessional && !isClient {
		return nil, ErrForbidden
	}

	if schedule.Status != "pending" && schedule.Status != "confirmed" {
		return nil, ErrInvalidTransition
	}

	now := time.Now().UTC()
	schedule.Status = "cancelled"
	schedule.CancelledBy = &requestingUserID
	schedule.CancelledAt = &now

	if err := db.Save(schedule).Error; err != nil {
		return nil, err
	}

	return schedule, nil
}

// Complete conclui um agendamento com status "confirmed".
// Apenas owner, admin ou o profissional responsável podem concluir.
// Retorna ErrForbidden, ErrScheduleNotFound, ErrOrgNotFound ou ErrInvalidTransition.
func Complete(db *gorm.DB, orgSlug string, scheduleID uint, requestingUserID uint, requestingRole string) (*models.Schedule, error) {
	_, schedule, err := fetchOrgAndSchedule(db, orgSlug, scheduleID)
	if err != nil {
		return nil, err
	}

	isPrivileged := requestingRole == "owner" || requestingRole == "admin"
	isProfessional := schedule.ProfessionalID == requestingUserID

	if !isPrivileged && !isProfessional {
		return nil, ErrForbidden
	}

	if schedule.Status != "confirmed" {
		return nil, ErrInvalidTransition
	}

	now := time.Now().UTC()
	schedule.Status = "completed"
	schedule.CompletedBy = &requestingUserID
	schedule.CompletedAt = &now

	if err := db.Save(schedule).Error; err != nil {
		return nil, err
	}

	return schedule, nil
}

// NoShow marca um agendamento confirmado como "no_show".
// Apenas owner, admin ou o profissional responsável podem marcar no-show.
// Retorna ErrForbidden, ErrScheduleNotFound, ErrOrgNotFound ou ErrInvalidTransition.
func NoShow(db *gorm.DB, orgSlug string, scheduleID uint, requestingUserID uint, requestingRole string) (*models.Schedule, error) {
	_, schedule, err := fetchOrgAndSchedule(db, orgSlug, scheduleID)
	if err != nil {
		return nil, err
	}

	isPrivileged := requestingRole == "owner" || requestingRole == "admin"
	isProfessional := schedule.ProfessionalID == requestingUserID

	if !isPrivileged && !isProfessional {
		return nil, ErrForbidden
	}

	if schedule.Status != "confirmed" {
		return nil, ErrInvalidTransition
	}

	now := time.Now().UTC()
	schedule.Status = "no_show"
	schedule.NoShowBy = &requestingUserID
	schedule.NoShowAt = &now

	if err := db.Save(schedule).Error; err != nil {
		return nil, err
	}

	return schedule, nil
}
