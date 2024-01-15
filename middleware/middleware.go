package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bonnguyenitc/shopee-stracks/back-end-go/database"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// for use on route (using a http.HandlerFunc)
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bearToken := strings.ReplaceAll(r.Header.Get("Authorization"), "Bearer ", "")

		claims := &Claims{}

		var secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

		tkn, err := jwt.ParseWithClaims(bearToken, claims, func(token *jwt.Token) (any, error) {
			return secretKey, nil
		})

		if err != nil {
			http.Error(w, "Token invalid", http.StatusUnauthorized)
			return
		}

		if !tkn.Valid {
			http.Error(w, "Token invalid", http.StatusUnauthorized)
			return
		}

		userID := claims.ID

		// check user_id exist in database
		mongoUserRepo := database.NewMongoUserRepository(database.MongoDB.Collection(database.UserCollectionName))
		userService := database.NewUserService(mongoUserRepo)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		user, err := userService.FindById(ctx, userID)

		if err != nil {
			http.Error(w, "User not exist", http.StatusUnauthorized)
			return
		}

		if !user.Verified {
			http.Error(w, "User not verified", http.StatusUnauthorized)
			return
		}

		// write user_id to context
		ctx = context.WithValue(r.Context(), "user_id", userID)

		next(w, r.WithContext(ctx))
	}
}
