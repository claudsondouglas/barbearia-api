package actions

import (
	"errors"
	"fmt"

	"barbearia-api/models"

	"gorm.io/gorm"
)

var ErrMemberNotFound = errors.New("member not found")
var ErrInvalidBusinessHour = errors.New("invalid business hour")

// UpdateBusinessHourInput contém os campos para atualizar um dia de horário de membro.
type UpdateBusinessHourInput struct {
	DayOfWeek int     `json:"day_of_week" binding:"min=0,max=6"`
	OpenTime  *string `json:"open_time"`
	CloseTime *string `json:"close_time"`
	Closed    bool    `json:"closed"`
}

// GetMemberBusinessHours retorna os 7 dias de horário de um membro específico na org.
// Retorna ErrOrgNotFound ou ErrMemberNotFound conforme o caso.
func GetMemberBusinessHours(db *gorm.DB, orgSlug string, userID uint) ([]models.MemberBusinessHour, error) {
	panic("not implemented")
}

// UpdateMemberBusinessHoursBatch atualiza em lote os dias informados nos updates.
// Apenas owner ou admin (requestingUserID) pode atualizar horários de targetUserID.
// Valida cada entrada com ValidateBusinessHour antes de persistir.
func UpdateMemberBusinessHoursBatch(db *gorm.DB, orgSlug string, requestingUserID, targetUserID uint, updates []UpdateBusinessHourInput) error {
	panic("not implemented")
}

// ValidateBusinessHour valida as regras de negócio de um horário:
//   - dayOfWeek deve estar entre 0 e 6.
//   - Se closed=true, openTime e closeTime devem ser nil.
//   - Se closed=false, openTime e closeTime são obrigatórios.
//   - closeTime deve ser posterior a openTime (sem overnight).
func ValidateBusinessHour(dayOfWeek int, openTime, closeTime *string, closed bool) error {
	if dayOfWeek < 0 || dayOfWeek > 6 {
		return fmt.Errorf("%w: day_of_week must be between 0 and 6", ErrInvalidBusinessHour)
	}

	if closed {
		if openTime != nil || closeTime != nil {
			return fmt.Errorf("%w: open_time and close_time must be null when closed=true", ErrInvalidBusinessHour)
		}
		return nil
	}

	// closed=false: ambos os horários são obrigatórios.
	if openTime == nil {
		return fmt.Errorf("%w: open_time is required when closed=false", ErrInvalidBusinessHour)
	}
	if closeTime == nil {
		return fmt.Errorf("%w: close_time is required when closed=false", ErrInvalidBusinessHour)
	}

	// closeTime deve ser estritamente posterior a openTime (sem overnight).
	if *closeTime <= *openTime {
		return fmt.Errorf("%w: close_time must be after open_time and overnight schedules are not supported", ErrInvalidBusinessHour)
	}

	return nil
}
