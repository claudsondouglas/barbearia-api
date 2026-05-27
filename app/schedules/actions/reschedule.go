package actions

import (
	"errors"
	"fmt"
	"time"

	"barbearia-api/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	var org models.Organization
	if err := db.Where("slug = ? AND deleted_at IS NULL", orgSlug).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}

	var schedule models.Schedule
	if err := db.Where("id = ? AND organization_id = ?", scheduleID, org.ID).First(&schedule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}

	isPrivileged := requestingRole == "owner" || requestingRole == "admin"
	isClient := schedule.ClientID != nil && *schedule.ClientID == requestingUserID

	isProfessional := false
	if !isPrivileged {
		var member models.OrgMember
		err := db.Where("organization_id = ? AND user_id = ? AND deleted_at IS NULL", org.ID, requestingUserID).First(&member).Error
		if err == nil {
			isProfessional = true
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	if !isPrivileged {
		if isProfessional {
			if schedule.ProfessionalID != requestingUserID {
				return nil, ErrForbidden
			}
		} else if !isClient {
			return nil, ErrForbidden
		}
	}

	if schedule.Status != "pending" && schedule.Status != "confirmed" {
		return nil, ErrInvalidTransition
	}

	newScheduledAt, err := time.Parse(time.RFC3339, input.ScheduledAt)
	if err != nil {
		return nil, fmt.Errorf("invalid scheduled_at: must be RFC3339")
	}
	newScheduledAt = newScheduledAt.UTC()

	if newScheduledAt.Before(time.Now().UTC()) {
		return nil, ErrPastScheduledAt
	}

	loc, err := time.LoadLocation(org.Timezone)
	if err != nil {
		return nil, err
	}

	scheduledAtLocal := newScheduledAt.In(loc)
	dateStr := scheduledAtLocal.Format("2006-01-02")
	newEndsAt := newScheduledAt.Add(time.Duration(schedule.DurationMinSnapshot) * time.Minute)

	var orgException *models.ScheduleException
	{
		var oe models.ScheduleException
		if err := db.Where("organization_id = ? AND user_id IS NULL AND date = ?", org.ID, dateStr).First(&oe).Error; err == nil {
			orgException = &oe
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	if orgException != nil && orgException.Closed {
		return nil, ErrUnavailable
	}

	var profMember models.OrgMember
	if err := db.Where("organization_id = ? AND user_id = ? AND deleted_at IS NULL", org.ID, schedule.ProfessionalID).First(&profMember).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProfessionalNotMember
		}
		return nil, err
	}

	var profOpenStr, profCloseStr string
	{
		var pe models.ScheduleException
		if err := db.Where("organization_id = ? AND user_id = ? AND date = ?", org.ID, schedule.ProfessionalID, dateStr).First(&pe).Error; err == nil {
			if pe.Closed {
				return nil, ErrUnavailable
			}
			if pe.OpenTime == nil || pe.CloseTime == nil {
				return nil, ErrUnavailable
			}
			profOpenStr = *pe.OpenTime
			profCloseStr = *pe.CloseTime
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			dayOfWeek := int(scheduledAtLocal.Weekday())
			var bh models.MemberBusinessHour
			if err := db.Where("org_member_id = ? AND day_of_week = ?", profMember.ID, dayOfWeek).First(&bh).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, ErrUnavailable
				}
				return nil, err
			}
			if bh.Closed || bh.OpenTime == nil || bh.CloseTime == nil {
				return nil, ErrUnavailable
			}
			profOpenStr = *bh.OpenTime
			profCloseStr = *bh.CloseTime
		} else {
			return nil, err
		}
	}

	effectiveOpenStr := profOpenStr
	effectiveCloseStr := profCloseStr
	if orgException != nil && !orgException.Closed {
		if orgException.OpenTime == nil || orgException.CloseTime == nil {
			return nil, ErrUnavailable
		}
		orgOpenStr := *orgException.OpenTime
		orgCloseStr := *orgException.CloseTime
		if profOpenStr < orgOpenStr {
			effectiveOpenStr = orgOpenStr
		}
		if profCloseStr > orgCloseStr {
			effectiveCloseStr = orgCloseStr
		}
		if effectiveOpenStr >= effectiveCloseStr {
			return nil, ErrUnavailable
		}
	}

	effectiveOpen, err := time.ParseInLocation("2006-01-02 15:04", dateStr+" "+effectiveOpenStr, loc)
	if err != nil {
		return nil, err
	}
	effectiveClose, err := time.ParseInLocation("2006-01-02 15:04", dateStr+" "+effectiveCloseStr, loc)
	if err != nil {
		return nil, err
	}

	if newScheduledAt.Before(effectiveOpen) || newEndsAt.After(effectiveClose) {
		return nil, ErrUnavailable
	}

	now := time.Now().UTC()

	err = db.Transaction(func(tx *gorm.DB) error {
		var conflicts []models.Schedule
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("professional_id = ? AND status IN ('pending','confirmed') AND scheduled_at < ? AND ends_at > ? AND id != ?",
				schedule.ProfessionalID, newEndsAt, newScheduledAt, scheduleID).
			Find(&conflicts).Error; err != nil {
			return err
		}
		if len(conflicts) > 0 {
			return ErrConflict
		}

		fromScheduledAt := schedule.ScheduledAt

		if schedule.OriginalScheduledAt == nil {
			schedule.OriginalScheduledAt = &fromScheduledAt
		}

		schedule.ScheduledAt = newScheduledAt
		schedule.EndsAt = newEndsAt
		schedule.RescheduledAt = &now
		schedule.RescheduledBy = &requestingUserID

		if isClient && !isPrivileged && !isProfessional {
			schedule.Status = "pending"
		}

		if err := tx.Save(&schedule).Error; err != nil {
			return err
		}

		history := models.ScheduleRescheduleHistory{
			ScheduleID:      schedule.ID,
			FromScheduledAt: fromScheduledAt,
			ToScheduledAt:   newScheduledAt,
			RescheduledBy:   requestingUserID,
			RescheduledAt:   now,
		}
		return tx.Create(&history).Error
	})
	if err != nil {
		return nil, err
	}

	return &schedule, nil
}

// GetRescheduleHistory retorna o histórico de reagendamentos de um agendamento.
// Owner/admin e o cliente dono do agendamento podem consultar.
// Retorna ErrForbidden, ErrScheduleNotFound ou ErrOrgNotFound conforme o caso.
func GetRescheduleHistory(db *gorm.DB, orgSlug string, scheduleID uint, requestingUserID uint, requestingRole string) ([]models.ScheduleRescheduleHistory, error) {
	var org models.Organization
	if err := db.Where("slug = ? AND deleted_at IS NULL", orgSlug).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}

	var schedule models.Schedule
	if err := db.Where("id = ? AND organization_id = ?", scheduleID, org.ID).First(&schedule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrScheduleNotFound
		}
		return nil, err
	}

	isPrivileged := requestingRole == "owner" || requestingRole == "admin"
	isClient := schedule.ClientID != nil && *schedule.ClientID == requestingUserID

	if !isPrivileged && !isClient {
		return nil, ErrForbidden
	}

	var history []models.ScheduleRescheduleHistory
	if err := db.Where("schedule_id = ?", schedule.ID).Order("rescheduled_at ASC").Find(&history).Error; err != nil {
		return nil, err
	}

	return history, nil
}
