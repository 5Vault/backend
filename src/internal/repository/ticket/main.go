package ticketRepo

import (
	"backend/src/internal/schemas"

	"gorm.io/gorm"
)

type TicketRepository struct {
	DB *gorm.DB
}

func New(db *gorm.DB) *TicketRepository {
	return &TicketRepository{DB: db}
}

func (r *TicketRepository) CreateTicket(t *schemas.Ticket) error {
	return r.DB.Create(t).Error
}

func (r *TicketRepository) CreateMessage(m *schemas.TicketMessage) error {
	return r.DB.Create(m).Error
}

func (r *TicketRepository) GetTicket(ticketID string) (*schemas.Ticket, error) {
	var t schemas.Ticket
	return &t, r.DB.Where("ticket_id = ?", ticketID).First(&t).Error
}

func (r *TicketRepository) GetTicketForUser(ticketID, userID string) (*schemas.Ticket, error) {
	var t schemas.Ticket
	return &t, r.DB.Where("ticket_id = ? AND user_id = ?", ticketID, userID).First(&t).Error
}

func (r *TicketRepository) ListByUserID(userID string, page, limit int) ([]schemas.Ticket, int64, error) {
	var tickets []schemas.Ticket
	var total int64
	r.DB.Model(&schemas.Ticket{}).Where("user_id = ?", userID).Count(&total)
	offset := (page - 1) * limit
	err := r.DB.Where("user_id = ?", userID).Order("created_at desc").Offset(offset).Limit(limit).Find(&tickets).Error
	return tickets, total, err
}

func (r *TicketRepository) ListAll(page, limit int, status string) ([]schemas.Ticket, int64, error) {
	var tickets []schemas.Ticket
	var total int64
	q := r.DB.Model(&schemas.Ticket{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q.Count(&total)
	offset := (page - 1) * limit
	err := q.Order("created_at desc").Offset(offset).Limit(limit).Find(&tickets).Error
	return tickets, total, err
}

func (r *TicketRepository) ListMessages(ticketID string) ([]schemas.TicketMessage, error) {
	var msgs []schemas.TicketMessage
	err := r.DB.Where("ticket_id = ?", ticketID).Order("created_at asc").Find(&msgs).Error
	return msgs, err
}

func (r *TicketRepository) SetStatus(ticketID string, status schemas.TicketStatus) error {
	return r.DB.Model(&schemas.Ticket{}).Where("ticket_id = ?", ticketID).Update("status", status).Error
}
