package main

import (
	"barbearia-api/cache"
	"barbearia-api/database"
	"barbearia-api/routes"
)

func main() {
	db := database.Connect()
	redisClient := cache.Connect()
	routes.Initialize(db, redisClient)
}
