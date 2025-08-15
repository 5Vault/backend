package routes

import (
	keyHndlr "backend/internal/key/handlers"
	keyRepo "backend/internal/key/repository"
	keySvc "backend/internal/key/service"
	lgnHndlr "backend/internal/login/handlers"
	lgnSvc "backend/internal/login/services"
	"backend/internal/middleware"
	usrHndlr "backend/internal/user/handlers"
	usrRepo "backend/internal/user/repositories"
	usrSvc "backend/internal/user/services"
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
	newMDW := middleware.NewMiddleWare(userRepo)
	keyMDW := middleware.NewKeyMiddleware(keyRepo.NewKeyRepository(db))
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
