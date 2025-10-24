package repository

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type User struct {
	ID        string    `firestore:"id" json:"id"`
	Username  string    `firestore:"username" json:"username"`
	Password  string    `firestore:"password" json:"-"`
	CreatedAt time.Time `firestore:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `firestore:"updatedAt" json:"updatedAt"`
}

type TokenBlacklist struct {
	Token string `firestore:"token"`
}

type UserRepo struct {
	client        *firestore.Client
	usersColl     string
	blacklistColl string
}

func NewUserRepo(client *firestore.Client) *UserRepo {
	return &UserRepo{
		client:        client,
		usersColl:     "users",
		blacklistColl: "blacklisted_tokens",
	}
}

func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	iter := r.client.Collection(r.usersColl).Where("username", "==", username).Limit(1).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}

	var user User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}
	user.ID = doc.Ref.ID
	return &user, nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, userID string) (*User, error) {
	doc, err := r.client.Collection(r.usersColl).Doc(userID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	var user User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}
	user.ID = doc.Ref.ID
	return &user, nil
}

func (r *UserRepo) SaveUser(ctx context.Context, user User) (*User, error) {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	
	docRef, _, err := r.client.Collection(r.usersColl).Add(ctx, user)
	if err != nil {
		return nil, err
	}
	
	user.ID = docRef.ID
	return &user, nil
}

func (r *UserRepo) AddToBlacklist(ctx context.Context, token string) error {
	_, _, err := r.client.Collection(r.blacklistColl).Add(ctx, TokenBlacklist{Token: token})
	return err
}

func (r *UserRepo) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	iter := r.client.Collection(r.blacklistColl).Where("token", "==", token).Limit(1).Documents(ctx)
	defer iter.Stop()

	_, err := iter.Next()
	if err == iterator.Done {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
