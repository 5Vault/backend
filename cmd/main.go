package main

import (
	"backend/config"
	"backend/internal/routes"
)

func main() {
	db := *database.ConnectPostgres()
	routes.StartApp(&db)
}
