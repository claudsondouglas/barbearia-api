package actions

import (
	"barbearia-api/models"

	"gorm.io/gorm"
)

// BusinessHourDay representa o horário de funcionamento agregado de um dia da semana.
type BusinessHourDay struct {
	DayOfWeek int     `json:"day_of_week"`
	OpenTime  *string `json:"open_time,omitempty"`
	CloseTime *string `json:"close_time,omitempty"`
	Closed    bool    `json:"closed"`
}

// GetBusinessHours busca os business hours agregados de uma org pelo slug.
// Retorna ErrOrgNotFound se a org não existir ou estiver soft-deleted.
func GetBusinessHours(db *gorm.DB, orgSlug string) ([]BusinessHourDay, error) {
	org, err := FindBySlug(db, orgSlug)
	if err != nil {
		return nil, err
	}

	var memberIDs []uint
	if err := db.Model(&models.OrgMember{}).
		Where("organization_id = ? AND deleted_at IS NULL", org.ID).
		Pluck("id", &memberIDs).Error; err != nil {
		return nil, err
	}

	if len(memberIDs) == 0 {
		return AggregateBusinessHours(nil), nil
	}

	var hours []models.MemberBusinessHour
	if err := db.Where("org_member_id IN ?", memberIDs).Find(&hours).Error; err != nil {
		return nil, err
	}

	return AggregateBusinessHours(hours), nil
}

// AggregateBusinessHours calcula o horário de funcionamento da organização
// a partir dos horários individuais dos membros ativos.
// Para cada dia da semana (0–6):
//   - Se ao menos um membro tiver closed=false, o dia é aberto.
//   - open_time = MIN(open_time dos membros abertos naquele dia).
//   - close_time = MAX(close_time dos membros abertos naquele dia).
//   - Se todos os membros estiverem fechados (ou não houver membros), closed=true.
func AggregateBusinessHours(hours []models.MemberBusinessHour) []BusinessHourDay {
	// Inicializa os 7 dias como fechados.
	days := make([]BusinessHourDay, 7)
	for i := range days {
		days[i] = BusinessHourDay{DayOfWeek: i, Closed: true}
	}

	// Para cada dia, calcula MIN open_time e MAX close_time entre membros com closed=false.
	for _, h := range hours {
		if h.Closed || h.DayOfWeek < 0 || h.DayOfWeek > 6 {
			continue
		}
		day := &days[h.DayOfWeek]
		day.Closed = false

		if h.OpenTime != nil {
			if day.OpenTime == nil || *h.OpenTime < *day.OpenTime {
				v := *h.OpenTime
				day.OpenTime = &v
			}
		}
		if h.CloseTime != nil {
			if day.CloseTime == nil || *h.CloseTime > *day.CloseTime {
				v := *h.CloseTime
				day.CloseTime = &v
			}
		}
	}

	// Garante que dias fechados não exponham horários.
	for i := range days {
		if days[i].Closed {
			days[i].OpenTime = nil
			days[i].CloseTime = nil
		}
	}

	return days
}
