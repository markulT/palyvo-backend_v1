package repository

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"palyvoua/internal/models"
	"palyvoua/tools"
)

type UserRepo interface {
	SaveUser(*models.User) error
	GetUserByEmail(ctx context.Context,email string) (models.User, error)
	UpdateCustomerIDByEmail(email string, cid string) error
	GetByCustomerID(cID string) (models.User,error)
}

func NewUserRepo() UserRepo {
	return &defaultUserRepo{}
}

type defaultUserRepo struct {

}

func (d *defaultUserRepo) GetByCustomerID(cID string) (models.User,error) {
	var user models.User
	var err error
	userCollection := tools.DB.Collection("users")
	res :=userCollection.FindOne(context.Background(), bson.M{"customerId":cID})
	if err = res.Err();err!=nil {
		fmt.Println(err)
		return models.User{}, err
	}
	if err = res.Decode(&user);err!=nil {
		fmt.Println(err)
		return models.User{}, err
	}
	return user,nil
}

func (d *defaultUserRepo) SaveUser(user *models.User) error {
	userCollection := tools.DB.Collection("users")
	_, err := userCollection.InsertOne(context.TODO(), *user)
	return err
}

func (d *defaultUserRepo) GetUserByEmail(ctx context.Context,email string) (models.User, error) {
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
