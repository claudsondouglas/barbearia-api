// Package models define as estruturas de dados compartilhadas entre as camadas da aplicação.
package models

import "time"

// User representa um usuário cadastrado na barbearia.
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Email     string    `gorm:"not null;uniqueIndex" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	Role      string    `gorm:"not null;default:user" json:"role"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

// CreateUserInput contém os campos obrigatórios para criar um novo usuário.
type CreateUserInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// UpdateUserInput contém os campos opcionais para atualizar um usuário existente.
// Apenas os campos não-nulos são aplicados.
type UpdateUserInput struct {
	Name     *string `json:"name" binding:"omitempty,min=1"`
	Email    *string `json:"email" binding:"omitempty,email"`
	Password *string `json:"password" binding:"omitempty,min=6"`
}
