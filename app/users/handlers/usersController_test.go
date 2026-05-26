package usershttp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"barbearia-api/app"
	"barbearia-api/app/middleware"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestDB cria um *gorm.DB com sqlmock.
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

// withAuthUser injeta AuthUser no contexto Gin sem passar pelo middleware real.
func withAuthUser(userID uint, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("auth_user", middleware.AuthUser{ID: userID, Role: role})
		c.Next()
	}
}

// requireAdmin rejeita quem não é admin, replicando o comportamento do middleware.
func requireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		u, ok := middleware.GetAuthUser(c)
		if !ok || u.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}

// setupRouter cria um roteador de teste sem auth (para testar 401).
func setupRouter(db *gorm.DB) *gin.Engine {
	h := &Handler{Handler: app.NewHandler(db)}
	r := gin.New()
	// rotas sem auth — para verificar que o handler responde 401 quando auth_user ausente
	r.GET("/users", func(c *gin.Context) {
		if _, ok := middleware.GetAuthUser(c); !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		h.List(c)
	})
	r.GET("/users/:id", func(c *gin.Context) {
		if _, ok := middleware.GetAuthUser(c); !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		h.Find(c)
	})
	r.POST("/users", func(c *gin.Context) {
		if _, ok := middleware.GetAuthUser(c); !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		h.Create(c)
	})
	r.PATCH("/users/:id", func(c *gin.Context) {
		if _, ok := middleware.GetAuthUser(c); !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		h.Update(c)
	})
	r.DELETE("/users/:id", func(c *gin.Context) {
		if _, ok := middleware.GetAuthUser(c); !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		h.Delete(c)
	})
	return r
}

// setupAdminRouter cria um roteador com middleware de admin injetado.
func setupAdminRouter(db *gorm.DB, userID uint, role string) *gin.Engine {
	h := &Handler{Handler: app.NewHandler(db)}
	r := gin.New()
	g := r.Group("/users")
	g.Use(withAuthUser(userID, role), requireAdmin())
	g.GET("", h.List)
	g.GET("/:id", h.Find)
	g.POST("", h.Create)
	g.PATCH("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
	return r
}

// makeRequest cria um request com body JSON.
func makeRequest(method, path string, body interface{}) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// --------------------------------------------------------------------------
// List
// --------------------------------------------------------------------------

// TestListUsers_Unauthenticated verifica que GET /users sem auth retorna 401.
func TestListUsers_Unauthenticated(t *testing.T) {
	db, _ := newTestDB(t)
	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestListUsers_ForbiddenForUser verifica que role=user recebe 403.
func TestListUsers_ForbiddenForUser(t *testing.T) {
	db, _ := newTestDB(t)
	r := setupAdminRouter(db, 1, "user")

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

// TestListUsers_AdminSuccess verifica que admin recebe 200 com lista de usuários.
func TestListUsers_AdminSuccess(t *testing.T) {
	db, mock := newTestDB(t)
	r := setupAdminRouter(db, 2, "admin")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(1, "Alice", "alice@test.com", "hashed", "user", time.Now(), time.Now()).
			AddRow(2, "Bob", "bob@test.com", "hashed", "admin", time.Now(), time.Now()))

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	// Verifica que password não está na resposta
	body := w.Body.String()
	if bytes.Contains(w.Body.Bytes(), []byte(`"password"`)) {
		t.Errorf("password must not appear in response, got: %s", body)
	}
}

// --------------------------------------------------------------------------
// Find
// --------------------------------------------------------------------------

// TestFindUser_Unauthenticated verifica que GET /users/:id sem auth retorna 401.
func TestFindUser_Unauthenticated(t *testing.T) {
	db, _ := newTestDB(t)
	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestFindUser_ForbiddenForUser verifica que role=user recebe 403.
func TestFindUser_ForbiddenForUser(t *testing.T) {
	db, _ := newTestDB(t)
	r := setupAdminRouter(db, 1, "user")

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

// TestFindUser_AdminSuccess verifica que admin recebe 200 com o usuário.
func TestFindUser_AdminSuccess(t *testing.T) {
	db, mock := newTestDB(t)
	r := setupAdminRouter(db, 2, "admin")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(1, "Alice", "alice@test.com", "hashed", "user", time.Now(), time.Now()))

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestFindUser_NotFound verifica que ID inexistente retorna 404.
func TestFindUser_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := setupAdminRouter(db, 2, "admin")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("999", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}))

	req := httptest.NewRequest(http.MethodGet, "/users/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// --------------------------------------------------------------------------
// Create
// --------------------------------------------------------------------------

// TestCreateUser_Handler_Unauthenticated verifica que POST /users sem auth retorna 401.
func TestCreateUser_Handler_Unauthenticated(t *testing.T) {
	db, _ := newTestDB(t)
	r := setupRouter(db)

	req := makeRequest(http.MethodPost, "/users", map[string]string{
		"name":     "Test",
		"email":    "test@test.com",
		"password": "pass123",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestCreateUser_Handler_ForbiddenForUser verifica que role=user recebe 403.
func TestCreateUser_Handler_ForbiddenForUser(t *testing.T) {
	db, _ := newTestDB(t)
	r := setupAdminRouter(db, 1, "user")

	req := makeRequest(http.MethodPost, "/users", map[string]string{
		"name":     "Test",
		"email":    "test@test.com",
		"password": "pass123",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

// TestCreateUser_Handler_Success verifica que admin pode criar usuário e recebe 201.
func TestCreateUser_Handler_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := setupAdminRouter(db, 2, "admin")

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectCommit()

	req := makeRequest(http.MethodPost, "/users", map[string]string{
		"name":     "New User",
		"email":    "new@test.com",
		"password": "pass123",
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestCreateUser_Handler_MissingFields verifica que body sem campos obrigatórios retorna 400.
func TestCreateUser_Handler_MissingFields(t *testing.T) {
	db, _ := newTestDB(t)
	r := setupAdminRouter(db, 2, "admin")

	req := makeRequest(http.MethodPost, "/users", map[string]string{
		"email": "test@test.com",
		// name e password ausentes
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// TestCreateUser_Handler_MalformedJSON verifica que body inválido retorna 400.
func TestCreateUser_Handler_MalformedJSON(t *testing.T) {
	db, _ := newTestDB(t)
	r := setupAdminRouter(db, 2, "admin")

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --------------------------------------------------------------------------
// Update
// --------------------------------------------------------------------------

// TestUpdateUser_Handler_Unauthenticated verifica que PATCH /users/:id sem auth retorna 401.
func TestUpdateUser_Handler_Unauthenticated(t *testing.T) {
	db, _ := newTestDB(t)
	r := setupRouter(db)

	req := makeRequest(http.MethodPatch, "/users/1", map[string]string{"name": "New"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestUpdateUser_Handler_ForbiddenForUser verifica que role=user recebe 403.
func TestUpdateUser_Handler_ForbiddenForUser(t *testing.T) {
	db, _ := newTestDB(t)
	r := setupAdminRouter(db, 1, "user")

	req := makeRequest(http.MethodPatch, "/users/1", map[string]string{"name": "New"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

// TestUpdateUser_Handler_Success verifica que admin pode atualizar usuário e recebe 200.
func TestUpdateUser_Handler_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := setupAdminRouter(db, 2, "admin")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(1, "Old Name", "test@test.com", "hashed", "user", time.Now(), time.Now()))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req := makeRequest(http.MethodPatch, "/users/1", map[string]string{"name": "New Name"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestUpdateUser_Handler_EmptyBody verifica que body sem campos retorna 400.
func TestUpdateUser_Handler_EmptyBody(t *testing.T) {
	db, mock := newTestDB(t)
	r := setupAdminRouter(db, 2, "admin")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(1, "Test", "test@test.com", "hashed", "user", time.Now(), time.Now()))

	req := makeRequest(http.MethodPatch, "/users/1", map[string]string{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestUpdateUser_Handler_NotFound verifica que ID inexistente retorna 404.
func TestUpdateUser_Handler_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := setupAdminRouter(db, 2, "admin")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("999", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}))

	req := makeRequest(http.MethodPatch, "/users/999", map[string]string{"name": "Any"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// --------------------------------------------------------------------------
// Delete
// --------------------------------------------------------------------------

// TestDeleteUser_Handler_Unauthenticated verifica que DELETE /users/:id sem auth retorna 401.
func TestDeleteUser_Handler_Unauthenticated(t *testing.T) {
	db, _ := newTestDB(t)
	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestDeleteUser_Handler_ForbiddenForUser verifica que role=user recebe 403.
func TestDeleteUser_Handler_ForbiddenForUser(t *testing.T) {
	db, _ := newTestDB(t)
	r := setupAdminRouter(db, 1, "user")

	req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

// TestDeleteUser_Handler_Success verifica que admin pode deletar usuário e recebe 204.
func TestDeleteUser_Handler_Success(t *testing.T) {
	db, mock := newTestDB(t)
	r := setupAdminRouter(db, 2, "admin")

	// Find chamado pelo handler Delete
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(1, "Alice", "alice@test.com", "hashed", "user", time.Now(), time.Now()))

	// Delete
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestDeleteUser_Handler_NotFound verifica que ID inexistente retorna 404.
func TestDeleteUser_Handler_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	r := setupAdminRouter(db, 2, "admin")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("999", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}))

	req := httptest.NewRequest(http.MethodDelete, "/users/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d (body: %s)", w.Code, w.Body.String())
	}
}
