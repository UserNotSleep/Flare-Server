package middleware

import (
	"context"
	"net/http"
	"strings"

	"Flare-server/internal/service"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserContextKey contextKey = "user"

func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				http.Error(w, "Bearer token required", http.StatusUnauthorized)
				return
			}

			isBlacklisted, err := authService.IsTokenBlacklisted(r.Context(), tokenString)
			if err != nil {
				http.Error(w, "Failed to validate token", http.StatusInternalServerError)
				return
			}
			if isBlacklisted {
				http.Error(w, "Token is blacklisted", http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return authService.JWTKey, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				userInfo := map[string]interface{}{
					"userID":   claims["userID"],
					"username": claims["username"],
				}
				ctx := context.WithValue(r.Context(), UserContextKey, userInfo)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}
		})
	}
}
