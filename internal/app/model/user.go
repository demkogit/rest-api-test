package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Nickname     string             `json:"nickname,omitempty" bson:"nickname,omitempty"`
	RefreshToken string             `json:"refreshToken,omitempty" bson:"refreshToken,omitempty"`
	AccessToken  string             `json:"accessToken,omitempty" bson:"accessToken,omitempty"`
}
