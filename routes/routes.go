// Package routes registra todas as rotas HTTP da aplicação.
package routes

import (
	"net/http"
	"time"

	"barbearia-api/app"
	auth "barbearia-api/app/auth/handlers"
	customers "barbearia-api/app/customers/handlers"
	members "barbearia-api/app/members/handlers"
	"barbearia-api/app/middleware"
	orgs "barbearia-api/app/organizations/handlers"
	schedules "barbearia-api/app/schedules/handlers"
	services "barbearia-api/app/services/handlers"
	users "barbearia-api/app/users/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	libredis "github.com/ulule/limiter/v3/drivers/store/redis"
	"gorm.io/gorm"
)

// Initialize configura o roteador Gin com todas as rotas e inicia o servidor HTTP.
func Initialize(db *gorm.DB, redisClient *redis.Client) {
	routes := gin.Default()

	routes.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	store, err := libredis.NewStoreWithOptions(redisClient, limiter.StoreOptions{Prefix: "rl"})
	if err != nil {
		panic("failed to create rate limit store: " + err.Error())
	}
	authLimiter := limiter.New(store, limiter.Rate{Period: time.Minute, Limit: 5})
	rl := middleware.RateLimit(authLimiter)

	base := app.NewHandler(db)
	authHandler := &auth.Handler{Handler: base, Redis: redisClient}
	usersHandler := &users.Handler{Handler: base}
	orgsHandler := &orgs.Handler{Handler: base}
	customersHandler := &customers.Handler{Handler: base}
	membersHandler := &members.Handler{Handler: base}
	servicesHandler := &services.Handler{Handler: base}
	schedulesHandler := &schedules.Handler{Handler: base}

	routes.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// Auth
	authGroup := routes.Group("/auth")
	{
		authGroup.POST("/register", rl, authHandler.Register)
		authGroup.POST("/login", rl, authHandler.Login)
		authGroup.GET("/verify", authHandler.Verify)
		authGroup.POST("/refresh", authHandler.Refresh)
		authGroup.POST("/forgot-password", rl, authHandler.ForgotPassword)
		authGroup.POST("/reset-password", authHandler.ResetPassword)
		authGroup.POST("/logout", authHandler.Logout)
	}

	// Users (admin only)
	usersGroup := routes.Group("/users")
	usersGroup.Use(middleware.Auth(redisClient), middleware.RequireAdmin())
	{
		usersGroup.GET("", usersHandler.List)
		usersGroup.GET("/:id", usersHandler.Find)
		usersGroup.POST("", usersHandler.Create)
		usersGroup.PATCH("/:id", usersHandler.Update)
		usersGroup.DELETE("/:id", usersHandler.Delete)
	}

	// Organizations
	routes.POST("/organizations", middleware.Auth(redisClient), orgsHandler.Create)
	routes.GET("/organizations", middleware.Auth(redisClient), orgsHandler.List)
	routes.GET("/organizations/:slug", orgsHandler.Find)
	routes.PATCH("/organizations/:slug", middleware.Auth(redisClient), orgsHandler.Update)
	routes.DELETE("/organizations/:slug", middleware.Auth(redisClient), orgsHandler.Delete)
	routes.GET("/my/organizations", middleware.Auth(redisClient), orgsHandler.MyOrgs)

	routes.GET("/organizations/:slug/business-hours", orgsHandler.GetBusinessHours)
	routes.GET("/organizations/:slug/availability", orgsHandler.GetAvailability)

	routes.GET("/organizations/:slug/exceptions", middleware.Auth(redisClient), orgsHandler.ListOrgExceptions)
	routes.POST("/organizations/:slug/exceptions", middleware.Auth(redisClient), orgsHandler.CreateOrgException)
	routes.DELETE("/organizations/:slug/exceptions/:id", middleware.Auth(redisClient), orgsHandler.DeleteOrgException)

	// Members
	routes.GET("/organizations/:slug/members", membersHandler.List)
	routes.POST("/organizations/:slug/members", middleware.Auth(redisClient), membersHandler.Add)
	routes.DELETE("/organizations/:slug/members/:user_id", middleware.Auth(redisClient), membersHandler.Remove)

	routes.GET("/organizations/:slug/members/:user_id/business-hours", middleware.Auth(redisClient), orgsHandler.GetMemberBusinessHours)
	routes.PATCH("/organizations/:slug/members/:user_id/business-hours", middleware.Auth(redisClient), orgsHandler.UpdateMemberBusinessHoursBatch)
	routes.PATCH("/organizations/:slug/members/:user_id/business-hours/:day", middleware.Auth(redisClient), orgsHandler.UpdateMemberBusinessHourDay)

	routes.GET("/organizations/:slug/members/:user_id/exceptions", middleware.Auth(redisClient), orgsHandler.ListMemberExceptions)
	routes.POST("/organizations/:slug/members/:user_id/exceptions", middleware.Auth(redisClient), orgsHandler.CreateMemberException)
	routes.DELETE("/organizations/:slug/members/:user_id/exceptions/:id", middleware.Auth(redisClient), orgsHandler.DeleteMemberException)

	// Customers
	routes.GET("/organizations/:slug/customers", middleware.Auth(redisClient), customersHandler.List)
	routes.POST("/organizations/:slug/customers", middleware.Auth(redisClient), customersHandler.Create)
	routes.GET("/organizations/:slug/customers/:id", middleware.Auth(redisClient), customersHandler.Find)
	routes.PATCH("/organizations/:slug/customers/:id", middleware.Auth(redisClient), customersHandler.Update)
	routes.DELETE("/organizations/:slug/customers/:id", middleware.Auth(redisClient), customersHandler.Delete)

	// Services
	routes.GET("/organizations/:slug/services", servicesHandler.List)
	routes.GET("/organizations/:slug/services/:id", servicesHandler.Find)
	routes.POST("/organizations/:slug/services", middleware.Auth(redisClient), servicesHandler.Create)
	routes.PATCH("/organizations/:slug/services/:id", middleware.Auth(redisClient), servicesHandler.Update)
	routes.DELETE("/organizations/:slug/services/:id", middleware.Auth(redisClient), servicesHandler.Delete)

	// Schedules
	routes.POST("/organizations/:slug/schedules", middleware.Auth(redisClient), schedulesHandler.Create)
	routes.GET("/organizations/:slug/schedules", middleware.Auth(redisClient), schedulesHandler.List)
	routes.GET("/organizations/:slug/schedules/:id", middleware.Auth(redisClient), schedulesHandler.Find)
	routes.GET("/my/schedules", middleware.Auth(redisClient), schedulesHandler.MySchedules)

	routes.PATCH("/organizations/:slug/schedules/:id/confirm", middleware.Auth(redisClient), schedulesHandler.Confirm)
	routes.PATCH("/organizations/:slug/schedules/:id/cancel", middleware.Auth(redisClient), schedulesHandler.Cancel)
	routes.PATCH("/organizations/:slug/schedules/:id/complete", middleware.Auth(redisClient), schedulesHandler.Complete)
	routes.PATCH("/organizations/:slug/schedules/:id/no-show", middleware.Auth(redisClient), schedulesHandler.NoShow)
	routes.PATCH("/organizations/:slug/schedules/:id/reschedule", middleware.Auth(redisClient), schedulesHandler.Reschedule)
	routes.GET("/organizations/:slug/schedules/:id/reschedule-history", middleware.Auth(redisClient), schedulesHandler.RescheduleHistory)

	routes.Run()
}
