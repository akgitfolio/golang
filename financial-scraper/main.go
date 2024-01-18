package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gocolly/colly"
	"github.com/tealeg/xlsx"
	"github.com/wcharczuk/go-chart/v2"
)

type FinancialData struct {
	Ticker string
	Price  float64
	Time   time.Time
}

var stockData []FinancialData
var cryptoData []FinancialData

func main() {
	// Scrape data
	scrapeStocks()
	scrapeCrypto()

	// Analyze data
	analysis := analyzeData(stockData, cryptoData)

	// Visualize data
	visualizeData(stockData, cryptoData)

	// Export data
	exportCSV(stockData, "stock_data.csv")
	exportCSV(cryptoData, "crypto_data.csv")
	exportExcel(stockData, "stock_data.xlsx")
	exportExcel(cryptoData, "crypto_data.xlsx")

	// Print analysis
	printAnalysis(analysis)
}

func scrapeStocks() {
	c := colly.NewCollector()

	c.OnHTML(".D(ib) .Fz(36px)", func(e *colly.HTMLElement) {
		priceStr := e.Text
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			log.Println("Failed to parse stock price:", err)
			return
		}
		data := FinancialData{
			Ticker: "AAPL",
			Price:  price,
			Time:   time.Now(),
		}
		stockData = append(stockData, data)
	})

	err := c.Visit("https://finance.yahoo.com/quote/AAPL?p=AAPL")
	if err != nil {
		log.Println("Failed to scrape stock data:", err)
	}
}

func scrapeCrypto() {
	c := colly.NewCollector()

	c.OnHTML(".priceValue___11gHJ", func(e *colly.HTMLElement) {
		priceStr := e.Text[1:] // Remove the dollar sign
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			log.Println("Failed to parse crypto price:", err)
			return
		}
		data := FinancialData{
			Ticker: "BTC-USD",
			Price:  price,
			Time:   time.Now(),
		}
		cryptoData = append(cryptoData, data)
	})

	err := c.Visit("https://coinmarketcap.com/currencies/bitcoin/")
	if err != nil {
		log.Println("Failed to scrape crypto data:", err)
	}
}

func analyzeData(stockData, cryptoData []FinancialData) map[string]float64 {
	analysis := make(map[string]float64)
	if len(stockData) > 0 {
		analysis["Stock Max Price"] = maxPrice(stockData)
		analysis["Stock Min Price"] = minPrice(stockData)
		analysis["Stock Average Price"] = avgPrice(stockData)
	}
	if len(cryptoData) > 0 {
		analysis["Crypto Max Price"] = maxPrice(cryptoData)
		analysis["Crypto Min Price"] = minPrice(cryptoData)
		analysis["Crypto Average Price"] = avgPrice(cryptoData)
	}
	return analysis
}

func maxPrice(data []FinancialData) float64 {
	max := data[0].Price
	for _, d := range data {
		if d.Price > max {
			max = d.Price
		}
	}
	return max
}

func minPrice(data []FinancialData) float64 {
	min := data[0].Price
	for _, d := range data {
		if d.Price < min {
			min = d.Price
		}
	}
	return min
}

func avgPrice(data []FinancialData) float64 {
	total := 0.0
	for _, d := range data {
		total += d.Price
	}
	return total / float64(len(data))
}

func visualizeData(stockData, cryptoData []FinancialData) {
	stockPrices := make([]float64, len(stockData))
	cryptoPrices := make([]float64, len(cryptoData))
	for i, data := range stockData {
		stockPrices[i] = data.Price
	}
	for i, data := range cryptoData {
		cryptoPrices[i] = data.Price
	}

	stockGraph := chart.Chart{
		Series: []chart.Series{
			chart.TimeSeries{
				Name: "Stock Prices",
				Style: chart.Style{
					Show:        true,
					StrokeColor: chart.ColorBlue,
				},
				XValues: timeSeries(stockData),
				YValues: stockPrices,
			},
		},
	}

	cryptoGraph := chart.Chart{
		Series: []chart.Series{
			chart.TimeSeries{
				Name: "Crypto Prices",
				Style: chart.Style{
					Show:        true,
					StrokeColor: chart.ColorGreen,
				},
				XValues: timeSeries(cryptoData),
				YValues: cryptoPrices,
			},
		},
	}

	f, _ := os.Create("stock_prices.png")
	defer f.Close()
	stockGraph.Render(chart.PNG, f)

	f, _ = os.Create("crypto_prices.png")
	defer f.Close()
	cryptoGraph.Render(chart.PNG, f)
}

func timeSeries(data []FinancialData) []time.Time {
	times := make([]time.Time, len(data))
	for i, d := range data {
		times[i] = d.Time
	}
	return times
}

func exportCSV(data []FinancialData, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal("Cannot create CSV file:", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Ticker", "Price", "Time"})
	for _, d := range data {
		writer.Write([]string{d.Ticker, fmt.Sprintf("%f", d.Price), d.Time.Format(time.RFC3339)})
	}
}

func exportExcel(data []FinancialData, filename string) {
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Sheet1")
	if err != nil {
		log.Fatal("Cannot create Excel sheet:", err)
	}

	row := sheet.AddRow()
	row.AddCell().Value = "Ticker"
	row.AddCell().Value = "Price"
	row.AddCell().Value = "Time"
	for _, d := range data {
		row := sheet.AddRow()
		row.AddCell().Value = d.Ticker
		row.AddCell().Value = fmt.Sprintf("%f", d.Price)
		row.AddCell().Value = d.Time.Format(time.RFC3339)
	}

	err = file.Save(filename)
	if err != nil {
		log.Fatal("Cannot save Excel file:", err)
	}
}

func printAnalysis(analysis map[string]float64) {
	analysisJSON, _ := json.MarshalIndent(analysis, "", "  ")
	fmt.Println("Analysis:")
	fmt.Println(string(analysisJSON))
}
