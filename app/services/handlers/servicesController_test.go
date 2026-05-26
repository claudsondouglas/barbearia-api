package serviceshttp

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
	"barbearia-api/models"

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

// serviceRows cria rows de sqlmock com as colunas básicas de um serviço.
func serviceRows(id uint, orgID uint, name string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "organization_id", "name", "description", "price", "duration_min",
		"active", "created_at", "updated_at", "deleted_at",
	}).AddRow(id, orgID, name, nil, 30.00, 30, true, time.Now(), time.Now(), nil)
}

// withAuthUser injeta um AuthUser no contexto Gin sem passar pelo middleware real.
func withAuthUser(userID uint, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("auth_user", middleware.AuthUser{ID: userID, Role: role})
		c.Next()
	}
}

// setupRouter cria um roteador Gin de teste com as rotas do Handler configuradas (sem auth).
func setupRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/organizations/:slug/services", handler.List)
	r.GET("/organizations/:slug/services/:id", handler.Find)
	r.POST("/organizations/:slug/services", handler.Create)
	r.PATCH("/organizations/:slug/services/:id", handler.Update)
	r.DELETE("/organizations/:slug/services/:id", handler.Delete)
	return r
}

// setupRouterWithAuth cria um roteador com middleware de auth injetado nas rotas protegidas.
func setupRouterWithAuth(handler *Handler, userID uint, role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/organizations/:slug/services", handler.List)
	r.GET("/organizations/:slug/services/:id", handler.Find)
	r.POST("/organizations/:slug/services", withAuthUser(userID, role), handler.Create)
	r.PATCH("/organizations/:slug/services/:id", withAuthUser(userID, role), handler.Update)
	r.DELETE("/organizations/:slug/services/:id", withAuthUser(userID, role), handler.Delete)
	return r
}

// --- List ---

// TestListServices_Public verifica que GET /services sem auth retorna 200 com serviços ativos.
func TestListServices_Public(t *testing.T) {
	db, mock := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(handler)

	// org existe
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, time.Now(), nil))

	// 2 serviços ativos
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "price", "duration_min",
			"active", "created_at", "updated_at", "deleted_at",
		}).
			AddRow(1, 1, "Corte", nil, 30.00, 30, true, time.Now(), time.Now(), nil).
			AddRow(3, 1, "Sobrancelha", nil, 10.00, 10, true, time.Now(), time.Now(), nil))

	req := httptest.NewRequest(http.MethodGet, "/organizations/barbearia-test/services", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	var result []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 services, got %d", len(result))
	}
}

// TestListServices_OrgNotFound verifica que slug inexistente retorna 404.
func TestListServices_OrgNotFound(t *testing.T) {
	db, mock := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(handler)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("nonexistent").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}))

	req := httptest.NewRequest(http.MethodGet, "/organizations/nonexistent/services", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// --- Find ---

// TestFindService_200 verifica que GET /services/:id retorna 200 para serviço ativo.
func TestFindService_200(t *testing.T) {
	db, mock := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(handler)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, time.Now(), nil))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(serviceRows(1, 1, "Corte"))

	req := httptest.NewRequest(http.MethodGet, "/organizations/barbearia-test/services/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestFindService_404 verifica que ID inexistente retorna 404.
func TestFindService_404(t *testing.T) {
	db, mock := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(handler)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, time.Now(), nil))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "price", "duration_min",
			"active", "created_at", "updated_at", "deleted_at",
		}))

	req := httptest.NewRequest(http.MethodGet, "/organizations/barbearia-test/services/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// --- Create ---

// TestCreateService_Owner_201 verifica que owner autenticado cria serviço e recebe 201.
func TestCreateService_Owner_201(t *testing.T) {
	db, mock := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	// userID=10 é owner
	r := setupRouterWithAuth(handler, 10, "user")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, time.Now(), nil))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(10, "Owner", "owner@test.com", "hashed", "user", time.Now(), time.Now()))

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "services"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, time.Now(), time.Now()))
	mock.ExpectCommit()

	input := models.CreateServiceInput{Name: "Corte", Price: 30.00, DurationMin: 30}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/organizations/barbearia-test/services", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestCreateService_Unauthenticated_401 verifica que POST sem auth retorna 401.
func TestCreateService_Unauthenticated_401(t *testing.T) {
	db, _ := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(handler)

	body := bytes.NewBufferString(`{"name":"Corte","price":30,"duration_min":30}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations/barbearia-test/services", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestCreateService_RegularUser_403 verifica que usuário sem permissão recebe 403.
func TestCreateService_RegularUser_403(t *testing.T) {
	db, mock := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	// userID=12 não é owner/admin
	r := setupRouterWithAuth(handler, 12, "user")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, time.Now(), nil))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(12)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(12, "Regular", "regular@test.com", "hashed", "user", time.Now(), time.Now()))

	body := bytes.NewBufferString(`{"name":"Corte","price":30,"duration_min":30}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations/barbearia-test/services", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestCreateService_MissingFields_400 verifica que body sem campos obrigatórios retorna 400.
func TestCreateService_MissingFields_400(t *testing.T) {
	db, _ := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(handler, 10, "user")

	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations/barbearia-test/services", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestCreateService_PriceZero_422 verifica que price=0 retorna 422.
func TestCreateService_PriceZero_422(t *testing.T) {
	db, mock := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(handler, 10, "user")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, time.Now(), nil))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(10, "Owner", "owner@test.com", "hashed", "user", time.Now(), time.Now()))

	body := bytes.NewBufferString(`{"name":"Corte","price":0,"duration_min":30}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations/barbearia-test/services", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestCreateService_DurationAboveMax_422 verifica que duration_min=241 retorna 422.
func TestCreateService_DurationAboveMax_422(t *testing.T) {
	db, mock := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(handler, 10, "user")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, time.Now(), nil))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(10, "Owner", "owner@test.com", "hashed", "user", time.Now(), time.Now()))

	body := bytes.NewBufferString(`{"name":"Corte","price":30,"duration_min":241}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations/barbearia-test/services", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// --- Update ---

// TestUpdateService_200 verifica que owner pode atualizar serviço e recebe 200.
func TestUpdateService_200(t *testing.T) {
	db, mock := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(handler, 10, "user")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, time.Now(), nil))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(10, "Owner", "owner@test.com", "hashed", "user", time.Now(), time.Now()))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(serviceRows(1, 1, "Corte"))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "services"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	body := bytes.NewBufferString(`{"name":"Corte Novo"}`)
	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/services/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestUpdateService_Unauthenticated_401 verifica que PATCH sem auth retorna 401.
func TestUpdateService_Unauthenticated_401(t *testing.T) {
	db, _ := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(handler)

	body := bytes.NewBufferString(`{"name":"Novo"}`)
	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/services/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestUpdateService_RegularUser_403 verifica que usuário comum recebe 403.
func TestUpdateService_RegularUser_403(t *testing.T) {
	db, mock := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(handler, 12, "user")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, time.Now(), nil))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(12)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(12, "Regular", "regular@test.com", "hashed", "user", time.Now(), time.Now()))

	body := bytes.NewBufferString(`{"name":"Novo"}`)
	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/services/1", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestUpdateService_DeletedService_422 verifica que PATCH em serviço deletado retorna 422.
func TestUpdateService_DeletedService_422(t *testing.T) {
	db, mock := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(handler, 10, "user")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, time.Now(), nil))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(10, "Owner", "owner@test.com", "hashed", "user", time.Now(), time.Now()))

	deletedAt := time.Now().Add(-24 * time.Hour)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "name", "description", "price", "duration_min",
			"active", "created_at", "updated_at", "deleted_at",
		}).AddRow(3, 1, "Sobrancelha", nil, 10.00, 10, false, time.Now(), time.Now(), deletedAt))

	body := bytes.NewBufferString(`{"name":"Sobrancelha Nova"}`)
	req := httptest.NewRequest(http.MethodPatch, "/organizations/barbearia-test/services/3", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// --- Delete ---

// TestDeleteService_204 verifica que owner deleta serviço sem agendamentos ativos e recebe 204.
func TestDeleteService_204(t *testing.T) {
	db, mock := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(handler, 10, "user")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, time.Now(), nil))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(10, "Owner", "owner@test.com", "hashed", "user", time.Now(), time.Now()))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(serviceRows(1, 1, "Corte"))

	// Sem agendamentos ativos
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "service_id", "status"}))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "services"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodDelete, "/organizations/barbearia-test/services/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Errorf("expected 204 or 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestDeleteService_ActiveSchedules_409 verifica que serviço com agendamentos ativos retorna 409.
func TestDeleteService_ActiveSchedules_409(t *testing.T) {
	db, mock := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(handler, 10, "user")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, time.Now(), nil))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(10, "Owner", "owner@test.com", "hashed", "user", time.Now(), time.Now()))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(serviceRows(1, 1, "Corte"))

	// Agendamento pendente encontrado
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "service_id", "status"}).
			AddRow(1, 1, "pending"))

	req := httptest.NewRequest(http.MethodDelete, "/organizations/barbearia-test/services/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestDeleteService_Unauthenticated_401 verifica que DELETE sem auth retorna 401.
func TestDeleteService_Unauthenticated_401(t *testing.T) {
	db, _ := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/barbearia-test/services/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestDeleteService_RegularUser_403 verifica que usuário comum recebe 403.
func TestDeleteService_RegularUser_403(t *testing.T) {
	db, mock := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(handler, 12, "user")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barbearia-test").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}).
			AddRow(1, "barbearia-test", 10, time.Now(), nil))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(uint(12)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(12, "Regular", "regular@test.com", "hashed", "user", time.Now(), time.Now()))

	req := httptest.NewRequest(http.MethodDelete, "/organizations/barbearia-test/services/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d (body: %s)", w.Code, w.Body.String())
	}
}
