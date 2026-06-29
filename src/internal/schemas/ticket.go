package schemas

import "time"

type TicketStatus string

const (
	TicketStatusOpen       TicketStatus = "open"
	TicketStatusInProgress TicketStatus = "in_progress"
	TicketStatusClosed     TicketStatus = "closed"
)

type Ticket struct {
	TicketID  string       `gorm:"primaryKey"`
	UserID    string       `gorm:"index;not null"`
	Subject   string       `gorm:"not null"`
	Category  string       `gorm:"default:'outros'"`
	Status    TicketStatus `gorm:"default:'open'"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

type TicketMessage struct {
	MessageID string     `gorm:"primaryKey"`
	TicketID  string     `gorm:"index;not null"`
	SenderID  string     `gorm:"not null"`
	Role      string     `gorm:"not null"` // "user" | "admin"
	Content   string     `gorm:"type:text;not null"`
	CreatedAt *time.Time
}

func (Ticket) TableName() string        { return "tickets" }
func (TicketMessage) TableName() string { return "ticket_messages" }
