// Package database gerencia a conexão com o PostgreSQL via GORM
// e aplica as migrations pendentes via goose ao iniciar.
package database

import (
	"embed"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Connect lê DATABASE_URL do ambiente, abre a conexão com o PostgreSQL
// e aplica todas as migrations pendentes.
func Connect() *gorm.DB {
	if err := godotenv.Load(); err != nil {
		log.Println("warning: .env file not found, relying on environment variables")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}

	goose.SetBaseFS(migrations)
	goose.SetLogger(goose.NopLogger())

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose dialect: %v", err)
	}

	if err := goose.Up(sqlDB, "migrations"); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	log.Println("database connection established")
	return db
}
