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
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/middleware"
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
type SendTokenRequest struct {
	Email string `json:"email" validate:"required,email"`
	Type  string `json:"type" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type VerifiedEmailRequest struct {
	Token string `json:"token" validate:"required"`
	Type  string `json:"type" validate:"required"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
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
				Type:      database.VerifyEmail,
				ExpiredAt: time.Now().Add(6 * time.Hour),
				UserId:    newUser.(primitive.ObjectID),
			})

			log.Println(utils.SendEmail(payload.Email, templates.CreateEmailSendTokenVerifyUserTemplate(templates.InfoEmailSendTokenVerifyUser{
				Email:    payload.Email,
				UrlToken: fmt.Sprintf("%s/verify-email/%s", os.Getenv("BASE_URL"), token),
				Title:    "Verify your email",
			})))
		}

		json.NewEncoder(w).Encode(common.ResponseApi{
			Status:   http.StatusOK,
			Message:  "Create new user success!",
			Metadata: true,
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
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.EmailOrPasswordWrongCode, common.EmailOrPasswordWrongMsg))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.EmailOrPasswordWrongCode, common.EmailOrPasswordWrongMsg))
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
			"access_token": tokenString,
		},
	})
}

func verifyTokenHandler(w http.ResponseWriter, r *http.Request) {
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
		"type":  payload.Type,
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
	if payload.Type == database.VerifyEmail {
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
	}

	// remove token from database
	if payload.Type == database.VerifyEmail {
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		done, err := tokenService.Remove(ctx, bson.M{
			"token": payload.Token,
		})

		if err != nil || !done {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerErrorCode, common.InternalServerMsg))
			return
		}
	}

	message := "Verify token verify email success!"
	if payload.Type == database.ResetPassword {
		message = "Verify token reset password success!"
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.ResponseApi{
		Status:   http.StatusOK,
		Message:  message,
		Metadata: true,
	})
}

func sendTokenResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var payload SendTokenRequest
	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerErrorCode, common.InternalServerMsg))
		return
	}

	token, err := utils.GenerateTokenVerifyEmail()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerErrorCode, common.InternalServerMsg))
		return
	}
	// save token to database
	mongoTokenRepo := database.NewMongoTokenRepository(database.MongoDB.Collection(database.TokenCollectionName))
	tokenService := database.NewTokenService(mongoTokenRepo)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// check email exist in database
	mongoUserRepo := database.NewMongoUserRepository(database.MongoDB.Collection(database.UserCollectionName))
	userService := database.NewUserService(mongoUserRepo)
	user, _ := userService.FindByEmail(ctx, payload.Email)

	if user.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerErrorCode, common.InternalServerMsg))
		return
	}

	_, err = tokenService.Insert(ctx, database.Token{
		Token:     token,
		Type:      database.ResetPassword,
		ExpiredAt: time.Now().Add(6 * time.Hour),
		UserId:    user.ID,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerErrorCode, common.InternalServerMsg))
		return
	}

	log.Println(utils.SendEmail(payload.Email, templates.CreateEmailSendTokenResetPasswordTemplate(templates.InfoEmailSendTokenResetPassword{
		Email:    payload.Email,
		UrlToken: fmt.Sprintf("%s/reset-password/%s", os.Getenv("BASE_URL"), token),
		Title:    "Reset your password",
	})))

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.ResponseApi{
		Status:   http.StatusOK,
		Message:  "Send token reset password success!",
		Metadata: true,
	})
}

func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var payload ResetPasswordRequest
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
		"type":  database.ResetPassword,
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerErrorCode, common.InternalServerMsg))
		return
	}

	err = userService.Update(ctx, userID, bson.M{
		"password":   string(hashedPassword),
		"updated_at": time.Now(),
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerErrorCode, common.InternalServerMsg))
		return
	}

	// remove token from database
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done, err := tokenService.Remove(ctx, bson.M{
		"token": payload.Token,
	})

	if err != nil || !done {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerErrorCode, common.InternalServerMsg))
		return
	}

	message := "Reset password success!"

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.ResponseApi{
		Status:   http.StatusOK,
		Message:  message,
		Metadata: true,
	})
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	// get user id from token
	userID := r.Context().Value("user_id").(string)

	// get user from database
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(common.ResponseApi{
		Status:  http.StatusOK,
		Message: "Get user success!",
		Metadata: map[string]any{
			// "user_id": user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func SetupUsersApiRoutes(router *mux.Router) {
	router.HandleFunc("/api/register", createUserHandler).Methods("POST")
	router.HandleFunc("/api/login", loginHandler).Methods("POST")
	router.HandleFunc("/api/verify-token", verifyTokenHandler).Methods("POST")
	router.HandleFunc("/api/send-token", sendTokenResetPasswordHandler).Methods("POST")
	router.HandleFunc("/api/reset-password", resetPasswordHandler).Methods("POST")
	router.HandleFunc("/api/user", middleware.AuthMiddleware(getUserHandler, middleware.ConditionAuth{
		NeedVerify: false,
	})).Methods("GET")
}
