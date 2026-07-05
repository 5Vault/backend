package database

import (
	"backend/src/internal/schemas"
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDB() *gorm.DB {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		panic("DATABASE_DSN environment variable not set")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	if err := db.AutoMigrate(
		&schemas.User{},
		&schemas.Bucket{},
		&schemas.Directory{},
		&schemas.File{},
		&schemas.Key{},
		&schemas.KeyBucketPermission{},
		&schemas.PaymentMethod{},
		&schemas.Payment{},
		&schemas.ActionLog{},
		&schemas.Ticket{},
		&schemas.TicketMessage{},
		&schemas.PasswordResetToken{},
		&schemas.Notification{},
		&schemas.BackupBucket{},
		&schemas.BackupSession{},
	); err != nil {
		panic(fmt.Sprintf("failed to migrate database: %v", err))
	}

	// Keys criadas antes do campo all_buckets existir ficam com false e sem bucket_perms.
	// Corrije retroativamente: se não tem nenhuma bucket_perm, considera acesso a todos.
	db.Exec(`
		UPDATE keys k
		LEFT JOIN key_bucket_permissions kbp ON kbp.key_id = k.id
		SET k.all_buckets = true
		WHERE k.all_buckets = false AND kbp.id IS NULL
	`)

	fmt.Println("Connected to MariaDB")
	return db
}
