package actions

import (
	"errors"
	"fmt"

	"barbearia-api/models"

	"gorm.io/gorm"
)

var ErrExceptionNotFound = errors.New("exception not found")
var ErrDuplicateException = errors.New("exception already exists for this date")
var ErrOrgClosedOnDate = errors.New("organization is closed on this date")
var ErrMemberWindowExceedsOrg = errors.New("member window exceeds org window")

// ListOrgExceptions retorna as exceções de horário no nível da organização (user_id IS NULL).
// Suporta paginação (limit/offset) e filtro de intervalo de datas (from, to).
// Requer que requestingUserID seja owner ou admin.
func ListOrgExceptions(db *gorm.DB, orgSlug string, requestingUserID uint, limit, offset int, from, to *string) ([]models.ScheduleException, error) {
	org, err := FindBySlug(db, orgSlug)
	if err != nil {
		return nil, err
	}

	if err := checkOwnerOrAdmin(db, org, requestingUserID); err != nil {
		return nil, err
	}

	q := db.Where("organization_id = ? AND user_id IS NULL", org.ID).
		Order("date ASC").
		Limit(limit).
		Offset(offset)

	if from != nil {
		q = q.Where("date >= ?", *from)
	}
	if to != nil {
		q = q.Where("date <= ?", *to)
	}

	var exceptions []models.ScheduleException
	if err := q.Find(&exceptions).Error; err != nil {
		return nil, err
	}
	return exceptions, nil
}

// CreateOrgException cria uma exceção de horário no nível da organização.
// Se closed=true, deleta em transação todas as exceções de membros para a mesma data.
// Retorna (exception, deletedMemberExceptionsCount, error).
func CreateOrgException(db *gorm.DB, orgSlug string, requestingUserID uint, input models.CreateExceptionInput) (*models.ScheduleException, int, error) {
	org, err := FindBySlug(db, orgSlug)
	if err != nil {
		return nil, 0, err
	}

	if err := checkOwnerOrAdmin(db, org, requestingUserID); err != nil {
		return nil, 0, err
	}

	if err := ValidateException(input.Closed, input.OpenTime, input.CloseTime); err != nil {
		return nil, 0, err
	}

	var existing models.ScheduleException
	err = db.Where("organization_id = ? AND user_id IS NULL AND date = ?", org.ID, input.Date).First(&existing).Error
	if err == nil {
		return nil, 0, ErrDuplicateException
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, 0, err
	}

	exception := models.ScheduleException{
		OrganizationID: org.ID,
		Date:           input.Date,
		Closed:         input.Closed,
		OpenTime:       input.OpenTime,
		CloseTime:      input.CloseTime,
		Reason:         input.Reason,
	}

	deletedCount := 0

	if input.Closed {
		txErr := db.Transaction(func(tx *gorm.DB) error {
			var count int64
			if err := tx.Model(&models.ScheduleException{}).
				Where("organization_id = ? AND user_id IS NOT NULL AND date = ?", org.ID, input.Date).
				Count(&count).Error; err != nil {
				return err
			}
			deletedCount = int(count)

			if count > 0 {
				if err := tx.Where("organization_id = ? AND user_id IS NOT NULL AND date = ?", org.ID, input.Date).
					Delete(&models.ScheduleException{}).Error; err != nil {
					return err
				}
			}

			return tx.Create(&exception).Error
		})
		if txErr != nil {
			return nil, 0, txErr
		}
	} else {
		if err := db.Create(&exception).Error; err != nil {
			return nil, 0, err
		}
	}

	return &exception, deletedCount, nil
}

// DeleteOrgException deleta uma exceção de horário no nível da organização.
// Verifica que a exceção pertence à org e que user_id IS NULL.
// Retorna ErrExceptionNotFound se não encontrar.
func DeleteOrgException(db *gorm.DB, orgSlug string, requestingUserID, exceptionID uint) error {
	org, err := FindBySlug(db, orgSlug)
	if err != nil {
		return err
	}

	if err := checkOwnerOrAdmin(db, org, requestingUserID); err != nil {
		return err
	}

	var exception models.ScheduleException
	if err := db.Where("id = ? AND organization_id = ? AND user_id IS NULL", exceptionID, org.ID).First(&exception).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrExceptionNotFound
		}
		return err
	}

	return db.Delete(&exception).Error
}

// ListMemberExceptions retorna as exceções de horário de um membro específico.
// Suporta paginação (limit/offset) e filtro de intervalo de datas (from, to).
// Requer que requestingUserID seja owner ou admin.
func ListMemberExceptions(db *gorm.DB, orgSlug string, requestingUserID, targetUserID uint, limit, offset int, from, to *string) ([]models.ScheduleException, error) {
	org, err := FindBySlug(db, orgSlug)
	if err != nil {
		return nil, err
	}

	if err := checkOwnerOrAdmin(db, org, requestingUserID); err != nil {
		return nil, err
	}

	q := db.Where("organization_id = ? AND user_id = ?", org.ID, targetUserID).
		Order("date ASC").
		Limit(limit).
		Offset(offset)

	if from != nil {
		q = q.Where("date >= ?", *from)
	}
	if to != nil {
		q = q.Where("date <= ?", *to)
	}

	var exceptions []models.ScheduleException
	if err := q.Find(&exceptions).Error; err != nil {
		return nil, err
	}
	return exceptions, nil
}

// CreateMemberException cria uma exceção de horário para um membro específico.
// Valida que a org não está fechada na data e que a janela do membro está dentro da janela da org.
// Retorna ErrOrgClosedOnDate, ErrMemberWindowExceedsOrg ou ErrDuplicateException conforme o caso.
func CreateMemberException(db *gorm.DB, orgSlug string, requestingUserID, targetUserID uint, input models.CreateExceptionInput) (*models.ScheduleException, error) {
	org, err := FindBySlug(db, orgSlug)
	if err != nil {
		return nil, err
	}

	if err := checkOwnerOrAdmin(db, org, requestingUserID); err != nil {
		return nil, err
	}

	if err := ValidateException(input.Closed, input.OpenTime, input.CloseTime); err != nil {
		return nil, err
	}

	var orgException models.ScheduleException
	orgExcErr := db.Where("organization_id = ? AND user_id IS NULL AND date = ?", org.ID, input.Date).First(&orgException).Error
	if orgExcErr != nil && !errors.Is(orgExcErr, gorm.ErrRecordNotFound) {
		return nil, orgExcErr
	}

	orgExceptionFound := orgExcErr == nil

	if orgExceptionFound && orgException.Closed {
		return nil, ErrOrgClosedOnDate
	}

	if !input.Closed && orgExceptionFound && !orgException.Closed {
		if orgException.OpenTime == nil || orgException.CloseTime == nil {
			return nil, fmt.Errorf("%w: org exception has no open/close times", ErrMemberWindowExceedsOrg)
		}
		if input.OpenTime == nil || input.CloseTime == nil {
			return nil, fmt.Errorf("%w: member exception has no open/close times", ErrMemberWindowExceedsOrg)
		}
		if err := ValidateMemberExceptionAgainstOrg(*input.OpenTime, *input.CloseTime, *orgException.OpenTime, *orgException.CloseTime); err != nil {
			return nil, err
		}
	}

	var existing models.ScheduleException
	err = db.Where("organization_id = ? AND user_id = ? AND date = ?", org.ID, targetUserID, input.Date).First(&existing).Error
	if err == nil {
		return nil, ErrDuplicateException
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	exception := models.ScheduleException{
		OrganizationID: org.ID,
		UserID:         &targetUserID,
		Date:           input.Date,
		Closed:         input.Closed,
		OpenTime:       input.OpenTime,
		CloseTime:      input.CloseTime,
		Reason:         input.Reason,
	}

	if err := db.Create(&exception).Error; err != nil {
		return nil, err
	}

	return &exception, nil
}

// DeleteMemberException deleta uma exceção de horário de um membro específico.
// Verifica que a exceção pertence ao membro correto na org.
// Retorna ErrExceptionNotFound se não encontrar.
func DeleteMemberException(db *gorm.DB, orgSlug string, requestingUserID, targetUserID, exceptionID uint) error {
	org, err := FindBySlug(db, orgSlug)
	if err != nil {
		return err
	}

	if err := checkOwnerOrAdmin(db, org, requestingUserID); err != nil {
		return err
	}

	var exception models.ScheduleException
	if err := db.Where("id = ? AND organization_id = ? AND user_id = ?", exceptionID, org.ID, targetUserID).First(&exception).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrExceptionNotFound
		}
		return err
	}

	return db.Delete(&exception).Error
}

// ValidateException valida as regras de negócio de uma exceção de horário:
//   - Se closed=true, openTime e closeTime devem ser nil.
//   - Se closed=false, openTime e closeTime são obrigatórios.
//   - closeTime deve ser estritamente posterior a openTime (sem overnight).
func ValidateException(closed bool, openTime, closeTime *string) error {
	if closed {
		if openTime != nil || closeTime != nil {
			return fmt.Errorf("%w: open_time and close_time must be null when closed=true", ErrInvalidBusinessHour)
		}
		return nil
	}

	if openTime == nil {
		return fmt.Errorf("%w: open_time is required when closed=false", ErrInvalidBusinessHour)
	}
	if closeTime == nil {
		return fmt.Errorf("%w: close_time is required when closed=false", ErrInvalidBusinessHour)
	}

	if *closeTime <= *openTime {
		return fmt.Errorf("%w: close_time must be after open_time and overnight schedules are not supported", ErrInvalidBusinessHour)
	}

	return nil
}

// ValidateMemberExceptionAgainstOrg verifica que a janela do membro está contida na janela da org.
// Retorna ErrMemberWindowExceedsOrg se o membro abrir antes ou fechar depois da org.
func ValidateMemberExceptionAgainstOrg(memberOpen, memberClose, orgOpen, orgClose string) error {
	if memberOpen < orgOpen {
		return fmt.Errorf("%w: member open_time (%s) is before org open_time (%s)", ErrMemberWindowExceedsOrg, memberOpen, orgOpen)
	}
	if memberClose > orgClose {
		return fmt.Errorf("%w: member close_time (%s) is after org close_time (%s)", ErrMemberWindowExceedsOrg, memberClose, orgClose)
	}
	return nil
}
