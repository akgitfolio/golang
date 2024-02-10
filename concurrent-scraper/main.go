package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
)

type Product struct {
	Name    string
	Price   float64
	URL     string
	Website string
}

type Scraper struct {
	URL           string
	Category      string
	Keyword       string
	PriceSelector string
}

func main() {
	targetWebsites := []Scraper{
		{URL: "https://www.amazon.com", Category: "Electronics", Keyword: "Laptop", PriceSelector: ".a-price-whole"},
		{URL: "https://www.ebay.com", Category: "Electronics", Keyword: "Laptop", PriceSelector: ".s-item__price"},
	}

	var wg sync.WaitGroup
	var errGroup errgroup.Group
	productChan := make(chan Product)

	for _, website := range targetWebsites {
		wg.Add(1)
		errGroup.Go(func(website Scraper) func() error {
			return func() error {
				defer wg.Done()
				products, err := scrapeWebsite(website)
				if err != nil {
					return err
				}
				for _, product := range products {
					productChan <- product
				}
				return nil
			}
		}(website))
	}

	go func() {
		defer close(productChan)
		wg.Wait()
		if err := errGroup.Wait(); err != nil {
			fmt.Println("Error occurred during scraping:", err)
			return
		}
		products := make(map[string][]Product)
		for product := range productChan {
			products[product.Name] = append(products[product.Name], product)
		}
		presentData(products)
	}()

	wg.Wait()
}

func scrapeWebsite(website Scraper) ([]Product, error) {
	products := []Product{}
	doc, err := getDocument(website.URL, website.Category, website.Keyword)
	if err != nil {
		return nil, err
	}

	doc.Find(website.PriceSelector).Each(func(i int, selection *goquery.Selection) {
		priceString := selection.Text()
		priceFloat, err := strconv.ParseFloat(strings.TrimSpace(priceString), 64)
		if err != nil {
			return
		}
		productURL, _ := selection.Parent().Find("a").Attr("href")
		productName := strings.TrimSpace(selection.Parent().Find("a").Text())
		products = append(products, Product{Name: productName, Price: priceFloat, URL: productURL, Website: website.URL})
	})

	return products, nil
}

func getDocument(url, category, keyword string) (*goquery.Document, error) {
	client := http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/s?k=%s&c=%s", url, keyword, category), nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.102 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return goquery.NewDocumentFromReader(resp.Body)
}

func presentData(products map[string][]Product) {
	fmt.Println("========== Price Comparison ==========")
	for name, productList := range products {
		fmt.Printf("Product Name: %s\n", name)
		for _, product := range productList {
			fmt.Printf("- Website: %s, Price: $%.2f, URL: %s\n", product.Website, product.Price, product.URL)
		}
		fmt.Println()
	}
}
