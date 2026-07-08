package main

import (
	"backend/src/config"
	"backend/src/internal/logger"
	"backend/src/internal/routes"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func init() {
	_ = godotenv.Load() // carrega .env se existir (silencioso em produção)
	loc := time.FixedZone("UTC-3", -3*60*60)
	time.Local = loc
	logger.Init()
}

func main() {
	defer logger.Sync()

	logger.Info("fivekeepr starting")

	db := *database.ConnectDB()
	logger.Info("database connected")

	redis := *database.ConnectRedis()
	logger.Info("redis connected", zap.String("component", "startup"))

	routes.StartApp(&db, &redis)
}
