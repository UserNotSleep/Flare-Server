package service

import (
	"context"
	"errors"
	"time"

	"Flare-server/internal/repository"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepo
	JWTKey   []byte
}

func NewAuthService(userRepo *repository.UserRepo, jwtKey []byte) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		JWTKey:   jwtKey,
	}
}

func (s *AuthService) Register(ctx context.Context, username, password string) (*repository.User, error) {
	_, err := s.userRepo.GetUserByUsername(ctx, username)
	if err == nil {
		return nil, errors.New("user already exists")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := repository.User{
		Username: username,
		Password: string(hashed),
	}

	return s.userRepo.SaveUser(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.MapClaims{
		"userID":   user.ID,
		"username": user.Username,
		"exp":      expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.JWTKey)
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.userRepo.AddToBlacklist(ctx, token)
}

func (s *AuthService) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	return s.userRepo.IsTokenBlacklisted(ctx, token)
}
