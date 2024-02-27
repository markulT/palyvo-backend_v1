package dto

type TicketDto struct {
	CreatedAt *int `bson:"createdAt" json:"createdAt"`
	ExpiresAt *int `bson:"expiresAt" json:"expiresAt"`
	ID *string `json:"id" bson:"_id"`
	UserId *string `json:"userId" bson:"userId"`
	Status *string `json:"status" bson:"status"`
	Amount *int `json:"amount" bson:"amount"`
	PaymentID *string `json:"paymentId" bson:"paymentId"`
	ProductTicketId *string `json:"productTicketId" bson:"productTicketId"`
}

type ProductTicketDto struct {
	ProductID *string `json:"productId" bson:"productId"`
	Amount *int `json:"amount" bson:"amount"`
	Title *string `json:"title" bson:"title"`
	Price *int `json:"price" bson:"price"`
	Currency *string `json:"currency" bson:"currency"`
	StripeID *string `json:"stripeId" bson:"stripeId"`
	Seller *string `json:"seller" bson:"seller"`
	FuelType *string `json:"fuelType" bson:"fuelType"`
}

type FullTicketDto struct {
	TicketDto
	ProductTicket ProductTicketDto `json:"productTicket" bson:"productTicket"`
}

