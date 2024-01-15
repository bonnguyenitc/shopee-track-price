package jobs

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/bonnguyenitc/shopee-stracks/back-end-go/crawl"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/database"
	"go.mongodb.org/mongo-driver/bson"
)

func crawlShop() {
	log.Println("Crawl data shop start...")
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
				"product_id": prod.ID,
				"created_at": bson.M{"$gte": startOfDay, "$lt": endOfDay},
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
	trackings, err := trackingService.FindAll(ctx, 10, 1)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(trackings)

	// get all price
	// compare price
	// send email
}

func RunCronJobs() {
	go func() {
		for {
			// crawlShop()
			notifyPriceChangeJob()
			<-time.After(10 * time.Minute)
		}
	}()
}
