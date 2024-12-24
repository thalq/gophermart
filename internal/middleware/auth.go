package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/thalq/gopher_mart/internal/constants"
	"github.com/thalq/gopher_mart/internal/models"
)

func AuthMiddleware(jwtSecret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, err := r.Cookie("Authorization")
			if err == nil {
				claims := &models.Claims{}
				token, err := jwt.ParseWithClaims(tokenString.Value, claims,
					func(t *jwt.Token) (interface{}, error) {
						if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
							return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
						}
						return []byte(jwtSecret), nil
					})
				if err != nil {
					http.Error(w, "Failed to parse token", http.StatusInternalServerError)
					return
				}
				if !token.Valid {
					http.Error(w, "Token is not valid", http.StatusUnauthorized)
					return
				}
				ctx := context.WithValue(r.Context(), constants.UserIDKey, claims.UserID)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}
