package models

import "github.com/google/uuid"

type Ticket struct {
	CreatedAt int `json:"createdAt" bson:"createdAt"`
	ExpiresAt int `json:"expiresAt" bson:"expiresAt"`
	ID uuid.UUID `bson:"_id" json:"id"`
	secret string `bson:"secret"`
	UserId uuid.UUID `bson:"userId" json:"userId"`
	Activated bool `json:"activated" bson:"activated"`
	Amount int `json:"amount" bson:"amount"`
}

func (t *Ticket) GetSecret() string {
	return t.secret
}

func (t *Ticket) SetSecret(s string) {
	t.secret = s
}
