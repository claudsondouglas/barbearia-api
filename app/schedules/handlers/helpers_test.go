package scheduleshttp

import (
	"testing"

	"barbearia-api/app"
	"barbearia-api/app/middleware"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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

// setupRouter registra todas as rotas do Handler para uso nos testes.
func setupRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Schedules
	r.POST("/organizations/:slug/schedules", h.Create)
	r.GET("/organizations/:slug/schedules", h.List)
	r.GET("/organizations/:slug/schedules/:id", h.Find)
	r.GET("/my/schedules", h.MySchedules)

	// Transitions
	r.PATCH("/organizations/:slug/schedules/:id/confirm", h.Confirm)
	r.PATCH("/organizations/:slug/schedules/:id/cancel", h.Cancel)
	r.PATCH("/organizations/:slug/schedules/:id/complete", h.Complete)
	r.PATCH("/organizations/:slug/schedules/:id/no-show", h.NoShow)

	// Reschedule
	r.PATCH("/organizations/:slug/schedules/:id/reschedule", h.Reschedule)
	r.GET("/organizations/:slug/schedules/:id/reschedule-history", h.RescheduleHistory)

	return r
}

// setupRouterWithAuth registra todas as rotas com middleware de auth injetado.
func setupRouterWithAuth(h *Handler, userID uint, role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	auth := withAuthUser(userID, role)

	r.POST("/organizations/:slug/schedules", auth, h.Create)
	r.GET("/organizations/:slug/schedules", auth, h.List)
	r.GET("/organizations/:slug/schedules/:id", auth, h.Find)
	r.GET("/my/schedules", auth, h.MySchedules)

	r.PATCH("/organizations/:slug/schedules/:id/confirm", auth, h.Confirm)
	r.PATCH("/organizations/:slug/schedules/:id/cancel", auth, h.Cancel)
	r.PATCH("/organizations/:slug/schedules/:id/complete", auth, h.Complete)
	r.PATCH("/organizations/:slug/schedules/:id/no-show", auth, h.NoShow)

	r.PATCH("/organizations/:slug/schedules/:id/reschedule", auth, h.Reschedule)
	r.GET("/organizations/:slug/schedules/:id/reschedule-history", auth, h.RescheduleHistory)

	return r
}

func newHandler(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	t.Helper()
	db, mock := newTestDB(t)
	return &Handler{Handler: app.NewHandler(db)}, mock
}
