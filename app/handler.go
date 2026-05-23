// Package app fornece o handler base compartilhado entre todos os módulos HTTP.
package app

import "gorm.io/gorm"

// Handler é o handler base que carrega a conexão com o banco de dados.
// Os handlers de cada módulo embute este tipo para ter acesso ao DB.
type Handler struct {
	DB *gorm.DB
}

// NewHandler cria um novo Handler com a conexão de banco fornecida.
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{DB: db}
}
