package actions

import (
	"errors"
	"fmt"
	"time"

	"barbearia-api/models"

	"gorm.io/gorm"
)

var ErrProfessionalNotMember = errors.New("professional is not a member")
var ErrServiceNotFound = errors.New("service not found")
var ErrDateInPast = errors.New("date is in the past")
var ErrDateBeyond90Days = errors.New("date is beyond 90 days")

// AvailabilityResult contém os slots disponíveis para um profissional em uma data.
type AvailabilityResult struct {
	Date           string   `json:"date"`
	ProfessionalID uint     `json:"professional_id"`
	ServiceID      uint     `json:"service_id"`
	DurationMin    int      `json:"duration_min"`
	Slots          []string `json:"slots"`
}

// GetAvailability calcula os slots disponíveis de um profissional para um serviço em uma data.
// Considera: timezone da org, exceções de horário (org e membro), horários semanais e agendamentos existentes.
// nowFn permite injeção do tempo atual nos testes.
func GetAvailability(db *gorm.DB, orgSlug string, professionalID, serviceID uint, date string, nowFn func() time.Time) (AvailabilityResult, error) {
	org, err := FindBySlug(db, orgSlug)
	if err != nil {
		return AvailabilityResult{}, err
	}

	loc, err := time.LoadLocation(org.Timezone)
	if err != nil {
		return AvailabilityResult{}, err
	}

	parsedDate, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		return AvailabilityResult{}, fmt.Errorf("invalid date format: must be YYYY-MM-DD")
	}

	now := nowFn()
	nowInLoc := now.In(loc)
	todayInLoc := time.Date(nowInLoc.Year(), nowInLoc.Month(), nowInLoc.Day(), 0, 0, 0, 0, loc)

	if parsedDate.Before(todayInLoc) {
		return AvailabilityResult{}, ErrDateInPast
	}
	if parsedDate.After(todayInLoc.AddDate(0, 0, 90)) {
		return AvailabilityResult{}, ErrDateBeyond90Days
	}

	var member models.OrgMember
	if err := db.Where("organization_id = ? AND user_id = ? AND deleted_at IS NULL", org.ID, professionalID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return AvailabilityResult{}, ErrProfessionalNotMember
		}
		return AvailabilityResult{}, err
	}

	var service models.Service
	if err := db.Where("id = ? AND organization_id = ? AND active = true AND deleted_at IS NULL", serviceID, org.ID).First(&service).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return AvailabilityResult{}, ErrServiceNotFound
		}
		return AvailabilityResult{}, err
	}

	empty := AvailabilityResult{
		Date:           date,
		ProfessionalID: professionalID,
		ServiceID:      serviceID,
		DurationMin:    service.DurationMin,
		Slots:          []string{},
	}

	// Check org exception for this date.
	var orgException *models.ScheduleException
	{
		var oe models.ScheduleException
		if err := db.Where("organization_id = ? AND user_id IS NULL AND date = ?", org.ID, date).First(&oe).Error; err == nil {
			orgException = &oe
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return empty, err
		}
	}

	if orgException != nil && orgException.Closed {
		return empty, nil
	}

	// Check professional exception or fall back to member business hours.
	var profOpenStr, profCloseStr string
	{
		var pe models.ScheduleException
		if err := db.Where("organization_id = ? AND user_id = ? AND date = ?", org.ID, professionalID, date).First(&pe).Error; err == nil {
			if pe.Closed {
				return empty, nil
			}
			if pe.OpenTime == nil || pe.CloseTime == nil {
				return empty, nil
			}
			profOpenStr = *pe.OpenTime
			profCloseStr = *pe.CloseTime
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			dayOfWeek := int(parsedDate.Weekday())
			var bh models.MemberBusinessHour
			if err := db.Where("org_member_id = ? AND day_of_week = ?", member.ID, dayOfWeek).First(&bh).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return empty, nil
				}
				return empty, err
			}
			if bh.Closed || bh.OpenTime == nil || bh.CloseTime == nil {
				return empty, nil
			}
			profOpenStr = *bh.OpenTime
			profCloseStr = *bh.CloseTime
		} else {
			return empty, err
		}
	}

	// Determine effective window (intersect with org exception if present).
	effectiveOpenStr := profOpenStr
	effectiveCloseStr := profCloseStr

	if orgException != nil && !orgException.Closed {
		if orgException.OpenTime == nil || orgException.CloseTime == nil {
			return empty, nil
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
			return empty, nil
		}
	}

	effectiveOpen, err := time.ParseInLocation("2006-01-02 15:04", date+" "+effectiveOpenStr, loc)
	if err != nil {
		return empty, err
	}
	effectiveClose, err := time.ParseInLocation("2006-01-02 15:04", date+" "+effectiveCloseStr, loc)
	if err != nil {
		return empty, err
	}

	slots := GenerateSlots(effectiveOpen, effectiveClose, service.DurationMin)

	if parsedDate.Equal(todayInLoc) {
		slots = DiscardPastSlots(slots, nowInLoc.Truncate(time.Minute))
	}

	if len(slots) > 0 {
		duration := time.Duration(service.DurationMin) * time.Minute
		dayStart := parsedDate
		dayEnd := parsedDate.Add(24 * time.Hour)

		var appointments []models.Schedule
		if err := db.Where(
			"professional_id = ? AND status IN ('pending','confirmed') AND scheduled_at < ? AND ends_at > ?",
			professionalID, dayEnd, dayStart,
		).Find(&appointments).Error; err != nil {
			return empty, err
		}

		var available []time.Time
		for _, slot := range slots {
			slotEnd := slot.Add(duration)
			conflict := false
			for _, appt := range appointments {
				if Overlaps(slot, slotEnd, appt.ScheduledAt, appt.EndsAt) {
					conflict = true
					break
				}
			}
			if !conflict {
				available = append(available, slot)
			}
		}
		slots = available
	}

	slotStrings := make([]string, 0, len(slots))
	for _, s := range slots {
		slotStrings = append(slotStrings, s.In(loc).Format("15:04"))
	}

	return AvailabilityResult{
		Date:           date,
		ProfessionalID: professionalID,
		ServiceID:      serviceID,
		DurationMin:    service.DurationMin,
		Slots:          slotStrings,
	}, nil
}

// GenerateSlots gera candidatos de slot a partir de uma janela de tempo e duração.
// O último slot gerado satisfaz: slotStart + durationMin <= windowClose.
// windowClose é exclusivo: o slot pode terminar exatamente em windowClose, mas não depois.
func GenerateSlots(windowOpen, windowClose time.Time, durationMin int) []time.Time {
	if durationMin <= 0 {
		return nil
	}

	duration := time.Duration(durationMin) * time.Minute
	var slots []time.Time

	current := windowOpen
	for !current.Add(duration).After(windowClose) {
		slots = append(slots, current)
		current = current.Add(duration)
	}

	return slots
}

// Overlaps verifica se o intervalo [slotStart, slotEnd) se sobrepõe a [apptStart, apptEnd).
// A condição é: apptStart < slotEnd AND apptEnd > slotStart.
// Intervalos adjacentes (touching) NÃO são considerados sobrepostos.
func Overlaps(slotStart, slotEnd, apptStart, apptEnd time.Time) bool {
	return apptStart.Before(slotEnd) && apptEnd.After(slotStart)
}

// DiscardPastSlots remove slots cujo início seja anterior ao instante now.
// Slots cujo início seja exatamente igual a now são mantidos.
func DiscardPastSlots(slots []time.Time, now time.Time) []time.Time {
	var result []time.Time
	for _, s := range slots {
		if !s.Before(now) {
			result = append(result, s)
		}
	}
	return result
}
