// Package actions implementa a lógica de negócio para o módulo de schedules.
package actions

import (
	"errors"
	"fmt"
	"time"

	"barbearia-api/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	if requestingRole != "user" && requestingRole != "admin" && requestingRole != "owner" {
		return nil, ErrForbidden
	}
	if input.ProfessionalID == 0 {
		return nil, fmt.Errorf("professional_id is required")
	}

	var org models.Organization
	if err := db.Where("slug = ? AND deleted_at IS NULL", orgSlug).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}

	isPrivileged := requestingRole == "admin" || requestingRole == "owner"

	// Validate professional is active member.
	var profMember models.OrgMember
	if err := db.Where("organization_id = ? AND user_id = ? AND deleted_at IS NULL", org.ID, input.ProfessionalID).First(&profMember).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProfessionalNotMember
		}
		return nil, err
	}

	// Active members can only book themselves as the professional.
	if !isPrivileged {
		var requesterMember models.OrgMember
		err := db.Where("organization_id = ? AND user_id = ? AND deleted_at IS NULL", org.ID, requestingUserID).First(&requesterMember).Error
		if err == nil && requestingUserID != input.ProfessionalID {
			return nil, ErrForbidden
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	var service models.Service
	if err := db.Where("id = ? AND organization_id = ? AND active = true AND deleted_at IS NULL", input.ServiceID, org.ID).First(&service).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrServiceNotFound
		}
		return nil, err
	}

	scheduledAt, err := time.Parse(time.RFC3339, input.ScheduledAt)
	if err != nil {
		return nil, fmt.Errorf("invalid scheduled_at: must be RFC3339")
	}
	scheduledAt = scheduledAt.UTC()

	// Determine status and walk-in flag.
	status := "pending"
	isWalkIn := false
	if input.Status == "completed" && (isPrivileged || input.ProfessionalID == requestingUserID) {
		status = "completed"
		isWalkIn = true
	}

	if !isWalkIn && scheduledAt.Before(time.Now().UTC()) {
		return nil, ErrPastScheduledAt
	}

	loc, err := time.LoadLocation(org.Timezone)
	if err != nil {
		return nil, err
	}

	scheduledAtLocal := scheduledAt.In(loc)
	dateStr := scheduledAtLocal.Format("2006-01-02")
	endsAt := scheduledAt.Add(time.Duration(service.DurationMin) * time.Minute)

	// Check org-level schedule exception.
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

	// Check professional exception or fall back to member business hours.
	var profOpenStr, profCloseStr string
	{
		var pe models.ScheduleException
		if err := db.Where("organization_id = ? AND user_id = ? AND date = ?", org.ID, input.ProfessionalID, dateStr).First(&pe).Error; err == nil {
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

	// Calculate effective window, intersecting with org exception if present.
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

	// [scheduledAt, endsAt) must be contained within [effectiveOpen, effectiveClose).
	if scheduledAt.Before(effectiveOpen) || endsAt.After(effectiveClose) {
		return nil, ErrUnavailable
	}

	// Create schedule inside a transaction with SELECT FOR UPDATE to prevent race conditions.
	var schedule models.Schedule
	err = db.Transaction(func(tx *gorm.DB) error {
		var conflicts []models.Schedule
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("professional_id = ? AND status IN ('pending','confirmed') AND scheduled_at < ? AND ends_at > ?",
				input.ProfessionalID, endsAt, scheduledAt).
			Find(&conflicts).Error; err != nil {
			return err
		}
		if len(conflicts) > 0 {
			return ErrConflict
		}

		schedule = models.Schedule{
			OrganizationID:      org.ID,
			ServiceID:           input.ServiceID,
			ProfessionalID:      input.ProfessionalID,
			ClientID:            &requestingUserID,
			Status:              status,
			ScheduledAt:         scheduledAt,
			EndsAt:              endsAt,
			PriceSnapshot:       service.Price,
			DurationMinSnapshot: service.DurationMin,
			Notes:               input.Notes,
		}

		if status == "completed" {
			now := time.Now().UTC()
			schedule.CompletedBy = &requestingUserID
			schedule.CompletedAt = &now
		}

		return tx.Create(&schedule).Error
	})
	if err != nil {
		return nil, err
	}

	return &schedule, nil
}

// List retorna os agendamentos da organização identificada por orgSlug, aplicando os filtros fornecidos.
// Owner/admin veem todos; profissional vê apenas os próprios; cliente recebe ErrForbidden.
func List(db *gorm.DB, orgSlug string, requestingUserID uint, requestingRole string, filter models.ListSchedulesFilter) ([]models.Schedule, error) {
	var org models.Organization
	if err := db.Where("slug = ? AND deleted_at IS NULL", orgSlug).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}

	isPrivileged := requestingRole == "owner" || requestingRole == "admin"

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

	if !isPrivileged && !isProfessional {
		return nil, ErrForbidden
	}

	query := db.Where("organization_id = ?", org.ID)

	if !isPrivileged {
		query = query.Where("professional_id = ?", requestingUserID)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.ProfessionalID != nil {
		query = query.Where("professional_id = ?", *filter.ProfessionalID)
	}

	if filter.Date != "" {
		loc, err := time.LoadLocation(org.Timezone)
		if err != nil {
			return nil, err
		}
		startOfDay, err := time.ParseInLocation("2006-01-02", filter.Date, loc)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
		startOfNextDay := startOfDay.AddDate(0, 0, 1)
		query = query.Where("scheduled_at >= ? AND scheduled_at < ?", startOfDay.UTC(), startOfNextDay.UTC())
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	var schedules []models.Schedule
	if err := query.Order("scheduled_at ASC").Limit(limit).Offset(offset).Find(&schedules).Error; err != nil {
		return nil, err
	}

	return schedules, nil
}

// Find retorna o detalhe de um agendamento pelo ID.
// Controla acesso: owner/admin veem qualquer; profissional vê apenas os próprios;
// cliente vê apenas os seus. Retorna ErrScheduleNotFound ou ErrForbidden conforme o caso.
func Find(db *gorm.DB, orgSlug string, scheduleID uint, requestingUserID uint, requestingRole string) (*models.Schedule, error) {
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
	if isPrivileged {
		return &schedule, nil
	}

	isProfessional := false
	var member models.OrgMember
	err := db.Where("organization_id = ? AND user_id = ? AND deleted_at IS NULL", org.ID, requestingUserID).First(&member).Error
	if err == nil {
		isProfessional = true
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if isProfessional {
		if schedule.ProfessionalID != requestingUserID {
			return nil, ErrForbidden
		}
		return &schedule, nil
	}

	// Client: can only see their own schedules.
	if schedule.ClientID == nil || *schedule.ClientID != requestingUserID {
		return nil, ErrForbidden
	}

	return &schedule, nil
}

// MySchedules retorna todos os agendamentos do usuário autenticado em todas as organizações.
func MySchedules(db *gorm.DB, userID uint, filter models.ListSchedulesFilter) ([]models.Schedule, error) {
	query := db.Where("client_id = ?", userID)

	if filter.OrganizationSlug != "" {
		query = query.Joins("JOIN organizations ON organizations.id = schedules.organization_id").
			Where("organizations.slug = ? AND organizations.deleted_at IS NULL", filter.OrganizationSlug)
	}

	if filter.Status != "" {
		query = query.Where("schedules.status = ?", filter.Status)
	}

	if filter.Date != "" {
		startOfDay, err := time.Parse("2006-01-02", filter.Date)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
		startOfNextDay := startOfDay.AddDate(0, 0, 1)
		query = query.Where("schedules.scheduled_at >= ? AND schedules.scheduled_at < ?", startOfDay.UTC(), startOfNextDay.UTC())
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	var schedules []models.Schedule
	if err := query.Order("schedules.scheduled_at ASC").Limit(limit).Offset(offset).Find(&schedules).Error; err != nil {
		return nil, err
	}

	return schedules, nil
}
