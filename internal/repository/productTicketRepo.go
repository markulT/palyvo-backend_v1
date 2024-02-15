package repository

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"palyvoua/internal/models"
	"palyvoua/tools"
)

type ProductTicketRepo interface {
	GetAllProductTickets(c context.Context) ([]models.ProductTicket, error)
	GetByID(c context.Context,id uuid.UUID) (models.ProductTicket, error)
	GetByOperator(c context.Context, op string) ([]models.ProductTicket, error)
	SaveProductTicket(c context.Context,ticket *models.ProductTicket) error
	UpdateStripeProductID(c context.Context,prID uuid.UUID, stripeProductID string) (models.ProductTicket,error)
	DeleteProductTicket(c context.Context,id uuid.UUID) error
	UpdateProductTicket(c context.Context,id uuid.UUID, ticket *models.ProductTicket) error
	GetByStripeProductID(context.Context,string) (models.ProductTicket, error)
}

func NewProductTicketRepo() ProductTicketRepo {
	return &defaultProductTicketRepo{}
}

type defaultProductTicketRepo struct {}

func (d *defaultProductTicketRepo) GetByOperator(c context.Context, op string) ([]models.ProductTicket, error) {
	var productTickets []models.ProductTicket
	var err error
	productTicketCollection := tools.DB.Collection("productTickets")
	cursor, err := productTicketCollection.Find(c, bson.M{"operator":op})
	if cursor.Err() !=nil {
		return nil,cursor.Err()
	}
	for cursor.Next(c) {
		var productTicket models.ProductTicket
		if err=cursor.Decode(&productTicket);err!=nil {
			return nil, err
		}
		productTickets = append(productTickets, productTicket)
	}
	return productTickets, nil
}

func (d *defaultProductTicketRepo) GetAllProductTickets(c context.Context) ([]models.ProductTicket, error) {
	var productTickets []models.ProductTicket
	var err error
	productTicketCollection := tools.DB.Collection("productTickets")
	cursor, err := productTicketCollection.Find(c, bson.M{})
	if err != nil {
		return nil, err
	}
	if cursor.Err() !=nil {
		return nil,cursor.Err()
	}
	for cursor.Next(c) {
		var productTicket models.ProductTicket
		if err=cursor.Decode(&productTicket);err!=nil {
			return nil, err
		}
		productTickets = append(productTickets, productTicket)
	}
	return productTickets, nil
}

func (d *defaultProductTicketRepo) GetByID(c context.Context, id uuid.UUID) (models.ProductTicket, error) {
	var productTicker models.ProductTicket
	var err error
	productTicketCollection := tools.DB.Collection("productTickets")
	res := productTicketCollection.FindOne(c, bson.M{"_id":id})
	if res.Err() != nil {
		return models.ProductTicket{}, err
	}
	if err=res.Decode(&productTicker);err!=nil {
		return models.ProductTicket{}, err
	}
	return productTicker, nil

}

func (d *defaultProductTicketRepo) SaveProductTicket(c context.Context, ticket *models.ProductTicket) error {
	var err error
	productTicketCollection := tools.DB.Collection("productTickets")
	_, err = productTicketCollection.InsertOne(c, *ticket)
	return err
}

func (d *defaultProductTicketRepo) UpdateStripeProductID(c context.Context, prID uuid.UUID, stripeProductID string) (models.ProductTicket, error) {
	var updatedProductTicket models.ProductTicket
	var err error
	productTicketCollection := tools.DB.Collection("productTickets")

	_, err = productTicketCollection.UpdateByID(c, prID, bson.M{"$set":bson.M{"stripeProductId":stripeProductID}})
	if err != nil {
		return models.ProductTicket{},err
	}

	updatedProductTicket, err = d.GetByID(c, prID)
	if err != nil {
		return models.ProductTicket{}, err
	}
	return updatedProductTicket, nil
}

func (d *defaultProductTicketRepo) DeleteProductTicket(c context.Context, id uuid.UUID) error {
	var err error
	productTicketCollection := tools.DB.Collection("productTickets")
	_,err = productTicketCollection.DeleteOne(c , bson.M{"_id":id})
	return err
}

func (d *defaultProductTicketRepo) UpdateProductTicket(c context.Context, id uuid.UUID, ticket *models.ProductTicket) error {
	var err error
	productTicketCollection := tools.DB.Collection("productTickets")

	_, err = productTicketCollection.UpdateByID(c, id, bson.M{"$set":*ticket})
	return err

}

func (d *defaultProductTicketRepo) GetByStripeProductID(ctx context.Context, s string) (models.ProductTicket, error) {
	var pt models.ProductTicket
	var err error
	productTicketCollection := tools.DB.Collection("productTickets")
	fmt.Println("searching by this stripe id")
	fmt.Println(s)
	res := productTicketCollection.FindOne(ctx, bson.M{"stripeProductId":s})
	if res.Err() != nil {
		return models.ProductTicket{}, res.Err()
	}
	if err = res.Decode(&pt);err!=nil {
		return models.ProductTicket{}, err
	}
	return pt, nil
}


