package models

import "github.com/google/uuid"

const (
	ACTIVATED = "ACTIVATED"
	NOT_ACTIVATED = "NOT_ACTIVATED"
	USED = "USED"
)

type Ticket struct {
	CreatedAt int `json:"createdAt" bson:"createdAt"`
	ExpiresAt int `json:"expiresAt" bson:"expiresAt"`
	ID uuid.UUID `bson:"_id" json:"id"`
	secret string `bson:"secret"`
	UserId uuid.UUID `bson:"userId" json:"userId"`
	Status string `json:"status" bson:"status"`
	Amount int `json:"amount" bson:"amount"`
	PaymentID string `json:"paymentId" bson:"paymentId"`
	ProductTicketID uuid.UUID `bson:"productTicketId" json:"productTicketId"`
}

func (t *Ticket) GetSecret() string {
	return t.secret
}

func (t *Ticket) SetSecret(s string) {
	t.secret = s
}
