package repository

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"palyvoua/internal/models"
	"palyvoua/tools"
)


type FindTicketParams struct {
	Operator *string
	FuelType *string
}

type ProductTicketRepo interface {
	GetAllProductTickets(c context.Context) ([]models.ProductTicket, error)
	GetByID(c context.Context,id uuid.UUID) (models.ProductTicket, error)
	GetByOperator(c context.Context, op string) ([]models.ProductTicket, error)
	FindByParams(c context.Context, params *FindTicketParams) ([]*models.ProductTicket, error)
	SaveProductTicket(c context.Context,ticket *models.ProductTicket) error
	UpdateStripeProductID(c context.Context,prID uuid.UUID, stripeProductID string) (models.ProductTicket,error)
	DeleteProductTicket(c context.Context,id uuid.UUID) error
	UpdateProductTicket(c context.Context,id uuid.UUID, ticket *models.ProductTicket) error
	GetByStripeProductID(context.Context,string) (models.ProductTicket, error)
	FindManyByProductID(ctx context.Context, productID uuid.UUID) ([]*models.ProductTicket, error)
	DeleteManyByProductID(ctx context.Context, productID uuid.UUID) error
}

func NewProductTicketRepo() ProductTicketRepo {

	repo := defaultProductTicketRepo{}
	repo.localCollection =tools.DB.Collection("productTickets")
	return &repo
}

type defaultProductTicketRepo struct {
	localCollection *mongo.Collection
}

func (d *defaultProductTicketRepo) FindByParams(c context.Context, params *FindTicketParams) ([]*models.ProductTicket, error) {
	var err error
	var productTickets []*models.ProductTicket
	filter := bson.M{}

	if params.Operator != nil {
		filter["seller"] = *params.Operator
	}

	if params.FuelType != nil {
		filter["fuelType"] = *params.FuelType
	}
	curs, err := d.localCollection.Find(c, filter)
	if err !=nil {
		return nil, err
	}
	defer curs.Close(c)

	for curs.Next(c) {
		var ticket models.ProductTicket
		if err = curs.Decode(&ticket);err!=nil{
			return nil, err
		}
		productTickets = append(productTickets, &ticket)
	}
	return productTickets,nil
}

func (d *defaultProductTicketRepo) DeleteManyByProductID(ctx context.Context, productID uuid.UUID) error {
	var err error

	_, err = d.localCollection.DeleteMany(ctx, bson.M{"productId":productID})
	if err !=nil {
		return err
	}
	return nil
}

func (d *defaultProductTicketRepo) FindManyByProductID(ctx context.Context, productID uuid.UUID) ([]*models.ProductTicket, error) {
	var productTickets []*models.ProductTicket

	curs, err := d.localCollection.Find(ctx, bson.M{"productId":productID})
	if err != nil {
		return nil, err
	}
	defer curs.Close(ctx)

	for curs.Next(ctx) {
		var pt models.ProductTicket
		if err=curs.Decode(&pt);err!=nil{
			return nil, err
		}
		productTickets = append(productTickets, &pt)
	}

	return productTickets,nil
}

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


