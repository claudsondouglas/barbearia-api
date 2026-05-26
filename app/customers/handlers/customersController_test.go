package customershttp

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"barbearia-api/app"
	"barbearia-api/app/middleware"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newTestDB cria um *gorm.DB com sqlmock para uso nos handler tests.
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

// withAuthUser injeta um AuthUser no contexto Gin sem passar pelo middleware real.
func withAuthUser(userID uint, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("auth_user", middleware.AuthUser{ID: userID, Role: role})
		c.Next()
	}
}

// setupRouter cria um roteador Gin de teste com as 5 rotas de customers.
func setupRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/organizations/:slug/customers", h.Create)
	r.GET("/organizations/:slug/customers", h.List)
	r.GET("/organizations/:slug/customers/:id", h.Find)
	r.PATCH("/organizations/:slug/customers/:id", h.Update)
	r.DELETE("/organizations/:slug/customers/:id", h.Delete)
	return r
}

// setupRouterWithAuth cria um roteador com middleware de auth injetado.
func setupRouterWithAuth(h *Handler, userID uint, role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	auth := withAuthUser(userID, role)
	r.POST("/organizations/:slug/customers", auth, h.Create)
	r.GET("/organizations/:slug/customers", auth, h.List)
	r.GET("/organizations/:slug/customers/:id", auth, h.Find)
	r.PATCH("/organizations/:slug/customers/:id", auth, h.Update)
	r.DELETE("/organizations/:slug/customers/:id", auth, h.Delete)
	return r
}

// --- Create ---

// TestCreate_201 verifica que owner cria customer e recebe 201.
func TestCreate_201(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 10, "user")

	mock.ExpectQuery(`SELECT.*organizations`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "name", "phone", "email", "street", "number", "neighborhood", "city", "state", "zip_code", "timezone", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, "Org", "11999990000", "org@test.com", "Rua A", "1", "Bairro", "Cidade", "SP", "00000-000", "America/Sao_Paulo", nil, nil, nil))
	mock.ExpectQuery(`SELECT.*customers`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "customers"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	body := bytes.NewBufferString(`{"name":"Carlos Souza","phone":"11999990004"}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations/barbearia-test/customers", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestCreate_401 verifica que request sem auth retorna 401.
func TestCreate_401(t *testing.T) {
	db, _ := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(h)

	body := bytes.NewBufferString(`{"name":"Carlos","phone":"11999990004"}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations/barbearia-test/customers", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestCreate_403_RegularUser verifica que usuário que não é owner recebe 403.
func TestCreate_403_RegularUser(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	// userID=20, role="user"; org tem owner_id=10 → não é owner → 403
	r := setupRouterWithAuth(h, 20, "user")

	mock.ExpectQuery(`SELECT.*organizations`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "name", "phone", "email", "street", "number", "neighborhood", "city", "state", "zip_code", "timezone", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, "Org", "11999990000", "org@test.com", "Rua A", "1", "Bairro", "Cidade", "SP", "00000-000", "America/Sao_Paulo", nil, nil, nil))

	body := bytes.NewBufferString(`{"name":"Carlos","phone":"11999990004"}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations/barbearia-test/customers", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestCreate_400_MissingFields verifica que body sem campos obrigatórios retorna 400.
func TestCreate_400_MissingFields(t *testing.T) {
	db, _ := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 10, "user")

	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations/barbearia-test/customers", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestCreate_409_DuplicatePhone verifica que phone duplicado retorna 409.
func TestCreate_409_DuplicatePhone(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 10, "user")

	mock.ExpectQuery(`SELECT.*organizations`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "name", "phone", "email", "street", "number", "neighborhood", "city", "state", "zip_code", "timezone", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, "Org", "11999990000", "org@test.com", "Rua A", "1", "Bairro", "Cidade", "SP", "00000-000", "America/Sao_Paulo", nil, nil, nil))
	mock.ExpectQuery(`SELECT.*customers`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "name", "phone"}).
			AddRow(1, 1, "Existente", "11999990004"))

	body := bytes.NewBufferString(`{"name":"Carlos","phone":"11999990004"}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations/barbearia-test/customers", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// --- List ---

// TestList_200 verifica que owner lista customers e recebe 200.
func TestList_200(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 10, "user")

	mock.ExpectQuery(`SELECT.*organizations`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "name", "phone", "email", "street", "number", "neighborhood", "city", "state", "zip_code", "timezone", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, "Org", "11999990000", "org@test.com", "Rua A", "1", "Bairro", "Cidade", "SP", "00000-000", "America/Sao_Paulo", nil, nil, nil))
	mock.ExpectQuery(`SELECT.*customers`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "name", "phone", "notes", "created_at", "updated_at"}))

	req := httptest.NewRequest(http.MethodGet, "/organizations/barbearia-test/customers", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestList_401 verifica que request sem auth retorna 401.
func TestList_401(t *testing.T) {
	db, _ := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/organizations/barbearia-test/customers", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestList_403 verifica que usuário que não é owner recebe 403.
func TestList_403(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 20, "user")

	mock.ExpectQuery(`SELECT.*organizations`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "name", "phone", "email", "street", "number", "neighborhood", "city", "state", "zip_code", "timezone", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, "Org", "11999990000", "org@test.com", "Rua A", "1", "Bairro", "Cidade", "SP", "00000-000", "America/Sao_Paulo", nil, nil, nil))

	req := httptest.NewRequest(http.MethodGet, "/organizations/barbearia-test/customers", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestList_404 verifica que slug inexistente retorna 404.
func TestList_404(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 10, "user")

	mock.ExpectQuery(`SELECT.*organizations`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id"}))

	req := httptest.NewRequest(http.MethodGet, "/organizations/nao-existe/customers", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// --- Find ---

// TestFind_200 verifica que owner encontra customer e recebe 200.
func TestFind_200(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 10, "user")

	mock.ExpectQuery(`SELECT.*organizations`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "name", "phone", "email", "street", "number", "neighborhood", "city", "state", "zip_code", "timezone", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, "Org", "11999990000", "org@test.com", "Rua A", "1", "Bairro", "Cidade", "SP", "00000-000", "America/Sao_Paulo", nil, nil, nil))
	mock.ExpectQuery(`SELECT.*customers`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "name", "phone", "notes", "created_at", "updated_at"}).
			AddRow(1, 1, nil, "Carlos Souza", "11999990004", nil, nil, nil))

	req := httptest.NewRequest(http.MethodGet, "/organizations/barbearia-test/customers/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestFind_401 verifica que request sem auth retorna 401.
func TestFind_401(t *testing.T) {
	db, _ := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/organizations/barbearia-test/customers/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestFind_403 verifica que usuário que não é owner recebe 403.
func TestFind_403(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 20, "user")

	mock.ExpectQuery(`SELECT.*organizations`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "name", "phone", "email", "street", "number", "neighborhood", "city", "state", "zip_code", "timezone", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, "Org", "11999990000", "org@test.com", "Rua A", "1", "Bairro", "Cidade", "SP", "00000-000", "America/Sao_Paulo", nil, nil, nil))

	req := httptest.NewRequest(http.MethodGet, "/organizations/barbearia-test/customers/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestFind_404 verifica que ID inexistente retorna 404.
func TestFind_404(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 10, "user")

	mock.ExpectQuery(`SELECT.*organizations`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "name", "phone", "email", "street", "number", "neighborhood", "city", "state", "zip_code", "timezone", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, "Org", "11999990000", "org@test.com", "Rua A", "1", "Bairro", "Cidade", "SP", "00000-000", "America/Sao_Paulo", nil, nil, nil))
	mock.ExpectQuery(`SELECT.*customers`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "name", "phone", "notes", "created_at", "updated_at"}))

	req := httptest.NewRequest(http.MethodGet, "/organizations/barbearia-test/customers/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// --- Update ---

// TestUpdate_200 verifica que owner atualiza customer e recebe 200.
func TestUpdate_200(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 10, "user")

	mock.ExpectQuery(`SELECT.*organizations`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "name", "phone", "email", "street", "number", "neighborhood", "city", "state", "zip_code", "timezone", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, "Org", "11999990000", "org@test.com", "Rua A", "1", "Bairro", "Cidade", "SP", "00000-000", "America/Sao_Paulo", nil, nil, nil))
	mock.ExpectQuery(`SELECT.*customers`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "name", "phone", "notes", "created_at", "updated_at"}).
			AddRow(1, 1, nil, "Carlos Souza", "11999990004", nil, nil, nil))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "customers"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	body := bytes.NewBufferString(`{"name":"Carlos Silva"}`)
	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/customers/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestUpdate_401 verifica que request sem auth retorna 401.
func TestUpdate_401(t *testing.T) {
	db, _ := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(h)

	body := bytes.NewBufferString(`{"name":"Carlos"}`)
	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/customers/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestUpdate_403 verifica que usuário que não é owner recebe 403.
func TestUpdate_403(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 20, "user")

	mock.ExpectQuery(`SELECT.*organizations`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "name", "phone", "email", "street", "number", "neighborhood", "city", "state", "zip_code", "timezone", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, "Org", "11999990000", "org@test.com", "Rua A", "1", "Bairro", "Cidade", "SP", "00000-000", "America/Sao_Paulo", nil, nil, nil))

	body := bytes.NewBufferString(`{"name":"Carlos"}`)
	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/customers/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestUpdate_400_EmptyBody verifica que body sem campos retorna 400.
func TestUpdate_400_EmptyBody(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 10, "user")

	mock.ExpectQuery(`SELECT.*organizations`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "name", "phone", "email", "street", "number", "neighborhood", "city", "state", "zip_code", "timezone", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, "Org", "11999990000", "org@test.com", "Rua A", "1", "Bairro", "Cidade", "SP", "00000-000", "America/Sao_Paulo", nil, nil, nil))
	mock.ExpectQuery(`SELECT.*customers`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "name", "phone", "notes", "created_at", "updated_at"}).
			AddRow(1, 1, nil, "Carlos Souza", "11999990004", nil, nil, nil))

	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/customers/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestUpdate_409_DuplicatePhone verifica que phone duplicado retorna 409.
func TestUpdate_409_DuplicatePhone(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 10, "user")

	mock.ExpectQuery(`SELECT.*organizations`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "name", "phone", "email", "street", "number", "neighborhood", "city", "state", "zip_code", "timezone", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, "Org", "11999990000", "org@test.com", "Rua A", "1", "Bairro", "Cidade", "SP", "00000-000", "America/Sao_Paulo", nil, nil, nil))
	mock.ExpectQuery(`SELECT.*customers`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "name", "phone", "notes", "created_at", "updated_at"}).
			AddRow(1, 1, nil, "Carlos Souza", "11999990004", nil, nil, nil))
	// Verifica duplicidade — outro customer com o mesmo phone
	mock.ExpectQuery(`SELECT.*customers`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "name", "phone"}).
			AddRow(2, 1, "Cliente Manual", "11999990099"))

	body := bytes.NewBufferString(`{"phone":"11999990099"}`)
	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/customers/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// --- Delete ---

// TestDelete_204 verifica que owner deleta customer sem schedules ativos e recebe 204.
func TestDelete_204(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 10, "user")

	mock.ExpectQuery(`SELECT.*organizations`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "name", "phone", "email", "street", "number", "neighborhood", "city", "state", "zip_code", "timezone", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, "Org", "11999990000", "org@test.com", "Rua A", "1", "Bairro", "Cidade", "SP", "00000-000", "America/Sao_Paulo", nil, nil, nil))
	mock.ExpectQuery(`SELECT.*customers`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "name", "phone", "notes", "created_at", "updated_at"}).
			AddRow(2, 1, nil, "Cliente Manual", "11999990099", nil, nil, nil))
	mock.ExpectQuery(`SELECT.*schedules`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id", "status"}))
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "customers"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodDelete, "/organizations/barbearia-test/customers/2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestDelete_401 verifica que request sem auth retorna 401.
func TestDelete_401(t *testing.T) {
	db, _ := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/barbearia-test/customers/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestDelete_403 verifica que usuário que não é owner recebe 403.
func TestDelete_403(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 20, "user")

	mock.ExpectQuery(`SELECT.*organizations`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "name", "phone", "email", "street", "number", "neighborhood", "city", "state", "zip_code", "timezone", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, "Org", "11999990000", "org@test.com", "Rua A", "1", "Bairro", "Cidade", "SP", "00000-000", "America/Sao_Paulo", nil, nil, nil))

	req := httptest.NewRequest(http.MethodDelete, "/organizations/barbearia-test/customers/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestDelete_409_ActiveSchedules verifica que customer com schedules ativos retorna 409.
func TestDelete_409_ActiveSchedules(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 10, "user")

	mock.ExpectQuery(`SELECT.*organizations`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "name", "phone", "email", "street", "number", "neighborhood", "city", "state", "zip_code", "timezone", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, "Org", "11999990000", "org@test.com", "Rua A", "1", "Bairro", "Cidade", "SP", "00000-000", "America/Sao_Paulo", nil, nil, nil))
	mock.ExpectQuery(`SELECT.*customers`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "name", "phone", "notes", "created_at", "updated_at"}).
			AddRow(1, 1, nil, "Carlos Souza", "11999990004", nil, nil, nil))
	mock.ExpectQuery(`SELECT.*schedules`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "customer_id", "status"}).
			AddRow(1, 1, "pending"))

	req := httptest.NewRequest(http.MethodDelete, "/organizations/barbearia-test/customers/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestDelete_404 verifica que ID inexistente retorna 404.
func TestDelete_404(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 10, "user")

	mock.ExpectQuery(`SELECT.*organizations`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "name", "phone", "email", "street", "number", "neighborhood", "city", "state", "zip_code", "timezone", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, "Org", "11999990000", "org@test.com", "Rua A", "1", "Bairro", "Cidade", "SP", "00000-000", "America/Sao_Paulo", nil, nil, nil))
	mock.ExpectQuery(`SELECT.*customers`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "name", "phone", "notes", "created_at", "updated_at"}))

	req := httptest.NewRequest(http.MethodDelete, "/organizations/barbearia-test/customers/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d (body: %s)", w.Code, w.Body.String())
	}
}
