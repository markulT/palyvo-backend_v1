package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"palyvoua/internal/models"
	"palyvoua/tools"
)

type UserRepo interface {
	SaveUser(*models.User) error
	GetUserByEmail(email string) (models.User, error)
	UpdateCustomerIDByEmail(email string, cid string) error
}

func NewUserRepo() UserRepo {
	return &defaultUserRepo{}
}

type defaultUserRepo struct {

}

func (d *defaultUserRepo) SaveUser(user *models.User) error {
	userCollection := tools.DB.Collection("users")
	_, err := userCollection.InsertOne(context.TODO(), *user)
	return err
}

func (d *defaultUserRepo) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	userCollection := tools.DB.Collection("users")
	err := userCollection.FindOne(context.TODO(), bson.M{"email":email}).Decode(&user)
	return user,err
}

func (d *defaultUserRepo) UpdateCustomerIDByEmail(email string, cid string) error {
	userCollection := tools.DB.Collection("users")
	_, err := userCollection.UpdateOne(context.TODO(), bson.M{"email": email}, bson.M{"$set": bson.M{"customerId": cid}})
	return err
}
