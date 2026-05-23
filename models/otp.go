package models

import "time"

// PasswordResetOTP armazena um código OTP para redefinição de senha.
// O código tem validade de 15 minutos e é invalidado após o uso.
type PasswordResetOTP struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"not null;index"`
	Code      string    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"not null;default:false"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}
