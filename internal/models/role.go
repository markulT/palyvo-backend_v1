package models

import "github.com/google/uuid"

type Role struct {
	ID uuid.UUID `json:"id" bson:"_id"`
	Name string `json:"name" bson:"name"`
	AuthorityLevel int `json:"authorityLevel" bson:"authorityLevel"`
}

type RoleFactory struct {

}

func (rl *RoleFactory) Create(name string, level int) (*Role, error) {
	roleID, err := uuid.NewRandom()
	if err != nil {
		return &Role{}, err
	}
	role := Role{
		ID:             roleID,
		Name:           name,
		AuthorityLevel: level,
	}
	return &role, err
}
