package store

import (
	"context"
	"time"

	"gihub.com/demkogit/rest-api/internal/app/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRepository struct {
	store *Store
}

func (r *UserRepository) FindById(id string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	u := &model.User{}
	collection := r.store.client.Database(r.store.config.DatabaseName).Collection("users")
	idHEX, _ := primitive.ObjectIDFromHex(id)
	err := collection.FindOne(ctx, bson.M{"_id": idHEX}).Decode(&u)
	if err != nil {
		return u, err
	}
	return u, nil
}

func (r *UserRepository) FindByRefreshToken(refreshToken string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	u := &model.User{}
	collection := r.store.client.Database(r.store.config.DatabaseName).Collection("users")
	err := collection.FindOne(ctx, bson.M{"refreshToken": refreshToken}).Decode(&u)
	if err != nil {
		return u, err
	}
	return u, nil
}

func (r *UserRepository) UpdateRefreshToken(oldToken string, newToken string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := r.store.client.Database(r.store.config.DatabaseName).Collection("users")

	_, err := collection.UpdateOne(ctx, bson.M{"refreshToken": oldToken}, bson.M{"$set": bson.M{"refreshToken": newToken}})
	if err != nil {
		return err
	}

	return nil
}
