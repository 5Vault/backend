package database

import (
	"backend/src/internal/schemas"
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectPostgres() *gorm.DB {
	dsn := os.Getenv("SUPABASE_DSN")
	if dsn == "" {
		panic("SUPABASE_DSN environment variable not set")
	}
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger:             logger.Default.LogMode(logger.Silent),
		PrepareStmt:        false,
		PrepareStmtMaxSize: 10,
		PrepareStmtTTL:     2 * time.Minute,
	})
	if err != nil {
		panic("failed to connect to Database")
	}
	err = db.AutoMigrate(&schemas.User{}, &schemas.File{}, &schemas.Key{}, &schemas.File{})
	if err != nil {
		panic(fmt.Sprintf("failed to migrate database: %v", err))
	}

	fmt.Println("Connected to Database")
	return db
}
