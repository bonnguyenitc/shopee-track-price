package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bonnguyenitc/shopee-stracks/back-end-go/common"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/database"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/templates"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/utils"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type UserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type VerifiedEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload UserRequest
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerErrorCode, common.InternalServerMsg))
		return
	}

	// Validate form data
	validate := validator.New()
	err = validate.Struct(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerErrorCode, common.InternalServerMsg))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoUserRepo := database.NewMongoUserRepository(database.MongoDB.Collection(database.UserCollectionName))

	userService := database.NewUserService(mongoUserRepo)

	user, _ := userService.FindByEmail(ctx, payload.Email)

	if user.Email != "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.EmailExistCode, common.EmailExistMsg))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerErrorCode, common.InternalServerMsg))
		return
	}

	newUser, err := userService.Insert(ctx, database.User{
		Email:    payload.Email,
		Role:     database.USER_ROLE,
		Verified: false,
		Password: string(hashedPassword),
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerErrorCode, common.InternalServerMsg))
		return
	}

	if newUser != nil {
		// generate token
		token, err := utils.GenerateTokenVerifyEmail()
		// save token to database
		if err == nil {
			mongoTokenRepo := database.NewMongoTokenRepository(database.MongoDB.Collection(database.TokenCollectionName))
			tokenService := database.NewTokenService(mongoTokenRepo)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			tokenService.Insert(ctx, database.Token{
				Token:     token,
				Type:      database.VerifyToken,
				ExpiredAt: time.Now().Add(6 * time.Hour),
				UserId:    newUser.(primitive.ObjectID),
			})

			log.Println(utils.SendEmail(payload.Email, templates.CreateEmailSendTokenVerifyUserTemplate(templates.InfoEmailSendTokenVerifyUser{
				Email:    payload.Email,
				UrlToken: fmt.Sprintf("%s/verify-email?token=%s", os.Getenv("BASE_URL"), token),
				Title:    "Verify your email",
			})))
		}

		json.NewEncoder(w).Encode(common.ResponseApi{
			Status:  http.StatusOK,
			Message: "Create new user success!",
		})
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.CanNotCreateUserCode, common.CanNotCreateUserMsg))
}

type Claims struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var payload LoginRequest
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerErrorCode, common.InternalServerMsg))
		return
	}

	// Validate form data
	validate := validator.New()
	err = validate.Struct(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerErrorCode, common.InternalServerMsg))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoUserRepo := database.NewMongoUserRepository(database.MongoDB.Collection(database.UserCollectionName))

	userService := database.NewUserService(mongoUserRepo)

	user, _ := userService.FindByEmail(ctx, payload.Email)

	if user.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.EmailNotFoundCode, common.EmailNotFoundMsg))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.PasswordWrongCode, common.PasswordWrongMsg))
		return
	}

	// Create jwt token
	var secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

	claims := &Claims{
		ID:   user.ID.Hex(),
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 60 * 24 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(secretKey)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerErrorCode, common.InternalServerMsg))
		return
	}

	json.NewEncoder(w).Encode(common.ResponseApi{
		Status:  http.StatusOK,
		Message: "Login success!",
		Metadata: map[string]any{
			"token": tokenString,
		},
	})
}

func verifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	var payload VerifiedEmailRequest
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerErrorCode, common.InternalServerMsg))
		return
	}

	// Validate form data
	validate := validator.New()
	err = validate.Struct(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerErrorCode, common.InternalServerMsg))
		return
	}

	// check token exist in database
	mongoTokenRepo := database.NewMongoTokenRepository(database.MongoDB.Collection(database.TokenCollectionName))
	tokenService := database.NewTokenService(mongoTokenRepo)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tokenObj, err := tokenService.FindOneByFilter(ctx, bson.M{
		"token": payload.Token,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerErrorCode, common.InternalServerMsg))
		return
	}

	userID := tokenObj.User.Map()["$id"].(primitive.ObjectID).Hex()

	// check token expired
	if tokenObj.ExpiredAt.Before(time.Now()) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.TokenVerifyExpiredCode, common.TokenVerifyExpiredMsg))
		return
	}

	// update user to database
	mongoUserRepo := database.NewMongoUserRepository(database.MongoDB.Collection(database.UserCollectionName))
	userService := database.NewUserService(mongoUserRepo)
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = userService.Update(ctx, userID, bson.M{
		"verified": true,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerErrorCode, common.InternalServerMsg))
		return
	}

	// remove token from database
	done, err := tokenService.Remove(ctx, bson.M{
		"token": payload.Token,
	})

	if err != nil || !done {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerErrorCode, common.InternalServerMsg))
		return
	}

	json.NewEncoder(w).Encode(common.ResponseApi{
		Status:  http.StatusOK,
		Message: "Verify email success!",
	})
}

func SetupUsersApiRoutes(router *mux.Router) {
	router.HandleFunc("/api/register", createUserHandler).Methods("POST")
	router.HandleFunc("/api/login", loginHandler).Methods("POST")
	router.HandleFunc("/api/verify", verifyEmailHandler).Methods("POST")
}
