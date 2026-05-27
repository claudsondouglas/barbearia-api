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
	org, err := FindBySlug(db, orgSlug)
	if err != nil {
		return nil, err
	}

	member, err := findActiveMember(db, org.ID, userID)
	if err != nil {
		return nil, err
	}

	var hours []models.MemberBusinessHour
	if err := db.Where("org_member_id = ?", member.ID).Order("day_of_week").Find(&hours).Error; err != nil {
		return nil, err
	}

	return hours, nil
}

// UpdateMemberBusinessHoursBatch atualiza em lote os dias informados nos updates.
// Apenas owner ou admin (requestingUserID) pode atualizar horários de targetUserID.
// Valida cada entrada com ValidateBusinessHour antes de persistir.
func UpdateMemberBusinessHoursBatch(db *gorm.DB, orgSlug string, requestingUserID, targetUserID uint, updates []UpdateBusinessHourInput) error {
	if len(updates) == 0 {
		return errors.New("no updates provided")
	}

	for _, u := range updates {
		if err := ValidateBusinessHour(u.DayOfWeek, u.OpenTime, u.CloseTime, u.Closed); err != nil {
			return err
		}
	}

	org, err := FindBySlug(db, orgSlug)
	if err != nil {
		return err
	}

	if err := checkOwnerOrAdmin(db, org, requestingUserID); err != nil {
		return err
	}

	member, err := findActiveMember(db, org.ID, targetUserID)
	if err != nil {
		return err
	}

	for _, u := range updates {
		var openTime, closeTime *string
		if !u.Closed {
			openTime = u.OpenTime
			closeTime = u.CloseTime
		}
		if err := db.Model(&models.MemberBusinessHour{}).
			Where("org_member_id = ? AND day_of_week = ?", member.ID, u.DayOfWeek).
			Updates(map[string]interface{}{
				"closed":     u.Closed,
				"open_time":  openTime,
				"close_time": closeTime,
			}).Error; err != nil {
			return err
		}
	}

	return nil
}

// UpdateMemberBusinessHourDay atualiza um único dia de horário de um membro.
// Apenas owner ou admin (requestingUserID) pode atualizar horários de targetUserID.
func UpdateMemberBusinessHourDay(db *gorm.DB, orgSlug string, requestingUserID, targetUserID uint, day int, input UpdateBusinessHourInput) error {
	input.DayOfWeek = day
	return UpdateMemberBusinessHoursBatch(db, orgSlug, requestingUserID, targetUserID, []UpdateBusinessHourInput{input})
}

func findActiveMember(db *gorm.DB, orgID, userID uint) (models.OrgMember, error) {
	var members []models.OrgMember
	if err := db.Where("organization_id = ? AND user_id = ?", orgID, userID).Find(&members).Error; err != nil {
		return models.OrgMember{}, err
	}
	if len(members) == 0 {
		return models.OrgMember{}, ErrMemberNotFound
	}
	return members[0], nil
}

func checkOwnerOrAdmin(db *gorm.DB, org *models.Organization, requestingUserID uint) error {
	if org.OwnerID == requestingUserID {
		return nil
	}
	var users []models.User
	if err := db.Where("id = ?", requestingUserID).Find(&users).Error; err != nil || len(users) == 0 {
		return ErrForbidden
	}
	if users[0].Role == "admin" {
		return nil
	}
	return ErrForbidden
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
