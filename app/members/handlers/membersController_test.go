package membershttp

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

// withAuthUser injeta um AuthUser no contexto Gin sem passar pelo middleware real.
func withAuthUser(userID uint, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("auth_user", middleware.AuthUser{ID: userID, Role: role})
		c.Next()
	}
}

// setupRouter cria um roteador Gin de teste com as rotas do Handler configuradas.
func setupRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/organizations/:slug/members", handler.Add)
	r.DELETE("/organizations/:slug/members/:user_id", handler.Remove)
	r.GET("/organizations/:slug/members", handler.List)
	return r
}

// setupRouterWithAuth cria um roteador com middleware de auth injetado nas rotas protegidas.
func setupRouterWithAuth(handler *Handler, userID uint, role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/organizations/:slug/members", withAuthUser(userID, role), handler.Add)
	r.DELETE("/organizations/:slug/members/:user_id", withAuthUser(userID, role), handler.Remove)
	r.GET("/organizations/:slug/members", handler.List)
	return r
}

// TestAddMember_Unauthenticated verifica que POST /members sem auth retorna 401.
func TestAddMember_Unauthenticated(t *testing.T) {
	db, _ := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(handler)

	body := bytes.NewBufferString(`{"user_id": 20}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations/barber-shop/members", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestAddMember_MissingBody verifica que POST sem user_id retorna 400.
func TestAddMember_MissingBody(t *testing.T) {
	db, _ := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(handler, 10, "admin")

	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations/barber-shop/members", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// TestAddMember_Success verifica que POST /members com dados válidos cria o membro.
func TestAddMember_Success(t *testing.T) {
	db, mock := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(handler, 10, "admin")

	// Mock: busca org
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barber-shop").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}).
			AddRow(1, "barber-shop", 10, time.Now(), nil))

	// Mock: busca usuário solicitante
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(10, "Owner", "owner@test.com", "hashed", "admin", time.Now(), time.Now()))

	// Mock: busca usuário alvo
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(20).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(20, "New Member", "new@test.com", "hashed", "user", time.Now(), time.Now()))

	// Mock: sem membro existente
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "org_members"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "created_at", "deleted_at"}))

	// Mock: INSERT org_member
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "org_members"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(100))
	mock.ExpectCommit()

	// Mock: INSERT 7 business hours
	for i := 0; i < 7; i++ {
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "member_business_hours"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint(i + 1)))
		mock.ExpectCommit()
	}

	input := models.AddMemberInput{UserID: 20}
	bodyBytes, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/organizations/barber-shop/members", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Errorf("expected 201 or 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestRemoveMember_Unauthenticated verifica que DELETE sem auth retorna 401.
func TestRemoveMember_Unauthenticated(t *testing.T) {
	db, _ := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/barber-shop/members/20", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// TestRemoveMember_OwnerCannotBeRemoved verifica que remover o dono retorna 422.
func TestRemoveMember_OwnerCannotBeRemoved(t *testing.T) {
	db, mock := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(handler, 10, "admin")

	// ownerID=10, targetUserID=10 → ErrOwnerCannotBeRemoved
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barber-shop").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}).
			AddRow(1, "barber-shop", 10, time.Now(), nil))

	req := httptest.NewRequest(http.MethodDelete, "/organizations/barber-shop/members/10", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestRemoveMember_WithPendingSchedules verifica que membro com schedules pendentes retorna 409.
func TestRemoveMember_WithPendingSchedules(t *testing.T) {
	db, mock := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(handler, 10, "admin")

	// Mock: busca org
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barber-shop").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}).
			AddRow(1, "barber-shop", 10, time.Now(), nil))

	// Mock: busca solicitante
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
			AddRow(10, "Owner", "owner@test.com", "hashed", "admin", time.Now(), time.Now()))

	// Mock: membro ativo
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "org_members"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "user_id", "created_at", "deleted_at"}).
			AddRow(50, 1, 20, time.Now(), nil))

	// Mock: schedules ativos (pending)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "schedules"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "org_member_id", "status"}).
			AddRow(1, 50, "pending"))

	req := httptest.NewRequest(http.MethodDelete, "/organizations/barber-shop/members/20", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestListMembers_Public verifica que GET /members sem auth retorna 200.
func TestListMembers_Public(t *testing.T) {
	db, mock := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(handler)

	// Mock: busca org
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barber-shop").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}).
			AddRow(1, "barber-shop", 10, time.Now(), nil))

	// Mock: lista membros
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "email", "created_at"}).
			AddRow(1, 10, "Alice", "alice@test.com", time.Now()).
			AddRow(2, 20, "Bob", "bob@test.com", time.Now()))

	req := httptest.NewRequest(http.MethodGet, "/organizations/barber-shop/members", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestListMembers_OrgNotFound verifica que GET /members com slug inválido retorna 404.
func TestListMembers_OrgNotFound(t *testing.T) {
	db, mock := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(handler)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("nonexistent").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}))

	req := httptest.NewRequest(http.MethodGet, "/organizations/nonexistent/members", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestListMembers_Empty verifica que GET /members retorna 200 com array vazio quando não há membros.
func TestListMembers_Empty(t *testing.T) {
	db, mock := newTestDB(t)
	handler := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(handler)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations" WHERE slug = $1`)).
		WithArgs("barber-shop").
		WillReturnRows(sqlmock.NewRows([]string{"id", "slug", "owner_id", "created_at", "deleted_at"}).
			AddRow(1, "barber-shop", 10, time.Now(), nil))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "email", "created_at"}))

	req := httptest.NewRequest(http.MethodGet, "/organizations/barber-shop/members", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	var result []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty array, got %d elements", len(result))
	}
}
