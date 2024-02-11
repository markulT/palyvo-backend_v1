package data

import (
	"context"
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