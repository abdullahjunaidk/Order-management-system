package src

import (
	pass "common/helpers/password"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type authSeed struct {
	store AuthStore
}

func NewAuthSeed(store AuthStore) *authSeed {
	return &authSeed{store: store}
}

func (s *authSeed) SeedSuperAdmin(ctx context.Context) error {

	username := "admin"

	exists, err := s.store.CheckAdminExistence(ctx, username)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	hashedPassword, err := pass.HashPassword("12345678", 10)
	if err != nil {
		return err
	}

	admin := User{
		ID:                 primitive.NewObjectID(),
		Name:               "admin",
		Username:           "admin",
		Email:              "admin@example.com",
		PasswordHash:       hashedPassword,
		Phone:              7234567890,
		Incentive:          0,
		IsActive:           true,
		IsSuperAdmin:       true,
		PasswordResetToken: "",
		AccessToken:        "",
		RefreshToken:       "",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	_, err = s.store.RegisterUser(ctx, &admin)
	return err
}
