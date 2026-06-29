package ws

import (
	"backend/src/internal/logger"
	"backend/src/internal/notif"
	ticketRepo "backend/src/internal/repository/ticket"
	userRepo "backend/src/internal/repository/user"
	"backend/src/internal/schemas"
	"backend/src/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var authSvc = utils.NewAuthService()

type incomingMsg struct {
	Content string `json:"content"`
}

type OutgoingMsg struct {
	MessageID string `json:"message_id"`
	TicketID  string `json:"ticket_id"`
	SenderID  string `json:"sender_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// TicketWS handles GET /ws/ticket/:ticketId?token=<jwt>
func TicketWS(tRepo *ticketRepo.TicketRepository, uRepo *userRepo.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		ticketID := c.Param("ticketId")

		// Authenticate via query param (WS handshake can't set headers).
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
			return
		}
		claims, err := authSvc.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		user, err := uRepo.GetUserByID(claims.UserID)
		if err != nil || user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		userID := claims.UserID
		role := user.Role // "user" | "admin"

		// Verify ticket access.
		var ticket *schemas.Ticket
		if role == "admin" {
			ticket, err = tRepo.GetTicket(ticketID)
		} else {
			ticket, err = tRepo.GetTicketForUser(ticketID, userID)
		}
		if err != nil || ticket == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "ticket não encontrado"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Warn("ws: upgrade failed", zap.Error(err))
			return
		}

		client := &Client{
			conn:     conn,
			send:     make(chan []byte, 64),
			TicketID: ticketID,
			UserID:   userID,
			Role:     role,
		}

		Global.Register(client)
		defer Global.Unregister(client)

		go client.WritePump()

		// Read loop: parse → persist → broadcast.
		for {
			_, raw, err := conn.ReadMessage()
			if err != nil {
				break
			}

			var inc incomingMsg
			if err := json.Unmarshal(raw, &inc); err != nil || inc.Content == "" {
				continue
			}

			// Re-fetch ticket status in case it changed.
			fresh, ferr := tRepo.GetTicket(ticketID)
			if ferr != nil || fresh.Status == schemas.TicketStatusClosed {
				continue
			}

			now := time.Now()
			msg := &schemas.TicketMessage{
				MessageID: utils.GenerateULID(),
				TicketID:  ticketID,
				SenderID:  userID,
				Role:      role,
				Content:   inc.Content,
				CreatedAt: &now,
			}
			if err := tRepo.CreateMessage(msg); err != nil {
				logger.Warn("ws: failed to save message", zap.Error(err))
				continue
			}

			// Admin reply → update status + notify ticket owner.
			if role == "admin" {
				if fresh.Status == schemas.TicketStatusOpen {
					_ = tRepo.SetStatus(ticketID, schemas.TicketStatusInProgress)
				}
				notif.Create(
					fresh.UserID,
					"ticket_reply",
					fmt.Sprintf("Resposta no ticket: %s", fresh.Subject),
					inc.Content,
					ticketID,
				)
			}

			out, _ := json.Marshal(OutgoingMsg{
				MessageID: msg.MessageID,
				TicketID:  ticketID,
				SenderID:  userID,
				Role:      role,
				Content:   inc.Content,
				CreatedAt: now.Format(time.RFC3339),
			})
			Global.Broadcast(ticketID, out)
		}
	}
}
