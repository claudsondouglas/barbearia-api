// Package routes registra todas as rotas HTTP da aplicação.
package routes

import (
	"net/http"
	"time"

	"barbearia-api/app"
	auth "barbearia-api/app/auth/handlers"
	"barbearia-api/app/middleware"
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
	usersHandler := &users.Handler{Handler: base}
	authHandler := &auth.Handler{Handler: base, Redis: redisClient}

	routes.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

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

	usersGroup := routes.Group("/users")
	usersGroup.Use(middleware.Auth(redisClient), middleware.RequireAdmin())
	{
		usersGroup.GET("", usersHandler.List)
		usersGroup.GET("/:id", usersHandler.Find)
		usersGroup.POST("", usersHandler.Create)
		usersGroup.PATCH("/:id", usersHandler.Update)
		usersGroup.DELETE("/:id", usersHandler.Delete)
	}

	routes.Run()
}
