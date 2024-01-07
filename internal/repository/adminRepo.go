package repository

import (
	"context"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"palyvoua/internal/models"
	"palyvoua/tools"
)

type AdminRepo interface {
	GetAllRoles() ([]models.Role, error)
	SaveRole(role models.Role) error
	DeleteRoleByID(roleID uuid.UUID) error
	GetRoleByID(uuid2 uuid.UUID) (models.Role, error)
}

func NewAdminRepo() AdminRepo {
	return &defaultAdminRepo{}
}

type defaultAdminRepo struct {

}

func (d *defaultAdminRepo) GetAllRoles() ([]models.Role, error) {
	var roles []models.Role
	roleCollection := tools.DB.Collection("roles")

	cursor, err := roleCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	if cursor.Err() != nil {
		return nil, cursor.Err()
	}

	if cursor.Next(context.TODO()) {
		var role models.Role
		if err := cursor.Decode(&role);err!=nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil

}

func (d *defaultAdminRepo) SaveRole(role models.Role) error {
	roleCollection := tools.DB.Collection("roles")
	_, err := roleCollection.InsertOne(context.TODO(), role)
	return err
}

func (d *defaultAdminRepo) DeleteRoleByID(roleID uuid.UUID) error {
	roleCollection := tools.DB.Collection("roles")
	_,err := roleCollection.DeleteOne(context.TODO(), bson.M{"_id":roleID})
	return err
}

func (d *defaultAdminRepo) GetRoleByID(uuid2 uuid.UUID) (models.Role, error) {
	var role models.Role
	roleCollection := tools.DB.Collection("roles")
	err := roleCollection.FindOne(context.TODO(), bson.M{"_id":uuid2}).Decode(&role)
	if err != nil {
		return models.Role{}, err
	}
	return role, nil
}
