package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/bonnguyenitc/shopee-stracks/back-end-go/crawl"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/database"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/middleware"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/bonnguyenitc/shopee-stracks/back-end-go/utils"
	"github.com/gorilla/mux"
)

type ResponseApi struct {
	Status   int            `json:"status"`
	Message  string         `json:"message"`
	Metadata map[string]any `json:"metadata"`
}

type TrackingRequest struct {
	Url string `json:"url"`
}

func trackingHandler(w http.ResponseWriter, r *http.Request) {
	// get body from request
	var payload TrackingRequest
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	url := payload.Url
	// check item exist in database
	productId, err := strconv.ParseFloat(utils.GetProductIDFromUrl(url), 64)

	if productId == 0 || err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// get product from database
	mongoProductRepo := database.NewMongoProductRepository(database.MongoDB.Collection(database.ProductCollectionName))
	productService := database.NewProductService(mongoProductRepo)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	p, err := productService.FindByIdShopee(ctx, productId)

	if p.IDShopee == 0 && err != nil {
		// insert product to database
		// get products from url
		var shopId string = utils.GetShopIdFromString(url)
		products, err := crawl.GetProductsByShopID(shopId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(products) == 0 {
			http.Error(w, "Product not found!", http.StatusBadRequest)
			return
		}

		// save Shop if not exist
		var shopIdFromDB primitive.ObjectID
		if shopId != "" && len(products) > 0 {
			shopService := database.NewShopService(database.NewMongoShopRepository(database.MongoDB.Collection(database.ShopCollectionName)))
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			shopId, _ := strconv.ParseFloat(shopId, 64)
			shopDB, err := shopService.FindByShopShopeeId(ctx, float64(shopId))
			if err != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				shopDB, err := shopService.Insert(ctx, database.Shop{
					ShopID:     shopId,
					Name:       products[0].ShopName,
					ShopRating: products[0].ShopRating,
				})
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
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
			if product.IDShopee == productId {
				found = true
				break
			}
		}

		if !found {
			http.Error(w, "Product not found!", http.StatusBadRequest)
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

	userID := r.Context().Value("user_id").(string)

	userIDObj, err := primitive.ObjectIDFromHex(userID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tracking, err := trackingService.FindByIDShopee(ctx, productId, userIDObj)

	if tracking.Status {
		http.Error(w, "Product already tracking!", http.StatusBadRequest)
		return
	}

	if !tracking.Status && err == nil {
		trackingService.Update(ctx, tracking.ID.Hex(), database.Tracking{
			Status:    true,
			ShopeeUrl: url,
		})
		json.NewEncoder(w).Encode(ResponseApi{
			Status:  http.StatusOK,
			Message: "Tracking success!",
		})
		return
	} else {
		_, err = trackingService.Insert(ctx, database.Tracking{
			IDShopee:  productId,
			UserID:    userIDObj,
			Status:    true,
			ShopeeUrl: url,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(ResponseApi{
			Status:  http.StatusOK,
			Message: "Tracking success!",
		})
		return
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

func SetupTrackingsApiRoutes(router *mux.Router) {
	router.HandleFunc("/api/tracking", middleware.AuthMiddleware(trackingHandler)).Methods("POST")
}
