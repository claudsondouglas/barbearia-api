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
	panic("not implemented")
}

// CreateOrgException cria uma exceção de horário no nível da organização.
// Se closed=true, deleta em transação todas as exceções de membros para a mesma data.
// Retorna (exception, deletedMemberExceptionsCount, error).
func CreateOrgException(db *gorm.DB, orgSlug string, requestingUserID uint, input models.CreateExceptionInput) (*models.ScheduleException, int, error) {
	panic("not implemented")
}

// DeleteOrgException deleta uma exceção de horário no nível da organização.
// Verifica que a exceção pertence à org e que user_id IS NULL.
// Retorna ErrExceptionNotFound se não encontrar.
func DeleteOrgException(db *gorm.DB, orgSlug string, requestingUserID, exceptionID uint) error {
	panic("not implemented")
}

// ListMemberExceptions retorna as exceções de horário de um membro específico.
// Suporta paginação (limit/offset) e filtro de intervalo de datas (from, to).
// Requer que requestingUserID seja owner ou admin.
func ListMemberExceptions(db *gorm.DB, orgSlug string, requestingUserID, targetUserID uint, limit, offset int, from, to *string) ([]models.ScheduleException, error) {
	panic("not implemented")
}

// CreateMemberException cria uma exceção de horário para um membro específico.
// Valida que a org não está fechada na data e que a janela do membro está dentro da janela da org.
// Retorna ErrOrgClosedOnDate, ErrMemberWindowExceedsOrg ou ErrDuplicateException conforme o caso.
func CreateMemberException(db *gorm.DB, orgSlug string, requestingUserID, targetUserID uint, input models.CreateExceptionInput) (*models.ScheduleException, error) {
	panic("not implemented")
}

// DeleteMemberException deleta uma exceção de horário de um membro específico.
// Verifica que a exceção pertence ao membro correto na org.
// Retorna ErrExceptionNotFound se não encontrar.
func DeleteMemberException(db *gorm.DB, orgSlug string, requestingUserID, targetUserID, exceptionID uint) error {
	panic("not implemented")
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
