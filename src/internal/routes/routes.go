package routes

import (
	adminHndlr "backend/src/internal/domain/admin/handlers"
	adminSvc "backend/src/internal/domain/admin/services"
	bucketHndlr "backend/src/internal/domain/bucket/handlers"
	bucketSvc "backend/src/internal/domain/bucket/services"
	paymentHndlr "backend/src/internal/domain/payment/handlers"
	paymentSvc "backend/src/internal/domain/payment/services"
	pwdHndlr "backend/src/internal/domain/password_reset/handlers"
	pwdSvc "backend/src/internal/domain/password_reset/services"
	strHndlr "backend/src/internal/domain/file/handlers"
	strSvc "backend/src/internal/domain/file/services"
	keyHndlr "backend/src/internal/domain/key/handlers"
	keySvc "backend/src/internal/domain/key/services"
	lgnHndlr "backend/src/internal/domain/login/handlers"
	lgnSvc "backend/src/internal/domain/login/services"
	oauthHndlr "backend/src/internal/domain/oauth/handlers"
	oauthSvc "backend/src/internal/domain/oauth/services"
	notifHndlr "backend/src/internal/domain/notification/handlers"
	ticketHndlr "backend/src/internal/domain/ticket/handlers"
	ticketSvc "backend/src/internal/domain/ticket/services"
	usrHndlr "backend/src/internal/domain/user/handlers"
	usrSvc "backend/src/internal/domain/user/services"
	tierHndlr "backend/src/internal/domain/user/tier/handler"
	tierSvc "backend/src/internal/domain/user/tier/service"
	"backend/src/external"
	"backend/src/internal/actionlog"
	"backend/src/internal/logger"
	"backend/src/internal/middleware"
	actionLogRepo "backend/src/internal/repository/action_log"
	strRepo "backend/src/internal/repository/file"
	keyRepo "backend/src/internal/repository/key"
	paymentRepoP "backend/src/internal/repository/payment"
	pmRepo "backend/src/internal/repository/payment_method"
	passwordResetRepo "backend/src/internal/repository/password_reset"
	bucketRepo "backend/src/internal/repository/storage_config"
	dirRepo "backend/src/internal/repository/storage"
	notifRepoP "backend/src/internal/repository/notification"
	ticketRepoP "backend/src/internal/repository/ticket"
	usrRepo "backend/src/internal/repository/user"
	"backend/src/internal/notif"
	"backend/src/internal/ws"
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func StartApp(db *gorm.DB, redisClient *redis.Client) {
	r := gin.New()
	r.Use(middleware.RequestLogger())

	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Content-Type", "Authorization", "X-Requested-With", "API-Key"},
		ExposeHeaders:   []string{"Content-Length"},
	}))

	rateMDW := middleware.NewRateMiddleware(redisClient)
	r.Use(rateMDW.RateLimitMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK", "timestamp": time.Now().Format(time.RFC3339)})
	})

	// ── Repositórios ──────────────────────────────────────────────────────────
	userRepository := usrRepo.NewUserRepository(db)
	keyRepository := keyRepo.NewKeyRepository(db)
	fileRepository := strRepo.NewStorageRepository(db)
	bucketRepository := bucketRepo.NewStorageConfigRepository(db)
	directoryRepository := dirRepo.NewStorageRepository(db)
	pmRepository := pmRepo.New(db)
	paymentRepository := paymentRepoP.New(db)
	actionLogRepository := actionLogRepo.New(db)
	ticketRepository := ticketRepoP.New(db)
	passwordResetRepository := passwordResetRepo.New(db)
	notificationRepository := notifRepoP.New(db)

	// ── Infra ─────────────────────────────────────────────────────────────────
	emailClient := external.NewEmailClient()
	actionlog.Init(db)
	notif.Init(db)

	// ── Middleware ────────────────────────────────────────────────────────────
	authMDW := middleware.NewMiddleWare(userRepository)
	keyMDW := middleware.NewKeyMiddleware(keyRepository)

	// ── Serviços ──────────────────────────────────────────────────────────────
	bucketService := bucketSvc.NewBucketService(bucketRepository, directoryRepository, userRepository, fileRepository, redisClient)
	go bucketService.EnsureDefaultDomain(context.Background())

	userService := usrSvc.NewUserService(userRepository, keyRepository, emailClient)
	loginService := lgnSvc.NewLoginService(userRepository, actionLogRepository, emailClient)
	keyService := keySvc.NewKeyService(keyRepository)
	fileService := strSvc.NewStorageService(fileRepository)
	oauthService := oauthSvc.NewOAuthService(userRepository, userService)
	discordOAuthService := oauthSvc.NewDiscordOAuthService(userRepository, userService)
	tierService := tierSvc.NewTierService()
	paymentService := paymentSvc.NewPaymentService()
	cardService := paymentSvc.NewCardService(pmRepository)
	adminService := adminSvc.NewAdminService(userRepository, bucketRepository)
	ticketService := ticketSvc.New(ticketRepository, userRepository, emailClient)
	pwdService := pwdSvc.New(passwordResetRepository, userRepository, emailClient)

	// ── Handlers ──────────────────────────────────────────────────────────────
	userHandler := usrHndlr.NewUserHandler(userService)
	loginHandler := lgnHndlr.NewLoginHandler(loginService)
	keyHandler := keyHndlr.NewKeyHandler(keyService)
	fileHandler := strHndlr.NewStorageHandler(fileService)
	oauthHandler := oauthHndlr.NewOAuthHandler(oauthService).WithDiscord(discordOAuthService)
	tierHandler := tierHndlr.NewTierHandler(tierService)
	bucketHandler := bucketHndlr.NewBucketHandler(bucketService)
	paymentHandler := paymentHndlr.NewPaymentHandler(paymentService, tierService, paymentRepository, userRepository, emailClient).WithCardService(cardService)
	cardHandler := paymentHndlr.NewCardHandler(cardService)
	adminHandler := adminHndlr.NewAdminHandler(adminService, userRepository, actionLogRepository, paymentRepository, bucketRepository)
	ticketHandler := ticketHndlr.New(ticketService)
	pwdHandler := pwdHndlr.New(pwdService)
	notifHandler := notifHndlr.New(notificationRepository)

	apiV1 := r.Group("/api/v1")

	apiV1.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": "1.0.0", "timestamp": time.Now().Format(time.RFC3339)})
	})

	// Auth
	authGroup := apiV1.Group("/auth")
	authGroup.POST("/login", loginHandler.Try)
	authGroup.GET("/google", oauthHandler.GoogleLogin)
	authGroup.GET("/google/callback", oauthHandler.GoogleCallback)
	authGroup.GET("/discord", oauthHandler.DiscordLogin)
	authGroup.GET("/discord/callback", oauthHandler.DiscordCallback)
	authGroup.POST("/forgot-password", pwdHandler.Request)
	authGroup.POST("/reset-password", pwdHandler.Reset)

	loginGroup := apiV1.Group("/login")
	loginGroup.POST("/try", loginHandler.Try)

	// Usuário
	userGroup := apiV1.Group("/user")
	userGroup.POST("/", userHandler.CreateUser)
	userGroup.GET("/", authMDW.AuthMiddleware(), userHandler.GetOwnUser)
	userGroup.GET("/:id", userHandler.GetUserByID)
	userGroup.PATCH("/extra-storage", authMDW.AuthMiddleware(), userHandler.SetExtraStorage)
	userGroup.POST("/avatar", authMDW.AuthMiddleware(), userHandler.UploadAvatar)
	userGroup.POST("/2fa/setup", authMDW.AuthMiddleware(), userHandler.Setup2FA)
	userGroup.POST("/2fa/verify", authMDW.AuthMiddleware(), userHandler.Verify2FA)
	userGroup.POST("/2fa/disable", authMDW.AuthMiddleware(), userHandler.Disable2FA)

	// Chaves de API
	keyGroup := apiV1.Group("/key", authMDW.AuthMiddleware())
	keyGroup.POST("/", keyHandler.CreateKey)
	keyGroup.GET("/", keyHandler.ListKeys)
	keyGroup.DELETE("/:id", keyHandler.DeleteKey)
	apiV1.GET("/key/validate", keyMDW.ValidateKeysMiddleware(), keyHandler.ValidateKey)

	// Tier
	apiV1.GET("/tier/", tierHandler.GetTiers)

	// Buckets e diretórios — requer auth
	bucketGroup := apiV1.Group("/bucket", authMDW.AuthMiddleware())
	bucketGroup.GET("/stats", bucketHandler.GetStats)
	bucketGroup.POST("/", bucketHandler.CreateBucket)
	bucketGroup.GET("/", bucketHandler.ListBuckets)
	bucketGroup.DELETE("/:bucketId", bucketHandler.DeleteBucket)
	bucketGroup.PATCH("/:bucketId/domain", bucketHandler.SetDomain)
	bucketGroup.POST("/:bucketId/public-access", bucketHandler.EnablePublicAccess)
	bucketGroup.POST("/:bucketId/dir", bucketHandler.CreateDirectory)
	bucketGroup.GET("/:bucketId/dir", bucketHandler.ListDirectories)
	bucketGroup.DELETE("/:bucketId/dir/:dirId", bucketHandler.DeleteDirectory)
	bucketGroup.POST("/:bucketId/dir/:dirId/files", bucketHandler.UploadFile)
	bucketGroup.GET("/:bucketId/dir/:dirId/files", bucketHandler.ListFiles)
	bucketGroup.DELETE("/:bucketId/dir/:dirId/files/:filename", bucketHandler.DeleteFile)

	// Buckets públicos (via API key)
	bucketPublicGroup := apiV1.Group("/bucket")
	bucketPublicGroup.POST("/:bucketId/upload", keyMDW.ValidateKeysMiddleware(), bucketHandler.UploadFilePublic)
	bucketPublicGroup.GET("/:bucketId/files", keyMDW.ValidateKeysMiddleware(), bucketHandler.ListFilesPublic)
	bucketPublicGroup.DELETE("/:bucketId/files", keyMDW.ValidateKeysMiddleware(), bucketHandler.DeleteFilePublic)

	// Arquivos públicos (via API key)
	fileGroup := apiV1.Group("/file")
	fileGroup.GET("/", keyMDW.ValidateKeysMiddleware(), fileHandler.GetFiles)
	fileGroup.GET("/stats", authMDW.AuthMiddleware(), fileHandler.GetFileStats)
	fileGroup.GET("/:id", fileHandler.GetFileByID)

	// Pagamento
	paymentGroup := apiV1.Group("/payment")
	paymentGroup.POST("/intent", authMDW.AuthMiddleware(), paymentHandler.CreateIntent)
	paymentGroup.POST("/save-card-from-intent", authMDW.AuthMiddleware(), paymentHandler.SaveCardFromIntent)
	paymentGroup.POST("/webhook", paymentHandler.Webhook)
	paymentGroup.GET("/history", authMDW.AuthMiddleware(), paymentHandler.ListPayments)
	paymentGroup.GET("/cards", authMDW.AuthMiddleware(), cardHandler.ListCards)
	paymentGroup.POST("/cards", authMDW.AuthMiddleware(), cardHandler.AttachCard)
	paymentGroup.PATCH("/cards/:pmId/default", authMDW.AuthMiddleware(), cardHandler.SetDefault)
	paymentGroup.DELETE("/cards/:pmId", authMDW.AuthMiddleware(), cardHandler.DeleteCard)

	// LGPD
	apiV1.POST("/lgpd/consent", authMDW.AuthMiddleware(), cardHandler.RecordLGPDConsent)

	// Notificações
	notifGroup := apiV1.Group("/notifications", authMDW.AuthMiddleware())
	notifGroup.GET("/", notifHandler.List)
	notifGroup.GET("/unread-count", notifHandler.UnreadCount)
	notifGroup.POST("/read-all", notifHandler.MarkAllRead)
	notifGroup.POST("/read-entity", notifHandler.MarkReadByEntity)
	notifGroup.PATCH("/:id/read", notifHandler.MarkRead)

	// Tickets de suporte
	ticketGroup := apiV1.Group("/ticket", authMDW.AuthMiddleware())
	ticketGroup.POST("/", ticketHandler.Create)
	ticketGroup.POST("", ticketHandler.Create)
	ticketGroup.GET("/", ticketHandler.List)
	ticketGroup.GET("", ticketHandler.List)
	ticketGroup.GET("/:ticketId", ticketHandler.Get)
	ticketGroup.POST("/:ticketId/reply", ticketHandler.Reply)

	// Admin
	adminGroup := apiV1.Group("/admin", authMDW.AuthMiddleware(), authMDW.AdminMiddleware())
	adminGroup.GET("/stats", adminHandler.GetStats)
	adminGroup.GET("/users", adminHandler.ListUsers)
	adminGroup.PATCH("/users/:id/tier", adminHandler.SetUserTier)
	adminGroup.PATCH("/users/:id/role", adminHandler.SetUserRole)
	adminGroup.DELETE("/users/:id", adminHandler.DeleteUser)
	adminGroup.GET("/users/:id/payments", adminHandler.GetUserPayments)
	adminGroup.GET("/users/:id/logs", adminHandler.GetUserLogs)
	adminGroup.GET("/users/:id/buckets", adminHandler.GetUserBuckets)
	adminGroup.GET("/tickets", ticketHandler.AdminList)
	adminGroup.GET("/tickets/:ticketId", ticketHandler.AdminGet)
	adminGroup.POST("/tickets/:ticketId/reply", ticketHandler.AdminReply)
	adminGroup.PATCH("/tickets/:ticketId/close", ticketHandler.AdminClose)

	// WebSocket — tempo real por ticket (token via query param)
	r.GET("/ws/ticket/:ticketId", ws.TicketWS(ticketRepository, userRepository))

	r.Static("/uploads", "./uploads")

	logger.Info("server starting", zap.String("addr", ":8000"))
	if err := r.Run(":8000"); err != nil {
		logger.Error("server failed", zap.Error(err))
	}
}
