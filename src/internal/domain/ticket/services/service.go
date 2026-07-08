package ticketServices

import (
	"backend/src/external"
	"backend/src/internal/actionlog"
	"backend/src/internal/logger"
	"backend/src/internal/schemas"
	ticketRepo "backend/src/internal/repository/ticket"
	usrRepo "backend/src/internal/repository/user"
	"backend/src/pkg/apperr"
	"backend/src/utils"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

type TicketService struct {
	Repo     *ticketRepo.TicketRepository
	UserRepo *usrRepo.UserRepository
	Email    *external.EmailClient
}

func New(repo *ticketRepo.TicketRepository, userRepo *usrRepo.UserRepository, email *external.EmailClient) *TicketService {
	return &TicketService{Repo: repo, UserRepo: userRepo, Email: email}
}

func (s *TicketService) Create(userID, subject, category, content string) (*schemas.Ticket, error) {
	if category == "" {
		category = "outros"
	}
	t := &schemas.Ticket{
		TicketID: utils.GenerateULID(),
		UserID:   userID,
		Subject:  subject,
		Category: category,
		Status:   schemas.TicketStatusOpen,
	}
	if err := s.Repo.CreateTicket(t); err != nil {
		return nil, apperr.Internal("erro ao criar ticket", err)
	}

	msg := &schemas.TicketMessage{
		MessageID: utils.GenerateULID(),
		TicketID:  t.TicketID,
		SenderID:  userID,
		Role:      "user",
		Content:   content,
	}
	if err := s.Repo.CreateMessage(msg); err != nil {
		return nil, apperr.Internal("erro ao criar mensagem", err)
	}

	actionlog.Log(userID, "ticket.create", "ticket", t.TicketID, "", nil)

	// notify admin by email
	go func() {
		adminEmail := os.Getenv("ADMIN_EMAIL")
		if adminEmail == "" {
			return
		}
		user, err := s.UserRepo.GetUserByID(userID)
		if err != nil {
			return
		}
		appURL := os.Getenv("APP_URL")
		if err := s.Email.RenderAndSend(adminEmail, fmt.Sprintf("[FiveKeepr] Novo ticket: %s", subject), "ticket_opened", map[string]any{
			"UserName":     user.Name,
			"UserEmail":    user.Email,
			"Subject":      subject,
			"FirstMessage": content,
			"AdminURL":     fmt.Sprintf("%s/admin", appURL),
		}); err != nil {
			logger.Warn("failed to send ticket_opened email", zap.Error(err))
		}
	}()

	return t, nil
}

func (s *TicketService) Reply(ticketID, senderID, role, content string) (*schemas.TicketMessage, error) {
	ticket, err := s.Repo.GetTicket(ticketID)
	if err != nil {
		return nil, apperr.NotFound("ticket não encontrado")
	}
	if ticket.Status == schemas.TicketStatusClosed {
		return nil, apperr.BadRequest("ticket está encerrado")
	}
	if role == "user" && ticket.UserID != senderID {
		return nil, apperr.Forbidden("acesso negado")
	}

	msg := &schemas.TicketMessage{
		MessageID: utils.GenerateULID(),
		TicketID:  ticketID,
		SenderID:  senderID,
		Role:      role,
		Content:   content,
	}
	if err := s.Repo.CreateMessage(msg); err != nil {
		return nil, apperr.Internal("erro ao salvar resposta", err)
	}

	// if admin replying, update status and email the user
	if role == "admin" {
		_ = s.Repo.SetStatus(ticketID, schemas.TicketStatusInProgress)
		now := time.Now()
		ticket.UpdatedAt = &now

		actionlog.Log(senderID, "ticket.reply", "ticket", ticketID, "", nil)

		go func() {
			user, err := s.UserRepo.GetUserByID(ticket.UserID)
			if err != nil {
				return
			}
			appURL := os.Getenv("APP_URL")
			if err := s.Email.RenderAndSend(user.Email, fmt.Sprintf("[FiveKeepr] Resposta ao ticket: %s", ticket.Subject), "ticket_reply", map[string]any{
				"Subject":      ticket.Subject,
				"ReplyContent": content,
				"TicketURL":    fmt.Sprintf("%s/suporte", appURL),
			}); err != nil {
				logger.Warn("failed to send ticket_reply email", zap.Error(err))
			}
		}()
	}

	return msg, nil
}

func (s *TicketService) Close(ticketID, adminID string) error {
	ticket, err := s.Repo.GetTicket(ticketID)
	if err != nil {
		return apperr.NotFound("ticket não encontrado")
	}
	if err := s.Repo.SetStatus(ticketID, schemas.TicketStatusClosed); err != nil {
		return apperr.Internal("erro ao encerrar ticket", err)
	}

	actionlog.Log(adminID, "ticket.close", "ticket", ticketID, "", nil)

	go func() {
		user, err := s.UserRepo.GetUserByID(ticket.UserID)
		if err != nil {
			return
		}
		appURL := os.Getenv("APP_URL")
		if err := s.Email.RenderAndSend(user.Email, fmt.Sprintf("[FiveKeepr] Ticket encerrado: %s", ticket.Subject), "ticket_closed", map[string]any{
			"Subject":    ticket.Subject,
			"SupportURL": fmt.Sprintf("%s/suporte", appURL),
		}); err != nil {
			logger.Warn("failed to send ticket_closed email", zap.Error(err))
		}
	}()

	return nil
}

func (s *TicketService) GetWithMessages(ticketID, userID, role string) (map[string]any, error) {
	var ticket *schemas.Ticket
	var err error
	if role == "admin" {
		ticket, err = s.Repo.GetTicket(ticketID)
	} else {
		ticket, err = s.Repo.GetTicketForUser(ticketID, userID)
	}
	if err != nil {
		return nil, apperr.NotFound("ticket não encontrado")
	}

	msgs, err := s.Repo.ListMessages(ticketID)
	if err != nil {
		return nil, apperr.Internal("erro ao listar mensagens", err)
	}

	return map[string]any{"ticket": ticket, "messages": msgs}, nil
}

func (s *TicketService) ListForUser(userID string, page, limit int) ([]schemas.Ticket, int64, error) {
	return s.Repo.ListByUserID(userID, page, limit)
}

func (s *TicketService) ListAll(page, limit int, status string) ([]schemas.Ticket, int64, error) {
	return s.Repo.ListAll(page, limit, status)
}
