package jobs

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/bonnguyenitc/shopee-stracks/back-end-go/crawl"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/database"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/templates"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func crawlShop() {
	// get all shop
	shopService := database.NewShopService(database.NewMongoShopRepository(database.MongoDB.Collection(database.ShopCollectionName)))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	shops, err := shopService.FindAll(ctx)
	if err != nil {
		return
	}
	for _, shop := range shops {
		shopId := shop.ShopID
		idString := strconv.FormatInt(shopId, 10)
		products, err := crawl.GetProductsByShopID(idString)

		if err != nil {
			fmt.Println(err)
			continue
		}

		if len(products) == 0 {
			continue
		}

		priceService := database.NewPriceService(database.NewMongoPriceRepository(database.MongoDB.Collection(database.PriceCollectionName)))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		for _, product := range products {
			// check product exist in database
			productService := database.NewProductService(database.NewMongoProductRepository(database.MongoDB.Collection(database.ProductCollectionName)))
			prod, err := productService.FindByIdShopee(ctx, product.IDShopee)
			if err != nil {
				continue
			}
			now := time.Now()
			startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			endOfDay := startOfDay.Add(24 * time.Hour)

			//
			p, err := priceService.FindOneByFilter(ctx, bson.M{
				"product.$id": prod.ID,
				"created_at":  bson.M{"$gte": startOfDay, "$lt": endOfDay},
			})

			if err != nil {
				// insert product to database
				_, err := priceService.Insert(ctx, database.Price{
					ProductID:              prod.ID,
					Stock:                  product.Stock,
					Sold:                   product.Sold,
					HistoricalSold:         product.HistoricalSold,
					LikedCount:             product.LikedCount,
					CmtCount:               product.CmtCount,
					Price:                  product.Price,
					PriceMin:               product.PriceMin,
					PriceMax:               product.PriceMax,
					PriceMinBeforeDiscount: product.PriceMinBeforeDiscount,
					PriceMaxBeforeDiscount: product.PriceMaxBeforeDiscount,
					PriceBeforeDiscount:    product.PriceBeforeDiscount,
					RawDiscount:            product.RawDiscount,
				})
				if err != nil {
					continue
				}
			} else {
				// update product to database
				_, err := priceService.Update(ctx, p.ID.Hex(), database.Price{
					Stock:                  product.Stock,
					Sold:                   product.Sold,
					HistoricalSold:         product.HistoricalSold,
					LikedCount:             product.LikedCount,
					CmtCount:               product.CmtCount,
					Price:                  product.Price,
					PriceMin:               product.PriceMin,
					PriceMax:               product.PriceMax,
					PriceMinBeforeDiscount: product.PriceMinBeforeDiscount,
					PriceMaxBeforeDiscount: product.PriceMaxBeforeDiscount,
					PriceBeforeDiscount:    product.PriceBeforeDiscount,
					RawDiscount:            product.RawDiscount,
				})
				if err != nil {
					continue
				}
			}
		}
	}
}

func notifyPriceChangeJob() {
	// get all tracking
	trackingService := database.NewTrackingService(database.NewMongoTrackingRepository(database.MongoDB.Collection(database.TrackingCollectionName)))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	trackings, err := trackingService.FindAll(ctx, 20, 1)

	if err != nil {
		return
	}

	trackingsPassed := utils.Filter[database.Tracking](trackings.Data, func(tracking database.Tracking) bool {
		return tracking.Status
	})

	for _, tracking := range trackingsPassed {
		// get all price
		priceService := database.NewPriceService(database.NewMongoPriceRepository(database.MongoDB.Collection(database.PriceCollectionName)))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		productID := tracking.Product.Map()["$id"].(primitive.ObjectID)
		prices, err := priceService.FindByProductID(ctx, productID)

		if err != nil {
			continue
		}

		if len(prices) <= 1 {
			continue
		}

		// compare price
		// get all condition per condition
		conditionService := database.NewTrackingConditionService(database.NewMongoTrackingConditionRepository(database.MongoDB.Collection(database.TrackingConditionCollectionName)))

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			// for less than
			ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			conditions, err := conditionService.FindAllByFilterWithUser(ctx, bson.M{
				"tracking":  bson.D{{Key: "$ref", Value: database.TrackingCollectionName}, {Key: "$id", Value: tracking.ID}},
				"condition": database.LESS_THAN,
				"active":    true,
			})

			if err != nil {
				wg.Done()
				return
			}
			latestPrice := prices[0]
			previousPrice := prices[1]
			// check condition for less than
			if latestPrice.Price < previousPrice.Price {
				// send email to user if price less than condition
				for _, condition := range conditions {
					email := condition.UserInfo[0].Email
					url := condition.TrackingInfo[0].ShopeeUrl
					log.Println(utils.SendEmail(email, templates.CreateEmailNotifyPriceTemplate(templates.InfoEmailNotifyPrice{
						Email:         email,
						Price:         latestPrice.Price,
						PricePrevious: previousPrice.Price,
						Title:         "Notify price",
						LinkProduct:   url,
					})))
				}
			}
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			// for greater than
			ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			conditions, err := conditionService.FindAllByFilter(ctx, bson.M{
				"tracking":  bson.D{{Key: "$ref", Value: database.TrackingCollectionName}, {Key: "$id", Value: tracking.ID}},
				"condition": database.GREATER_THAN,
				"active":    true,
			})
			if err != nil {
				wg.Done()
				return
			}
			latestPrice := prices[0]
			previousPrice := prices[1]
			// check condition for greater than
			if latestPrice.Price > previousPrice.Price {
				// send email to user if price greater than condition
				for _, condition := range conditions {
					email := condition.UserInfo[0].Email
					url := condition.TrackingInfo[0].ShopeeUrl
					log.Println(utils.SendEmail(email, templates.CreateEmailNotifyPriceTemplate(templates.InfoEmailNotifyPrice{
						Email:         email,
						Price:         latestPrice.Price,
						PricePrevious: previousPrice.Price,
						Title:         "Notify price",
						LinkProduct:   url,
					})))
				}
			}
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			// for equal
			ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			conditions, err := conditionService.FindAllByFilter(ctx, bson.M{
				"tracking":  bson.D{{Key: "$ref", Value: database.TrackingCollectionName}, {Key: "$id", Value: tracking.ID}},
				"condition": database.EQUAL,
				"active":    true,
			})
			if err != nil {
				wg.Done()
				return
			}
			latestPrice := prices[0]
			// check condition for equal
			for _, condition := range conditions {
				if latestPrice.Price <= condition.Price {
					// send email to user if price equal condition
					email := condition.UserInfo[0].Email
					url := condition.TrackingInfo[0].ShopeeUrl
					log.Println(utils.SendEmail(email, templates.CreateEmailNotifyPriceTemplate(templates.InfoEmailNotifyPrice{
						Email:         email,
						Price:         latestPrice.Price,
						PricePrevious: latestPrice.Price,
						Title:         "Notify price",
						LinkProduct:   url,
					})))
				}
			}
			wg.Done()
		}()
		// wait for all goroutine done
		wg.Wait()
		// END
	}
}

func RunCronJobs() {
	go func() {
		for {
			crawlShop()
			notifyPriceChangeJob()
			<-time.After(10 * time.Minute)
		}
	}()
}
