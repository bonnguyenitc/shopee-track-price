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
	"github.com/bonnguyenitc/shopee-stracks/back-end-go/utils"
	"github.com/chromedp/chromedp"
)

func GetProductsByShopID(shopID string) ([]database.Product, error) {
	var url = fmt.Sprintf("https://shopee.vn/api/v4/recommend/recommend?bundle=shop_page_product_tab_main&limit=999&offset=0&section=shop_page_product_tab_main_sec&shopid=%s", shopID)
	// create a new Chrome browser context
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// navigate to a page, retrieve the page source
	var html string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.OuterHTML("html", &html),
	)
	if err != nil {
		return nil, err
	}

	// parse the page HTML with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	products := []database.Product{}

	// find and print all links
	doc.Find("pre").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		var result map[string]interface{}

		err := json.Unmarshal([]byte(text), &result)

		if err != nil {
			return
		}

		error := result["error"].(float64)

		if error != 0 {
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
				IDShopee:               itemReal["itemid"].(float64),
				ShopName:               itemReal["shop_name"].(string),
				ShopRating:             itemReal["shop_rating"].(float64),
				Name:                   itemReal["name"].(string),
				Stock:                  itemReal["stock"].(float64),
				Sold:                   itemReal["sold"].(float64),
				HistoricalSold:         itemReal["historical_sold"].(float64),
				LikedCount:             itemReal["liked_count"].(float64),
				CmtCount:               itemReal["cmt_count"].(float64),
				Price:                  itemReal["price"].(float64),
				PriceMin:               itemReal["price_min"].(float64),
				PriceMax:               itemReal["price_max"].(float64),
				PriceMinBeforeDiscount: itemReal["price_min_before_discount"].(float64),
				PriceMaxBeforeDiscount: itemReal["price_max_before_discount"].(float64),
				PriceBeforeDiscount:    itemReal["price_before_discount"].(float64),
				RawDiscount:            itemReal["raw_discount"].(float64),
				Images:                 utils.CreateListImageFromIds(images),
			})
		}
	})

	return products, nil
}