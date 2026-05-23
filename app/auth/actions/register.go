package authactions

import (
	"errors"
	"time"

	"barbearia-api/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ErrEmailAlreadyInUse é retornado quando o e-mail informado já está cadastrado.
var ErrEmailAlreadyInUse = errors.New("email already in use")

// Register cria um novo usuário e retorna um par de tokens JWT (access + refresh).
// Retorna ErrEmailAlreadyInUse se o e-mail já estiver cadastrado.
func Register(db *gorm.DB, input models.CreateUserInput) (*models.TokenResponse, error) {
	var existing models.User
	if result := db.Where("email = ?", input.Email).First(&existing); result.Error == nil {
		return nil, ErrEmailAlreadyInUse
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hash),
		Role:     "user",
	}

	if result := db.Create(&user); result.Error != nil {
		return nil, result.Error
	}

	accessToken, err := generateToken(user.ID, user.Email, user.Role, "access", 15*time.Minute)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateToken(user.ID, user.Email, user.Role, "refresh", 7*24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
