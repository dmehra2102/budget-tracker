package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/dmehra2102/budget-tracker/internal/domain"
	"github.com/dmehra2102/budget-tracker/internal/utils"
	"github.com/dmehra2102/budget-tracker/pkg/response"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func AuthMiddleware(jwtAuth *utils.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, domain.ErrUnauthorized, http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Error(w, domain.ErrInvalidToken, http.StatusUnauthorized)
				return
			}

			token := parts[1]
			claims, err := jwtAuth.ValidateToken(token)
			if err != nil {
				if err == domain.ErrTokenExpired {
					response.Error(w, err, http.StatusUnauthorized)
					return
				}
				response.Error(w, domain.ErrInvalidToken, http.StatusUnauthorized)
				return
			}

			userID, err := primitive.ObjectIDFromHex(claims.UserID)
			if err != nil {
				response.Error(w, domain.ErrInvalidToken, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(ctx context.Context) (primitive.ObjectID, error) {
	userID, ok := ctx.Value(UserIDKey).(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, domain.ErrUnauthorized
	}
	return userID, nil
}
