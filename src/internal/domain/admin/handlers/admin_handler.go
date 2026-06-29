package adminHandlers

import (
	"backend/src/internal/domain/admin/services"
	usrRepo "backend/src/internal/repository/user"
	actionLogRepo "backend/src/internal/repository/action_log"
	paymentRepo "backend/src/internal/repository/payment"
	bucketRepo "backend/src/internal/repository/storage_config"
	"backend/src/pkg/apperr"
	"backend/src/pkg/respond"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	Svc           *adminServices.AdminService
	UserRepo      *usrRepo.UserRepository
	ActionLogRepo *actionLogRepo.ActionLogRepository
	PaymentRepo   *paymentRepo.PaymentRepository
	BucketRepo    *bucketRepo.BucketRepository
}

func NewAdminHandler(svc *adminServices.AdminService, userRepo *usrRepo.UserRepository, alRepo *actionLogRepo.ActionLogRepository, pRepo *paymentRepo.PaymentRepository, bRepo *bucketRepo.BucketRepository) *AdminHandler {
	return &AdminHandler{Svc: svc, UserRepo: userRepo, ActionLogRepo: alRepo, PaymentRepo: pRepo, BucketRepo: bRepo}
}

func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, err := h.Svc.GetStats()
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, stats)
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	users, total, err := h.UserRepo.ListUsers(page, limit, c.Query("search"), c.Query("tier"))
	if err != nil {
		respond.Err(c, apperr.Internal("failed to list users", err))
		return
	}

	type userRow struct {
		UserID       string  `json:"user_id"`
		Username     string  `json:"username"`
		Name         string  `json:"name"`
		Email        string  `json:"email"`
		Tier         string  `json:"tier"`
		Role         string  `json:"role"`
		AuthProvider string  `json:"auth_provider"`
		CreatedAt    *string `json:"created_at"`
	}

	rows := make([]userRow, 0, len(users))
	for _, u := range users {
		var ca *string
		if u.CreatedAt != nil {
			s := u.CreatedAt.Format("2006-01-02 15:04:05")
			ca = &s
		}
		rows = append(rows, userRow{
			UserID: u.UserID, Username: u.Username, Name: u.Name,
			Email: u.Email, Tier: u.Tier, Role: u.Role,
			AuthProvider: u.AuthProvider, CreatedAt: ca,
		})
	}

	respond.OK(c, gin.H{"users": rows, "total": total, "page": page, "limit": limit})
}

func (h *AdminHandler) SetUserTier(c *gin.Context) {
	var body struct {
		Tier string `json:"tier" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respond.Err(c, apperr.BadRequest("tier is required"))
		return
	}
	if err := h.UserRepo.SetUserTier(c.Param("id"), body.Tier); err != nil {
		respond.Err(c, apperr.Internal("failed to update tier", err))
		return
	}
	respond.OK(c, gin.H{"message": "tier updated"})
}

func (h *AdminHandler) SetUserRole(c *gin.Context) {
	var body struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		respond.Err(c, apperr.BadRequest("role is required"))
		return
	}
	if body.Role != "user" && body.Role != "admin" {
		respond.Err(c, apperr.BadRequest("role must be 'user' or 'admin'"))
		return
	}
	if err := h.UserRepo.SetUserRole(c.Param("id"), body.Role); err != nil {
		respond.Err(c, apperr.Internal("failed to update role", err))
		return
	}
	respond.OK(c, gin.H{"message": "role updated"})
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	if err := h.UserRepo.HardDeleteUser(c.Param("id")); err != nil {
		respond.Err(c, apperr.Internal("failed to delete user", err))
		return
	}
	respond.NoContent(c)
}

// GET /admin/users/:id/payments
func (h *AdminHandler) GetUserPayments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	payments, total, err := h.PaymentRepo.ListByUserID(c.Param("id"), page, limit)
	if err != nil {
		respond.Err(c, apperr.Internal("failed to list payments", err))
		return
	}
	respond.OK(c, gin.H{"payments": payments, "total": total})
}

// GET /admin/users/:id/logs
func (h *AdminHandler) GetUserLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	logs, total, err := h.ActionLogRepo.ListByUserID(c.Param("id"), page, limit)
	if err != nil {
		respond.Err(c, apperr.Internal("failed to list logs", err))
		return
	}
	respond.OK(c, gin.H{"logs": logs, "total": total})
}

// GET /admin/users/:id/buckets
func (h *AdminHandler) GetUserBuckets(c *gin.Context) {
	buckets, err := h.BucketRepo.ListByUserID(c.Param("id"))
	if err != nil {
		respond.Err(c, apperr.Internal("failed to list buckets", err))
		return
	}
	respond.OK(c, gin.H{"buckets": buckets})
}
