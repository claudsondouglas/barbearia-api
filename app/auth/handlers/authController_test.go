package authhttp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
	"time"

	"barbearia-api/app"
	authactions "barbearia-api/app/auth/actions"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestDB cria um *gorm.DB com sqlmock para testes de handler.
func newTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	return gormDB, mock
}

// newTestRedis cria um cliente Redis apontando para miniredis.
func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	mr := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

// setupAuthRouter cria um roteador Gin de teste com todas as rotas de auth.
func setupAuthRouter(t *testing.T) (*gin.Engine, *gorm.DB, sqlmock.Sqlmock, *redis.Client) {
	t.Helper()
	db, mock := newTestDB(t)
	rdb := newTestRedis(t)

	h := &Handler{
		Handler: app.NewHandler(db),
		Redis:   rdb,
	}

	r := gin.New()
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)
	r.GET("/auth/verify", h.Verify)
	r.POST("/auth/refresh", h.Refresh)
	r.POST("/auth/forgot-password", h.ForgotPassword)
	r.POST("/auth/reset-password", h.ResetPassword)
	r.POST("/auth/logout", h.Logout)

	return r, db, mock, rdb
}

// makeRequest é um helper para criar requests com body JSON.
func makeRequest(method, path string, body interface{}) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// testClaims é uma cópia local de authactions.Claims para gerar tokens de teste.
type testClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

// generateTestToken cria um JWT assinado diretamente (sem depender do export_test.go).
func generateTestToken(t *testing.T, userID uint, email, role, tokenType string, duration time.Duration) string {
	t.Helper()
	secret := []byte(os.Getenv("JWT_SECRET"))
	claims := testClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(secret)
	if err != nil {
		t.Fatalf("generateTestToken: %v", err)
	}
	return signed
}

// generateAccessToken gera um access token válido para testes.
func generateAccessToken(t *testing.T, userID uint, email, role string) string {
	t.Helper()
	return generateTestToken(t, userID, email, role, "access", 15*time.Minute)
}

// generateRefreshToken gera um refresh token válido para testes.
func generateRefreshToken(t *testing.T, userID uint, email, role string) string {
	t.Helper()
	return generateTestToken(t, userID, email, role, "refresh", 7*24*time.Hour)
}

// --------------------------------------------------------------------------
// Register
// --------------------------------------------------------------------------

// TestRegister_Success verifica que POST /auth/register com dados válidos retorna 201 com tokens.
func TestRegister_Success(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, mock, _ := setupAuthRouter(t)

	// Register verifica se e-mail já existe primeiro
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("new@test.com", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	// INSERT do novo usuário
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	req := makeRequest(http.MethodPost, "/auth/register", map[string]string{
		"name":     "Test User",
		"email":    "new@test.com",
		"password": "pass123",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d (body: %s)", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["access_token"] == "" || resp["refresh_token"] == "" {
		t.Errorf("expected access_token and refresh_token in response, got: %v", resp)
	}
	if _, ok := resp["password"]; ok {
		t.Error("password must not appear in response")
	}
}

// TestRegister_MissingName verifica que body sem name retorna 400.
func TestRegister_MissingName(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	req := makeRequest(http.MethodPost, "/auth/register", map[string]string{
		"email":    "test@test.com",
		"password": "pass123",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// TestRegister_MissingEmail verifica que body sem email retorna 400.
func TestRegister_MissingEmail(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	req := makeRequest(http.MethodPost, "/auth/register", map[string]string{
		"name":     "Test",
		"password": "pass123",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// TestRegister_MissingPassword verifica que body sem password retorna 400.
func TestRegister_MissingPassword(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	req := makeRequest(http.MethodPost, "/auth/register", map[string]string{
		"name":  "Test",
		"email": "test@test.com",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// TestRegister_ShortPassword verifica que senha com menos de 6 caracteres retorna 400.
func TestRegister_ShortPassword(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	req := makeRequest(http.MethodPost, "/auth/register", map[string]string{
		"name":     "Test",
		"email":    "test@test.com",
		"password": "12345",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for short password, got %d", w.Code)
	}
}

// TestRegister_DuplicateEmail verifica que e-mail já cadastrado retorna 409.
func TestRegister_DuplicateEmail(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, mock, _ := setupAuthRouter(t)

	// Retorna um usuário existente → ErrEmailAlreadyInUse
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("existing@test.com", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(1, "Existing", "existing@test.com", "hashed", "user", time.Now(), time.Now()))

	req := makeRequest(http.MethodPost, "/auth/register", map[string]string{
		"name":     "Test",
		"email":    "existing@test.com",
		"password": "pass123",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestRegister_MalformedJSON verifica que body inválido retorna 400.
func TestRegister_MalformedJSON(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --------------------------------------------------------------------------
// Login
// --------------------------------------------------------------------------

// TestLogin_Handler_Success verifica que POST /auth/login com credenciais válidas retorna 200.
func TestLogin_Handler_Success(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, mock, _ := setupAuthRouter(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.MinCost)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("user@test.com", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(1, "Test", "user@test.com", string(hash), "user", time.Now(), time.Now()))

	req := makeRequest(http.MethodPost, "/auth/login", map[string]string{
		"email":    "user@test.com",
		"password": "pass123",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestLogin_Handler_WrongPassword verifica que senha errada retorna 401.
func TestLogin_Handler_WrongPassword(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, mock, _ := setupAuthRouter(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("user@test.com", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(1, "Test", "user@test.com", string(hash), "user", time.Now(), time.Now()))

	req := makeRequest(http.MethodPost, "/auth/login", map[string]string{
		"email":    "user@test.com",
		"password": "wrong",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestLogin_Handler_MissingEmail verifica que body sem email retorna 400.
func TestLogin_Handler_MissingEmail(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	req := makeRequest(http.MethodPost, "/auth/login", map[string]string{
		"password": "pass123",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// TestLogin_Handler_MissingPassword verifica que body sem password retorna 400.
func TestLogin_Handler_MissingPassword(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	req := makeRequest(http.MethodPost, "/auth/login", map[string]string{
		"email": "user@test.com",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// TestLogin_Handler_MalformedJSON verifica que body inválido retorna 400.
func TestLogin_Handler_MalformedJSON(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --------------------------------------------------------------------------
// Verify
// --------------------------------------------------------------------------

// TestVerify_Handler_ValidToken verifica que GET /auth/verify com token válido retorna 200.
func TestVerify_Handler_ValidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	token := generateAccessToken(t, 1, "user@test.com", "user")

	req := httptest.NewRequest(http.MethodGet, "/auth/verify", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["user_id"] == nil || resp["email"] == nil || resp["role"] == nil {
		t.Errorf("expected user_id, email, role in response, got: %v", resp)
	}
}

// TestVerify_Handler_MissingHeader verifica que ausência do header Authorization retorna 401.
func TestVerify_Handler_MissingHeader(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/auth/verify", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestVerify_Handler_WrongPrefix verifica que header com prefixo errado (Token em vez de Bearer) retorna 401.
func TestVerify_Handler_WrongPrefix(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	token := generateAccessToken(t, 1, "user@test.com", "user")

	req := httptest.NewRequest(http.MethodGet, "/auth/verify", nil)
	req.Header.Set("Authorization", "Token "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestVerify_Handler_BlacklistedToken verifica que token revogado retorna 401 com "token revoked".
func TestVerify_Handler_BlacklistedToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, rdb := setupAuthRouter(t)

	token := generateAccessToken(t, 1, "user@test.com", "user")

	// Revogar o token via Logout
	if err := authactions.Logout(httptest.NewRequest(http.MethodGet, "/", nil).Context(), rdb, token); err != nil {
		t.Fatalf("logout: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/verify", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["error"] != "token revoked" {
		t.Errorf("expected 'token revoked', got: %s", resp["error"])
	}
}

// --------------------------------------------------------------------------
// Refresh
// --------------------------------------------------------------------------

// TestRefresh_Handler_Success verifica que POST /auth/refresh com refresh token válido retorna 200.
func TestRefresh_Handler_Success(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	refreshToken := generateRefreshToken(t, 1, "user@test.com", "user")

	req := makeRequest(http.MethodPost, "/auth/refresh", map[string]string{
		"refresh_token": refreshToken,
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["access_token"] == "" {
		t.Error("expected access_token in response")
	}
}

// TestRefresh_Handler_InvalidToken verifica que token inválido retorna 401.
func TestRefresh_Handler_InvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	req := makeRequest(http.MethodPost, "/auth/refresh", map[string]string{
		"refresh_token": "invalid.token.here",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestRefresh_Handler_MissingBody verifica que body sem refresh_token retorna 400.
func TestRefresh_Handler_MissingBody(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	req := makeRequest(http.MethodPost, "/auth/refresh", map[string]string{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --------------------------------------------------------------------------
// ForgotPassword
// --------------------------------------------------------------------------

// TestForgotPassword_Handler_AlwaysReturns200 verifica que mesmo com e-mail não cadastrado retorna 200.
// O handler retorna 200 imediatamente quando o mail client falha (sem env RESEND_API_KEY).
func TestForgotPassword_Handler_AlwaysReturns200(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	req := makeRequest(http.MethodPost, "/auth/forgot-password", map[string]string{
		"email": "notexist@test.com",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// O handler sempre retorna 200 independente do e-mail existir (anti-enumeração)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestForgotPassword_Handler_InvalidEmail verifica que body com e-mail mal formatado retorna 400.
func TestForgotPassword_Handler_InvalidEmail(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	req := makeRequest(http.MethodPost, "/auth/forgot-password", map[string]string{
		"email": "not-an-email",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// TestForgotPassword_Handler_MissingEmail verifica que body sem email retorna 400.
func TestForgotPassword_Handler_MissingEmail(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	req := makeRequest(http.MethodPost, "/auth/forgot-password", map[string]string{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --------------------------------------------------------------------------
// ResetPassword
// --------------------------------------------------------------------------

// TestResetPassword_Handler_MissingCode verifica que body sem code retorna 400.
func TestResetPassword_Handler_MissingCode(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	req := makeRequest(http.MethodPost, "/auth/reset-password", map[string]string{
		"new_password": "newpass123",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// TestResetPassword_Handler_MissingPassword verifica que body sem new_password retorna 400.
func TestResetPassword_Handler_MissingPassword(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	req := makeRequest(http.MethodPost, "/auth/reset-password", map[string]string{
		"code": "123456",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// TestResetPassword_Handler_InvalidOTP verifica que OTP inválido retorna 400 (não 500).
func TestResetPassword_Handler_InvalidOTP(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, mock, _ := setupAuthRouter(t)

	// OTP não encontrado → ErrInvalidOrExpiredOTP → 400
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "password_reset_otps"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "code", "used", "expires_at"}))

	req := makeRequest(http.MethodPost, "/auth/reset-password", map[string]string{
		"code":         "000000",
		"new_password": "newpass123",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// --------------------------------------------------------------------------
// Logout
// --------------------------------------------------------------------------

// TestLogout_Handler_Success verifica que POST /auth/logout com token válido retorna 200.
func TestLogout_Handler_Success(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	token := generateAccessToken(t, 1, "user@test.com", "user")

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestLogout_Handler_MissingHeader verifica que ausência do Authorization retorna 401.
func TestLogout_Handler_MissingHeader(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestLogout_Handler_InvalidToken verifica que token malformado retorna 401.
func TestLogout_Handler_InvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r, _, _, _ := setupAuthRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
