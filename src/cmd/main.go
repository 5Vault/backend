package main

import (
	"backend/src/config"
	"backend/src/internal/routes"
)

func main() {
	db := *database.ConnectPostgres()
	routes.StartApp(&db)
}
