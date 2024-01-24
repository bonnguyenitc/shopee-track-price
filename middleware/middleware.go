package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bonnguyenitc/shopee-stracks/back-end-go/common"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/database"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type ConditionAuth struct {
	NeedVerify bool
}

// for use on route (using a http.HandlerFunc)
func AuthMiddleware(next http.HandlerFunc, condition ConditionAuth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bearToken := strings.ReplaceAll(r.Header.Get("Authorization"), "Bearer ", "")

		claims := &Claims{}

		var secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

		tkn, err := jwt.ParseWithClaims(bearToken, claims, func(token *jwt.Token) (any, error) {
			return secretKey, nil
		})

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusUnauthorized, common.UnauthorizedCode, common.UnauthorizedMsg))
			return
		}

		if !tkn.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusUnauthorized, common.UnauthorizedCode, common.UnauthorizedMsg))
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
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusUnauthorized, common.UnauthorizedCode, common.UnauthorizedMsg))
			return
		}

		if !user.Verified && condition.NeedVerify {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.EmailNotVerifiedCode, common.EmailNotVerifiedMsg))
			return
		}

		// write user_id to context
		ctx = context.WithValue(r.Context(), "user_id", userID)

		next(w, r.WithContext(ctx))
	}
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Method: %s, URL: %s, RemoteAddr: %s\n", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
