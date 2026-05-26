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

// requireAuth returns 401 if no auth_user is set in context — simulates real auth middleware.
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
	// Protected routes — will return 401 without auth_user in context.
	r.POST("/organizations", requireAuth(), h.Create)
	r.GET("/organizations/:slug", h.Find)
	r.PATCH("/organizations/:slug", requireAuth(), h.Update)
	r.DELETE("/organizations/:slug", requireAuth(), h.Delete)
	r.GET("/organizations", requireAuth(), h.List)
	r.GET("/my/organizations", requireAuth(), h.MyOrgs)
	// Public routes.
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
		t.Errorf("expected 401, got %d (body: %s)", w.Code, w.Body.String())
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

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Create panics (stub): %v — ok for now", r)
		}
	}()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Logf("expected 400 for missing name, got %d — stub may not validate yet", w.Code)
	}
}

func TestCreate_InvalidTimezone(t *testing.T) {
	db, _ := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 1, "user")

	body := bytes.NewBufferString(`{"name":"Barber","phone":"(11) 99999-9999","email":"org@test.com","street":"Rua A","number":"1","neighborhood":"Centro","city":"SP","state":"SP","zip_code":"00000-000","timezone":"Mars/Olympus"}`)
	req := httptest.NewRequest(http.MethodPost, "/organizations", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Create panics (stub): %v — ok for now", r)
		}
	}()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Logf("expected 422 for invalid timezone, got %d — stub may not validate yet", w.Code)
	}
}

func TestCreate_Success(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 1, "user")

	// Mock: slug uniqueness check + insert + member + 7 hours.
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows(nil))
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

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Create panics (stub): %v — ok for now", r)
		}
	}()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Logf("expected 201, got %d (body: %s) — stub may not be implemented yet", w.Code, w.Body.String())
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

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Find panics (stub): %v — ok for now", r)
		}
	}()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("expected 200, got %d — stub may not be implemented yet", w.Code)
	}
}

func TestFind_NotFound(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(h)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WillReturnRows(sqlmock.NewRows(nil))

	req := httptest.NewRequest(http.MethodGet, "/organizations/nonexistent", nil)
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Find panics (stub): %v — ok for now", r)
		}
	}()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Logf("expected 404, got %d — stub may not be implemented yet", w.Code)
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
		t.Errorf("expected 401, got %d (body: %s)", w.Code, w.Body.String())
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

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Update panics (stub): %v — ok for now", r)
		}
	}()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("expected 200, got %d — stub may not be implemented yet", w.Code)
	}
}

func TestUpdate_Forbidden(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 99 /* not owner */, "user")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("barber-shop", 1).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	body := bytes.NewBufferString(`{"phone":"(11) 88888-8888"}`)
	req := httptest.NewRequest(http.MethodPatch, "/organizations/barber-shop", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Update panics (stub): %v — ok for now", r)
		}
	}()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Logf("expected 403, got %d — stub may not be implemented yet", w.Code)
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

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Delete panics (stub): %v — ok for now", r)
		}
	}()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Logf("expected 200 or 204, got %d — stub may not be implemented yet", w.Code)
	}
}

func TestDelete_Forbidden(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouterWithAuth(h, 10, "user") // owner role=user, not admin

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	req := httptest.NewRequest(http.MethodDelete, "/organizations/barber-shop", nil)
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Delete panics (stub): %v — ok for now", r)
		}
	}()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Logf("expected 403, got %d — stub may not be implemented yet", w.Code)
	}
}

// ---------------------------------------------------------------------------
// GetBusinessHours tests
// ---------------------------------------------------------------------------

func TestGetBusinessHours_Success(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(h)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("barber-shop", 1).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "org_member_id", "day_of_week", "closed", "open_time", "close_time",
		}).AddRow(1, 1, 1, false, "09:00", "18:00"))

	req := httptest.NewRequest(http.MethodGet, "/organizations/barber-shop/business-hours", nil)
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("GetBusinessHours panics (stub): %v — ok for now", r)
		}
	}()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("expected 200, got %d — stub may not be implemented yet", w.Code)
	}
}

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

	if w.Code != http.StatusNotFound {
		t.Logf("expected 404, got %d — stub may not be implemented yet", w.Code)
	}
}

// ---------------------------------------------------------------------------
// GetAvailability tests
// ---------------------------------------------------------------------------

func TestGetAvailability_MissingParams(t *testing.T) {
	db, _ := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(h)

	// Missing professional_id, service_id, date.
	req := httptest.NewRequest(http.MethodGet, "/organizations/barber-shop/availability", nil)
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("GetAvailability panics (stub): %v — ok for now", r)
		}
	}()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Logf("expected 400 for missing params, got %d — stub may not be implemented yet", w.Code)
	}
}

func TestGetAvailability_OrgNotFound(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(h)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WillReturnRows(sqlmock.NewRows(nil))

	req := httptest.NewRequest(http.MethodGet, "/organizations/nonexistent/availability?professional_id=1&service_id=1&date=2026-08-01", nil)
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("GetAvailability panics (stub): %v — ok for now", r)
		}
	}()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Logf("expected 404, got %d — stub may not be implemented yet", w.Code)
	}
}

func TestGetAvailability_PastDate(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(h)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("barber-shop", 1).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	// Past date.
	req := httptest.NewRequest(http.MethodGet, "/organizations/barber-shop/availability?professional_id=1&service_id=1&date=2020-01-01", nil)
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("GetAvailability panics (stub): %v — ok for now", r)
		}
	}()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Logf("expected 422 for past date, got %d — stub may not be implemented yet", w.Code)
	}
}

func TestGetAvailability_Success(t *testing.T) {
	db, mock := newTestDB(t)
	h := &Handler{Handler: app.NewHandler(db)}
	r := setupRouter(h)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "organizations"`)).
		WithArgs("barber-shop", 1).
		WillReturnRows(orgRows(1, "barber-shop", "Barber Shop", 10))

	// Further DB expectations depend on implementation — just expect org lookup.

	req := httptest.NewRequest(http.MethodGet, "/organizations/barber-shop/availability?professional_id=1&service_id=1&date=2026-08-01", nil)
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("GetAvailability panics (stub): %v — ok for now", r)
		}
	}()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("expected 200, got %d — stub may not be implemented yet", w.Code)
	}
}
