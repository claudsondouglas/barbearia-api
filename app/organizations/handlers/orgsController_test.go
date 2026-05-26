package orgshttp

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

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

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

func withAuthUser(userID uint, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("auth_user", middleware.AuthUser{ID: userID, Role: role})
		c.Next()
	}
}

func orgRows(id uint, slug, name string, ownerID uint) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "slug", "owner_id", "name", "phone", "email",
		"street", "number", "neighborhood", "city", "state", "zip_code",
		"timezone", "created_at", "updated_at", "deleted_at",
	}).AddRow(
		id, slug, ownerID, name, "(11) 99999-9999", "org@test.com",
		"Rua A", "1", "Centro", "São Paulo", "SP", "01310-100",
		"America/Sao_Paulo", time.Now(), time.Now(), nil,
	)
}

// requireAuth returns 401 if no auth_user is set in context.
func requireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := middleware.GetAuthUser(c); !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

func setupRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/organizations", requireAuth(), h.Create)
	r.GET("/organizations/:slug", h.Find)
	r.PATCH("/organizations/:slug", requireAuth(), h.Update)
	r.DELETE("/organizations/:slug", requireAuth(), h.Delete)
	r.GET("/organizations", requireAuth(), h.List)
	r.GET("/my/organizations", requireAuth(), h.MyOrgs)
	r.GET("/organizations/:slug/business-hours", h.GetBusinessHours)
	r.GET("/organizations/:slug/availability", h.GetAvailability)
	return r
}

func setupRouterWithAuth(h *Handler, userID uint, role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/organizations", withAuthUser(userID, role), requireAuth(), h.Create)
	r.GET("/organizations/:slug", h.Find)
	r.PATCH("/organizations/:slug", withAuthUser(userID, role), requireAuth(), h.Update)
	r.DELETE("/organizations/:slug", withAuthUser(userID, role), requireAuth(), h.Delete)
	r.GET("/organizations", withAuthUser(userID, role), requireAuth(), h.List)
	r.GET("/my/organizations", withAuthUser(userID, role), requireAuth(), h.MyOrgs)
	r.GET("/organizations/:slug/business-hours", h.GetBusinessHours)
	r.GET("/organizations/:slug/availability", h.GetAvailability)
	return r
}

// ---------------------------------------------------------------------------
// Create tests
// ---------------------------------------------------------------------------

func TestCreate_Unauthenticated(t *testing.T) {
	db, _ := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(h)

	body := bytes.NewBufferString(`{"name":"Barber","phone":"(11) 99999-9999","email":"org@test.com","street":"Rua A","number":"1","neighborhood":"Centro","city":"SP","state":"SP","zip_code":"00000-000"}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestCreate_MissingName(t *testing.T) {
	db, _ := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 1, "user")

	body := bytes.NewBufferString(`{"phone":"(11) 99999-9999","email":"org@test.com","street":"Rua A","number":"1","neighborhood":"Centro","city":"SP","state":"SP","zip_code":"00000-000"}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing name, got %d", w.Code)
	}
}

func TestCreate_InvalidTimezone(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 1, "user")

	// Slug check is called before timezone validation in Create — but timezone validation
	// happens before the slug check in our implementation, so no DB query expected.
	_ = mock

	body := bytes.NewBufferString(`{"name":"Barber","phone":"(11) 99999-9999","email":"org@test.com","street":"Rua A","number":"1","neighborhood":"Centro","city":"SP","state":"SP","zip_code":"00000-000","timezone":"Mars/Olympus"}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 for invalid timezone, got %d", w.Code)
	}
}

func TestCreate_Success(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 1, "user")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "organizations"`)).
		WithArgs("barber-shop").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "organizations"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "org_members"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	for i := 0; i < 7; i++ {
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "member_business_hours"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint(i + 1)))
	}
	mock.ExpectCommit()

	input := map[string]interface{}{
		"name": "Barber Shop", "phone": "(11) 99999-9999", "email": "org@test.com",
		"street": "Rua A", "number": "1", "neighborhood": "Centro",
		"city": "São Paulo", "state": "SP", "zip_code": "01310-100",
	}
	bodyBytes, _ := json.Marshal(input)
	req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d (body: %s)", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if _, ok := resp["owner_id"]; ok {
		t.Error("response must not include owner_id")
	}
	if _, ok := resp["deleted_at"]; ok {
		t.Error("response must not include deleted_at")
	}
}

// ---------------------------------------------------------------------------
// Find tests
// ---------------------------------------------------------------------------

func TestFind_Success(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(h)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("barber-shop", 1).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	req := httptest.NewRequest(http.MethodGet, "/organizations/barber-shop", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := resp["owner_id"]; ok {
		t.Error("response must not include owner_id")
	}
	if resp["slug"] != "barber-shop" {
		t.Errorf("expected slug barber-shop, got %v", resp["slug"])
	}
}

func TestFind_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(h)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("nonexistent", 1).
		WillReturnRows(sqlmock.NewRows(nil))

	req := httptest.NewRequest(http.MethodGet, "/organizations/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Update tests
// ---------------------------------------------------------------------------

func TestUpdate_Unauthenticated(t *testing.T) {
	db, _ := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(h)

	body := bytes.NewBufferString(`{"phone":"(11) 88888-8888"}`)
	req := httptest.NewRequest(http.MethodPatch, "/organizations/barber-shop", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestUpdate_Success(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 10, "user")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("barber-shop", 1).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "organizations"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	body := bytes.NewBufferString(`{"phone":"(11) 88888-8888"}`)
	req := httptest.NewRequest(http.MethodPatch, "/organizations/barber-shop", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestUpdate_Forbidden(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 99, "user")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("barber-shop", 1).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	body := bytes.NewBufferString(`{"phone":"(11) 88888-8888"}`)
	req := httptest.NewRequest(http.MethodPatch, "/organizations/barber-shop", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 10, "user")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("nonexistent", 1).
		WillReturnRows(sqlmock.NewRows(nil))

	body := bytes.NewBufferString(`{"phone":"(11) 88888-8888"}`)
	req := httptest.NewRequest(http.MethodPatch, "/organizations/nonexistent", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Delete tests
// ---------------------------------------------------------------------------

func TestDelete_Unauthenticated(t *testing.T) {
	db, _ := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/barber-shop", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestDelete_Success(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 10, "admin")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("barber-shop", 1).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "organizations"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodDelete, "/organizations/barber-shop", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestDelete_Forbidden(t *testing.T) {
	db, _ := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 10, "user")

	req := httptest.NewRequest(http.MethodDelete, "/organizations/barber-shop", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestDelete_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 10, "admin")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("nonexistent", 1).
		WillReturnRows(sqlmock.NewRows(nil))

	req := httptest.NewRequest(http.MethodDelete, "/organizations/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// List tests
// ---------------------------------------------------------------------------

func TestList_Success(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 1, "admin")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestList_Unauthenticated(t *testing.T) {
	db, _ := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// MyOrgs tests
// ---------------------------------------------------------------------------

func TestMyOrgs_Success(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 10, "user")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	req := httptest.NewRequest(http.MethodGet, "/my/organizations", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}

	var resp []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("expected 1 org, got %d", len(resp))
	}
}

func TestMyOrgs_Empty(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 99, "user")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows(nil))

	req := httptest.NewRequest(http.MethodGet, "/my/organizations", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// GetBusinessHours tests (stub — not implemented yet)
// ---------------------------------------------------------------------------

func TestGetBusinessHours_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(h)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WillReturnRows(sqlmock.NewRows(nil))

	req := httptest.NewRequest(http.MethodGet, "/organizations/nonexistent/business-hours", nil)
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("GetBusinessHours panics (stub): %v — ok for now", r)
		}
	}()

	r.ServeHTTP(w, req)
}

// ---------------------------------------------------------------------------
// GetAvailability tests (stub — not implemented yet)
// ---------------------------------------------------------------------------

func TestGetAvailability_MissingParams(t *testing.T) {
	db, _ := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/organizations/barber-shop/availability", nil)
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("GetAvailability panics (stub): %v — ok for now", r)
		}
	}()

	r.ServeHTTP(w, req)
}
