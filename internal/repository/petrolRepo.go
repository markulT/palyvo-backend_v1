package repository

import (
	"context"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"palyvoua/internal/models"
	"palyvoua/tools"
)

type ProductRepo interface {
	DecreaseProductAmount(c context.Context, pid uuid.UUID,amount int) error
	GetProduct(c context.Context, pid uuid.UUID) (models.Product, error)
	GetAllProducts(c context.Context) ([]models.Product, error)
	SaveProduct(context.Context, *models.Product) error
	DeleteProduct(c context.Context, pid uuid.UUID) error
	UpdateProductAmount(c context.Context, pid uuid.UUID, amount int) error
}

func NewProductRepo() ProductRepo {
	return &defaultProductRepo{}
}

type defaultProductRepo struct {}

func (pr *defaultProductRepo) GetAllProducts(c context.Context) ([]models.Product, error) {
	return nil,nil
}

func (pr *defaultProductRepo) UpdateProductAmount(c context.Context, pid uuid.UUID, amount int) error {
	//TODO implement me
	return nil
}

func (pr *defaultProductRepo) DeleteProduct(c context.Context, pid uuid.UUID) error {
	return nil
}

func (pr *defaultProductRepo) DecreaseProductAmount(c context.Context,pID uuid.UUID,amount int) error {

	productCollection := tools.DB.Collection("product")

	session, err := tools.DB.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(c)

	err = mongo.WithSession(c, session, func(sc mongo.SessionContext) error {
		err :=  session.StartTransaction()
		if err != nil {
			return err
		}

		p, err := pr.GetProduct(sc, pID)

		if err != nil {
			return err
		}

		if p.Amount < amount {
			return session.AbortTransaction(sc)
		}

		_, err = productCollection.UpdateOne(sc, bson.M{"_id":pID}, bson.M{"amount":p.Amount-amount})

		return nil
	})
	return nil
}

func (pr *defaultProductRepo) GetProduct(c context.Context, pid uuid.UUID) (models.Product, error) {
	return models.Product{}, nil
}

func (pr *defaultProductRepo) SaveProduct(c context.Context, p *models.Product) error {

	return nil
}