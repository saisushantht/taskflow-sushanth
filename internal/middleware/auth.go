package middleware

import (
	"context"
	"net/http"
	"strings"
	"taskflow/internal/util"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"
const UserEmailKey contextKey = "user_email"

func Authenticate(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				util.ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				util.ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				util.ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims["user_id"].(string))
			ctx = context.WithValue(ctx, UserEmailKey, claims["email"].(string))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
