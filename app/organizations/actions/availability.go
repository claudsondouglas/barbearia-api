package actions

import (
	"errors"
	"time"

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
	panic("not implemented")
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
