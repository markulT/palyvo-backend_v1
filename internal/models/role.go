package models

import "github.com/google/uuid"

type Role struct {
	ID uuid.UUID `json:"id" bson:"_id"`
	Name string `json:"name" bson:"name"`
	AuthorityLevel int `json:"authorityLevel" bson:"authorityLevel"`
}
