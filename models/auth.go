package models

// LoginInput contém as credenciais necessárias para autenticação.
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse é retornado após uma autenticação bem-sucedida.
// AccessToken expira em 15 minutos; RefreshToken expira em 7 dias.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
