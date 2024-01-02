package models

import "github.com/google/uuid"

type User struct {
	Email string `json:"email" bson:"email"`
	Password string `bson:"password"`
	ID uuid.UUID `bson:"_id" json:"id"`
	CustomerID string `json:"customerId" bson:"customerId"`
	Role uuid.UUID `json:"role" bson:"role"`
}
