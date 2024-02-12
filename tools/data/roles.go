package data

import (
	"context"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"palyvoua/internal/models"
	"palyvoua/tools"
)

func (s *defaultDBSeeder) createDefaultRoles() error {
	ctx := context.Background()
	roleFactory := models.RoleFactory{}
	var insertion []interface{}


	userRole, err := roleFactory.Create("ROLE_USER",1 )
	if err !=nil {
		return err
	}
	insertion = append(insertion, userRole)
	operatorRole, err := roleFactory.Create("ROLE_OPERATOR",2)
	if err !=nil {
		return err
	}
	insertion = append(insertion, operatorRole)
	adminRole, err := roleFactory.Create("ROLE_ADMIN",3)
	if err !=nil {
		return err
	}
	insertion = append(insertion, adminRole)

	collection := tools.DB.Collection("roles")
	_, err = collection.InsertMany(ctx, insertion)

	if err !=nil {
		return err
	}




	return nil
}

func (s *defaultDBSeeder) createUser(c context.Context,email string, password string, roleName string) (*models.User, error) {

	roleCollection :=tools.DB.Collection("roles")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err!=nil {
		return nil,err
	}

	userID, err := uuid.NewRandom()
	if err!=nil {
		return nil,err
	}

	var adminRole models.Role

	res := roleCollection.FindOne(c, bson.M{"name":roleName})
	if res.Err()!= nil {
		return nil,res.Err()
	}

	if err = res.Decode(&adminRole);err!=nil {
		return nil,err
	}

	admin := models.User{
		Email:      email,
		Password:   string(hashedPassword),
		ID:         userID,
		CustomerID: "",
		Role:       adminRole.ID,
	}

	return &admin, nil
}

func (s *defaultDBSeeder) createDefaultUsers() error {
	ctx := context.Background()
	var insert []interface{}
	var err error
	userCollection :=tools.DB.Collection("users")
	admin, err := s.createUser(ctx, "admin", "1234", "ROLE_ADMIN")
	if err != nil {
		return err
	}
	insert = append(insert, admin)

	_,err = userCollection.InsertMany(ctx, insert)
	if err!=nil {
		return err
	}
	return nil
}