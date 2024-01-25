package crawl

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/database"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/logs"
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/utils"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	fakeUseragent "github.com/eddycjy/fake-useragent"
	"github.com/sirupsen/logrus"
)

func GetProductsByShopID(shopID string) ([]database.Product, error) {
	var url = fmt.Sprintf("https://shopee.vn/api/v4/recommend/recommend?bundle=shop_page_product_tab_main&limit=999&offset=0&section=shop_page_product_tab_main_sec&shopid=%s", shopID)

	random := fakeUseragent.Random()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.UserAgent(random),
	)
	ctx, cancel = chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	allocCtx, cancel := chromedp.NewRemoteAllocator(ctx, "http://headless-shell:9222")
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// navigate to a page, retrieve the page source
	var html string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.OuterHTML("html", &html),
	)
	if err != nil {
		logs.LogWarning(logrus.Fields{
			"shopID": shopID,
			"data":   err.Error(),
		}, "GetProductsByShopID run chromedp")
		return nil, err
	}

	// parse the page HTML with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		logs.LogWarning(logrus.Fields{
			"shopID": shopID,
			"data":   err.Error(),
		}, "GetProductsByShopID parse html")
		return nil, err
	}

	products := []database.Product{}

	doc.Find("pre").Each(func(i int, s *goquery.Selection) {
		text := s.Text()

		var result map[string]interface{}

		err := json.Unmarshal([]byte(text), &result)

		if err != nil {
			logs.LogWarning(logrus.Fields{
				"shopID": shopID,
				"data":   err.Error(),
			}, "GetProductsByShopID unmarshal json")
			return
		}

		error := result["error"].(float64)

		if error != 0 {
			logs.LogWarning(logrus.Fields{
				"shopID": shopID,
				"data":   error,
			}, "GetProductsByShopID is blocked")
			return
		}

		data := result["data"].(map[string]interface{})

		sections := data["sections"].([]interface{})

		firstItem := sections[0].(map[string]interface{})

		firstItemData := firstItem["data"].(map[string]interface{})

		items := firstItemData["item"].([]interface{})

		for _, item := range items {
			itemReal := item.(map[string]interface{})
			images := itemReal["images"].([]interface{})
			products = append(products, database.Product{
				IDShopee:               utils.ConvertFloat64ToInt64(itemReal["itemid"].(float64)),
				ShopName:               itemReal["shop_name"].(string),
				ShopRating:             itemReal["shop_rating"].(float64),
				Name:                   itemReal["name"].(string),
				Stock:                  utils.ConvertFloat64ToInt32(itemReal["stock"].(float64)),
				Sold:                   utils.ConvertFloat64ToInt32(itemReal["sold"].(float64)),
				HistoricalSold:         utils.ConvertFloat64ToInt32(itemReal["historical_sold"].(float64)),
				LikedCount:             utils.ConvertFloat64ToInt32(itemReal["liked_count"].(float64)),
				CmtCount:               utils.ConvertFloat64ToInt32(itemReal["cmt_count"].(float64)),
				Price:                  utils.ConvertFloat64ToInt64(itemReal["price"].(float64)),
				PriceMin:               utils.ConvertFloat64ToInt64(itemReal["price_min"].(float64)),
				PriceMax:               utils.ConvertFloat64ToInt64(itemReal["price_max"].(float64)),
				PriceMinBeforeDiscount: utils.ConvertFloat64ToInt64(itemReal["price_min_before_discount"].(float64)),
				PriceMaxBeforeDiscount: utils.ConvertFloat64ToInt64(itemReal["price_max_before_discount"].(float64)),
				PriceBeforeDiscount:    utils.ConvertFloat64ToInt64(itemReal["price_before_discount"].(float64)),
				RawDiscount:            float32(itemReal["raw_discount"].(float64)),
				Images:                 utils.CreateListImageFromIds(images),
			})
		}
	})

	return products, nil
}

// ScrapeProductDetail scrape product detail
func ScrapeProductDetail() (database.Product, error) {
	var url = ""

	productId := utils.GetProductIDFromUrl(url)

	random := fakeUseragent.Random()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoDefaultBrowserCheck,
		// Remove this if you have not proxy server
		// chromedp.ProxyServer("210.211.113.35:80"),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("headless", false),
		chromedp.Flag("start-fullscreen", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
		chromedp.UserAgent(random),
		chromedp.NoSandbox,
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("shm-size", "4GB"),
	)
	ctx, cancel = chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	allocCtx, cancel := chromedp.NewRemoteAllocator(ctx, "http://headless-shell:9222")
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	defer chromedp.Cancel(ctx)

	// navigate to a page, retrieve the page source
	// var nodes []*cdp.Node
	task := chromedp.Tasks{
		chromedp.Navigate("https://shopee.vn/mall"),
		chromedp.Navigate(url),
		chromedp.WaitReady("#main", chromedp.ByID),
		// chromedp.Nodes(".G27FPf", &nodes, chromedp.AtLeast(1), chromedp.ByQueryAll),
		chromedp.ActionFunc(func(ctx context.Context) error {
			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			res, err := dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
			if err != nil {
				return err
			}
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(res))
			if err != nil {
				return err
			}

			productView := doc.Find(".product-briefing").First()

			// get name of product
			nameProductNode := productView.Find("section span").First()
			name := nameProductNode.Text()
			log.Println("name of product:", name)

			// get like of product
			likeProductNode := productView.Find(".rhG6k7").Last()
			like := likeProductNode.Text()
			log.Println("like of product:", like)

			// get rating of product
			ratingProductNode := productView.Find(".F9RHbS").First()
			rating := ratingProductNode.Text()
			log.Println("rating of product:", rating)

			// get price of product
			priceProductNode := productView.Find(".G27FPf").First()
			price := priceProductNode.Text()
			log.Println("price of product:", price)

			// get old price of product
			oldPriceProductNode := productView.Find(".qg2n76").First()
			oldPrice := oldPriceProductNode.Text()
			log.Println("old price of product:", oldPrice)

			// get discount of product
			discountProductNode := productView.Find(".o_z7q9").First()
			discount := discountProductNode.Text()
			log.Println("discount of product:", discount)

			// get sold of product
			soldProductNode := productView.Find(".AcmPRb").First()
			sold := soldProductNode.Text()
			log.Println("sold of product:", sold)

			// get stock of product
			stockProductNode := productView.Find(".OaFP0p .flex div").Last()
			stock := stockProductNode.Text()
			log.Println("stock of product:", stock)

			// get review of product
			reviewProductNode := productView.Find(".F9RHbS").First()
			review := reviewProductNode.Text()
			log.Println("review of product:", review)

			// shop info
			shopInfo := doc.Find(".page-product__shop").First()

			// get shop name
			shopNameNode := shopInfo.Find(".fV3TIn").First()
			shopName := shopNameNode.Text()
			log.Println("shop name:", shopName)

			// get shop image
			shopImageNode := shopInfo.Find(".Qm507c").First()
			shopImage, _ := shopImageNode.Attr("src")
			log.Println("shop image:", shopImage)

			// get flash sale
			flashSaleNode := doc.Find(".shopee-countdown-timer").First()
			flashSale, _ := flashSaleNode.Attr("aria-label")
			log.Println("flash sale:", flashSale)

			return nil
		}),
	}

	if err := chromedp.Run(ctx,
		network.ClearBrowserCookies(),
	); err != nil {
		log.Fatal(err)
	}

	err := chromedp.Run(ctx, task)
	if err != nil {
		logs.LogWarning(logrus.Fields{
			"shopID": productId,
			"data":   err.Error(),
		}, "ScrapeProductDetail run chromedp")
		return database.Product{}, err
	}

	product := database.Product{}

	log.Println("End scrape product detail")

	return product, nil
}
