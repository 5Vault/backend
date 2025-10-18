package routes

import (
	strHndlr "backend/src/internal/domain/file/handlers"
	strSvc "backend/src/internal/domain/file/services"
	keyHndlr "backend/src/internal/domain/key/handlers"
	keySvc "backend/src/internal/domain/key/services"
	lgnHndlr "backend/src/internal/domain/login/handlers"
	lgnSvc "backend/src/internal/domain/login/services"
	usrHndlr "backend/src/internal/domain/user/handlers"
	usrSvc "backend/src/internal/domain/user/services"
	tierHndlr "backend/src/internal/domain/user/tier/handler"
	tierSvc "backend/src/internal/domain/user/tier/service"
	middleware2 "backend/src/internal/middleware"
	strRepo "backend/src/internal/repository/file"
	keyRepo "backend/src/internal/repository/key"
	usrRepo "backend/src/internal/repository/user"

	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func StartApp(db *gorm.DB, redis *redis.Client) {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"*", "Content-Type", "Authorization", "X-Requested-With", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Access-Control-Allow-Methods", "XSRF-Token"},
		ExposeHeaders:   []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Access-Control-Allow-Methods"},
	}))

	// Add rate limit middleware
	rateMDW := middleware2.NewRateMiddleware(redis)
	r.Use(rateMDW.RateLimitMiddleware())

	apiV1 := r.Group("/api/v1")
	apiV1.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":   "Hello World",
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "1.0.0",
		})
	})

	userGroup := apiV1.Group("/user")
	userRepo := usrRepo.NewUserRepository(db)
	keyRepository := keyRepo.NewKeyRepository(db)
	newMDW := middleware2.NewMiddleWare(userRepo)
	keyMDW := middleware2.NewKeyMiddleware(keyRepository)
	userService := usrSvc.NewUserService(userRepo, keyRepository)
	userHandler := usrHndlr.NewUserHandler(userService)

	userGroup.POST("/", userHandler.CreateUser)
	userGroup.GET("/", newMDW.AuthMiddleware(), userHandler.GetOwnUser)
	userGroup.GET("/:id", userHandler.GetUserByID)

	loginGroup := apiV1.Group("/login")
	loginService := lgnSvc.NewLoginService(userRepo)
	loginHandler := lgnHndlr.NewLoginHandler(loginService)
	loginGroup.POST("/try", loginHandler.Try)

	keyGroup := apiV1.Group("/key")

	keyService := keySvc.NewKeyService(keyRepository)
	keyHandler := keyHndlr.NewKeyHandler(keyService)
	keyGroup.POST("/", keyHandler.CreateKey)
	keyGroup.PUT("/reset", newMDW.AuthMiddleware(), keyHandler.CreateKey)
	keyGroup.GET("/validate", keyMDW.ValidateKeysMiddleware(), keyHandler.ValidateKey)

	storageGroup := apiV1.Group("/file")
	storageRepository := strRepo.NewStorageRepository(db)
	storageService := strSvc.NewStorageService(storageRepository)
	storageHandler := strHndlr.NewStorageHandler(storageService)
	storageGroup.POST("/upload", keyMDW.ValidateKeysMiddleware(), storageHandler.UploadFile)
	storageGroup.GET("/", keyMDW.ValidateKeysMiddleware(), storageHandler.GetFiles)
	storageGroup.GET("/:id", keyMDW.ValidateKeysMiddleware(), storageHandler.GetFileByID)
	storageGroup.GET("/stats", newMDW.AuthMiddleware(), storageHandler.GetFileStats)

	tierGroup := apiV1.Group("/tier")
	tierService := tierSvc.NewTierService()
	tierHandler := tierHndlr.NewTierHandler(tierService)
	tierGroup.GET("/", tierHandler.GetTiers)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "OK",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	if err := r.Run(":8000"); err != nil {
		log.Fatalf("panic: %v", err)
		return
	}
}
