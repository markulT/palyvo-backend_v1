package models

import "github.com/google/uuid"

type Product struct {
	Amount int `json:"amount" bson:"amount"`
	ID uuid.UUID `bson:"_id" json:"id"`
	Title string `json:"title" bson:"title"`
	Price int `json:"price" bson:"price"`
	Currency string `json:"currency" bson:"currency"`
}
