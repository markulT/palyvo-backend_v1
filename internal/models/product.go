package models

import "github.com/google/uuid"

type Product struct {
	Amount int `json:"amount" bson:"amount"`
	ID uuid.UUID `bson:"_id" json:"id"`
	Title string `json:"title" bson:"title"`
	Price int `json:"price" bson:"price"`
	Currency string `json:"currency" bson:"currency"`
	Seller string `json:"seller" bson:"seller"`
	FuelType string `json:"fuelType" bson:"fuelType"`
}

type ProductTicket struct {
	ID uuid.UUID `bson:"_id" json:"id"`
	ProductID uuid.UUID `json:"productId" bson:"productId"`
	Amount int `json:"amount" bson:"amount"`
	Title string `json:"title" bson:"title"`
	Price int `json:"price" bson:"price"`
	Currency string `json:"currency" bson:"currency"`
	StripeID string `bson:"stripeProductId" json:"stripeProductId"`
}