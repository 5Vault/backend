package main

import (
	"backend/src/config"
	"backend/src/internal/routes"
)

func main() {
	db := *database.ConnectPostgres()
	redis := *database.ConnectRedis()

	routes.StartApp(&db, &redis)
}
