package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/bonnguyenitc/shopee-stracks/back-end-go/common"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/crawl"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/database"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/bonnguyenitc/shopee-stracks/back-end-go/utils"
	"github.com/gorilla/mux"
)

type TrackingRequest struct {
	Url string `json:"url"`
}

func trackingHandler(w http.ResponseWriter, r *http.Request) {
	// get user id from context
	userID := r.Context().Value("user_id").(string)
	// get body from request
	var payload TrackingRequest
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerMsg, common.InternalServerMsg))
		return
	}
	url := payload.Url

	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.ProductNotFoundCode, common.ProductNotFoundMessage))
		return
	}

	// check item exist in database
	productIdShopee, err := strconv.ParseInt(utils.GetProductIDFromUrl(url), 10, 64)

	if productIdShopee == 0 || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.ProductNotFoundCode, common.ProductNotFoundMessage))
		return
	}
	// get product from database
	mongoProductRepo := database.NewMongoProductRepository(database.MongoDB.Collection(database.ProductCollectionName))
	productService := database.NewProductService(mongoProductRepo)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	productExist, err := productService.FindByIdShopee(ctx, productIdShopee)

	if productExist.IDShopee == 0 && err != nil {
		// insert product to database
		// get products from url
		var shopId string = utils.GetShopIdFromString(url)

		products, err := crawl.GetProductsByShopID(shopId)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerMsg, common.InternalServerMsg))
			return
		}

		if len(products) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.ProductNotFoundCode, common.ProductNotFoundMessage))
			return
		}

		// save Shop if not exist
		var shopIdFromDB primitive.ObjectID
		if shopId != "" && len(products) > 0 {
			shopService := database.NewShopService(database.NewMongoShopRepository(database.MongoDB.Collection(database.ShopCollectionName)))
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			shopId, _ := strconv.ParseInt(shopId, 10, 64)
			shopDB, err := shopService.FindByShopShopeeId(ctx, shopId)
			if err != nil {
				ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				shopDB, err = shopService.Insert(ctx, database.Shop{
					ShopID:     shopId,
					Name:       products[0].ShopName,
					ShopRating: products[0].ShopRating,
				})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerMsg, common.InternalServerMsg))
					return
				} else {
					shopIdFromDB = shopDB.ID
				}
			} else {
				shopIdFromDB = shopDB.ID
			}
		}

		// find productId in product list
		found := false
		for _, product := range products {
			if product.IDShopee == productIdShopee {
				found = true
				break
			}
		}

		if !found {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.ProductNotFoundCode, common.ProductNotFoundMessage))
			return
		}

		// insert product to database
		done := make(chan bool)
		go insertProductToDatabase(products, shopIdFromDB, done)
		<-done
	}

	// insert tracking to database
	mongoTrackingRepo := database.NewMongoTrackingRepository(database.MongoDB.Collection(database.TrackingCollectionName))
	trackingService := database.NewTrackingService(mongoTrackingRepo)
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if productExist.IDShopee == 0 {
		productExist, err = productService.FindByIdShopee(ctx, productIdShopee)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.ProductNotFoundCode, common.ProductNotFoundMessage))
			return
		}

	}

	userIDObj, err := primitive.ObjectIDFromHex(userID)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusInternalServerError, common.InternalServerMsg, common.InternalServerMsg))
		return
	}

	// find product tracked
	tracking, err := trackingService.FindByIDShopee(ctx, productIdShopee)

	if err != nil {
		// insert tracking to database
		pp := database.Tracking{
			IDShopee:  productIdShopee,
			UserID:    userIDObj,
			Status:    true,
			ShopeeUrl: url,
		}

		if productExist.IDShopee != 0 {
			pp.Product = bson.D{
				{Key: "$ref", Value: database.ProductCollectionName}, {Key: "$id", Value: productExist.ID},
			}
		}

		trackingID, err := trackingService.Insert(ctx, pp)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.TrackingFailCode, common.TrackingFailMessage))
			return
		}

		// insert tracking condition to database
		trackingConditionService := database.NewTrackingConditionService(database.NewMongoTrackingConditionRepository(database.MongoDB.Collection(database.TrackingConditionCollectionName)))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err = trackingConditionService.Insert(ctx, database.TrackingCondition{
			TrackingID: trackingID.(primitive.ObjectID),
			Condition:  database.LESS_THAN,
			UserID:     userIDObj,
		})

		if err != nil {
			trackingService.Remove(ctx, trackingID.(primitive.ObjectID))
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.TrackingFailCode, common.TrackingFailMessage))
			return
		}

		json.NewEncoder(w).Encode(common.ResponseApi{
			Status:   http.StatusOK,
			Message:  common.TrackingSuccessMessage,
			Metadata: true,
		})
		return
	}

	if tracking.Product == nil && productExist.IDShopee != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, err = trackingService.Update(ctx, tracking.ID, bson.M{
			"product": bson.D{
				{Key: "$ref", Value: database.ProductCollectionName}, {Key: "$id", Value: productExist.ID},
			},
		})

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.TrackingFailCode, common.TrackingFailMessage))
			return
		}
	}

	if !tracking.Status {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.TrackingFailCode, common.TrackingFailMessage))
		return
	}

	// find user in trackings list of product
	exist, err := trackingService.CheckUserInTracking(ctx, tracking.ID, userIDObj)

	if err != nil {
		// insert user to trackings list of product
		tracking, err = trackingService.AddNewUserToTracking(ctx, tracking.ID, userIDObj)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.TrackingFailCode, common.TrackingFailMessage))
			return
		}

		// insert tracking condition to database
		trackingConditionService := database.NewTrackingConditionService(database.NewMongoTrackingConditionRepository(database.MongoDB.Collection(database.TrackingConditionCollectionName)))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err = trackingConditionService.Insert(ctx, database.TrackingCondition{
			TrackingID: tracking.ID,
			Condition:  database.LESS_THAN,
			UserID:     userIDObj,
		})

		if err != nil {
			trackingService.UnTracking(ctx, tracking.ID, userIDObj)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.TrackingFailCode, common.TrackingFailMessage))
			return
		}

		json.NewEncoder(w).Encode(common.ResponseApi{
			Status:   http.StatusOK,
			Message:  common.TrackingSuccessMessage,
			Metadata: true,
		})
	}

	if exist {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.TrackingExistCode, common.TrackingExistMessage))
	}
}

func insertProductToDatabase(products []database.Product, shopID primitive.ObjectID, done chan bool) {
	mongoProductRepo := database.NewMongoProductRepository(database.MongoDB.Collection(database.ProductCollectionName))
	productService := database.NewProductService(mongoProductRepo)
	var wg sync.WaitGroup
	for _, product := range products {
		wg.Add(1)
		go func(prod database.Product) {
			defer wg.Done()
			// insert product to database
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			existed, err := productService.FindByIdShopee(ctx, prod.IDShopee)
			if err != nil {
				prod.ShopID = shopID
				productService.Insert(ctx, prod)
			} else {
				productService.Update(ctx, existed.ID.Hex(), prod)
			}
		}(product)
	}
	wg.Wait()
	done <- true
}

func unTrackingHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	vars := mux.Vars(r)
	id := vars["id"]
	trackingService := database.NewTrackingService(database.NewMongoTrackingRepository(database.MongoDB.Collection(database.TrackingCollectionName)))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	idObj, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.UnTrackingFailCode, common.UnTrackingFailMsg))
		return
	}

	tracking, err := trackingService.FindById(ctx, idObj)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.UnTrackingFailCode, common.UnTrackingFailMsg))
		return
	}

	userIdObj, err := primitive.ObjectIDFromHex(userID)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.UnTrackingFailCode, common.UnTrackingFailMsg))
		return
	}

	exist, _ := trackingService.CheckUserInTracking(ctx, tracking.ID, userIdObj)

	if !exist {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.TrackingNotFoundCode, common.TrackingNotFoundMsg))
		return
	}

	_, err = trackingService.UnTracking(ctx, tracking.ID, userIdObj)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.UnTrackingFailCode, common.UnTrackingFailMsg))
		return
	}

	// remove tracking condition
	trackingConditionService := database.NewTrackingConditionService(database.NewMongoTrackingConditionRepository(database.MongoDB.Collection(database.TrackingConditionCollectionName)))
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	removed, err := trackingConditionService.RemoveByFilter(ctx, bson.M{
		"tracking": bson.D{{Key: "$ref", Value: database.TrackingCollectionName}, {Key: "$id", Value: tracking.ID}},
		"user":     bson.D{{Key: "$ref", Value: database.UserCollectionName}, {Key: "$id", Value: userIdObj}},
	})

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.UnTrackingFailCode, common.UnTrackingFailMsg))
		return
	}

	if removed {
		json.NewEncoder(w).Encode(common.ResponseApi{
			Status:  http.StatusOK,
			Message: common.UnTrackingSuccessCode,
		})
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(common.ReturnErrorApi(http.StatusBadRequest, common.UnTrackingFailCode, common.UnTrackingFailMsg))
}

func SetupTrackingsApiRoutes(router *mux.Router) {
	router.HandleFunc("/api/tracking-product", middleware.AuthMiddleware(trackingHandler, middleware.ConditionAuth{
		NeedVerify: true,
	})).Methods("POST")
	router.HandleFunc("/api/un-tracking-product/{id}", middleware.AuthMiddleware(unTrackingHandler, middleware.ConditionAuth{
		NeedVerify: true,
	})).Methods("GET")
}
