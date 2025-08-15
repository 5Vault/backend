package routes

import (
	keyHndlr "backend/src/internal/domain/key/handlers"
	keyRepo "backend/src/internal/domain/key/repository"
	keySvc "backend/src/internal/domain/key/service"
	lgnHndlr "backend/src/internal/domain/login/handlers"
	lgnSvc "backend/src/internal/domain/login/services"
	usrHndlr "backend/src/internal/domain/user/handlers"
	usrRepo "backend/src/internal/domain/user/repositories"
	usrSvc "backend/src/internal/domain/user/services"
	middleware2 "backend/src/internal/middleware"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func StartApp(db *gorm.DB) {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"*"},
	}))
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
	newMDW := middleware2.NewMiddleWare(userRepo)
	keyMDW := middleware2.NewKeyMiddleware(keyRepo.NewKeyRepository(db))
	userService := usrSvc.NewUserService(userRepo)
	userHandler := usrHndlr.NewUserHandler(userService)

	userGroup.POST("/", userHandler.CreateUser)
	userGroup.GET("/", newMDW.AuthMiddleware(), userHandler.GetOwnUser)
	userGroup.GET("/:id", userHandler.GetUserByID)

	loginGroup := apiV1.Group("/login")
	loginService := lgnSvc.NewLoginService(userRepo)
	loginHandler := lgnHndlr.NewLoginHandler(loginService)
	loginGroup.POST("/", loginHandler.Try)

	keyGroup := apiV1.Group("/key")

	keyRepository := keyRepo.NewKeyRepository(db)
	keyService := keySvc.NewKeyService(keyRepository)
	keyHandler := keyHndlr.NewKeyHandler(keyService)
	keyGroup.POST("/", keyHandler.CreateKey)
	keyGroup.GET("/validate", keyMDW.ValidateKeysMiddleware(), keyHandler.ValidateKey)

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
