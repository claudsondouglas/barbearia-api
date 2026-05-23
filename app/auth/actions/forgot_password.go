package authactions

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"barbearia-api/models"

	"gorm.io/gorm"
)

// Sender é a interface necessária para enviar o código OTP ao usuário.
// Permite substituir o cliente de e-mail real por um mock em testes.
type Sender interface {
	SendEmail(to, subject, body string) error
}

// ForgotPassword gera um OTP de seis dígitos, remove OTPs anteriores do e-mail,
// persiste o novo código com validade de 15 minutos e o envia por e-mail.
// Retorna nil silenciosamente quando o e-mail não está cadastrado,
// evitando a enumeração de usuários.
func ForgotPassword(db *gorm.DB, email string, sender Sender) error {
	var user models.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil
	}

	db.Where("email = ?", email).Delete(&models.PasswordResetOTP{})

	code, err := generateOTP()
	if err != nil {
		return fmt.Errorf("generate otp: %w", err)
	}

	otp := models.PasswordResetOTP{
		Email:     email,
		Code:      code,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	if err := db.Create(&otp).Error; err != nil {
		return fmt.Errorf("save otp: %w", err)
	}

	body := fmt.Sprintf("Seu código de redefinição de senha: %s\nExpira em 15 minutos.", code)
	if err := sender.SendEmail(email, "Redefinição de senha", body); err != nil {
		return fmt.Errorf("send otp: %w", err)
	}

	return nil
}

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
