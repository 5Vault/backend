package database

import (
	// schemas2 "backend/internal/user/schemas"
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
	//err = db.AutoMigrate(&schemas2.User{}, &schemas2.File{}, &schemas2.Storage{}, &schemas2.Key{})
	//if err != nil {
	//	panic(fmt.Sprintf("failed to migrate database: %v", err))
	//}

	fmt.Println("Connected to Database")
	return db
}
