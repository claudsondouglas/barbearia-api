package authactions

import (
	"errors"
	"fmt"
	"time"

	"barbearia-api/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ErrInvalidOrExpiredOTP é retornado quando o código OTP não é encontrado,
// já foi utilizado ou o prazo de validade expirou.
var ErrInvalidOrExpiredOTP = errors.New("invalid or expired OTP")

// ResetPasswordInput contém os dados necessários para redefinir a senha via OTP.
type ResetPasswordInput struct {
	Code        string `json:"code"         binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ResetPassword valida o OTP informado, atualiza a senha do usuário com bcrypt
// e marca o código como utilizado para impedir reutilização.
// Retorna ErrInvalidOrExpiredOTP se o código for inválido, já usado ou expirado.
func ResetPassword(db *gorm.DB, input ResetPasswordInput) error {
	var otp models.PasswordResetOTP
	err := db.Where(
		"code = ? AND used = false AND expires_at > ?",
		input.Code, time.Now(),
	).First(&otp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidOrExpiredOTP
		}
		return fmt.Errorf("find otp: %w", err)
	}

	var user models.User
	if err := db.Where("email = ?", otp.Email).First(&user).Error; err != nil {
		return fmt.Errorf("find user: %w", err)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := db.Model(&user).Update("password", string(hashed)).Error; err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	if err := db.Model(&otp).Update("used", true).Error; err != nil {
		return fmt.Errorf("invalidate otp: %w", err)
	}

	return nil
}
