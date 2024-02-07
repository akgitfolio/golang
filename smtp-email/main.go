package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"

	"github.com/jordan-wright/email"
)

type AuctionItem struct {
	ID         int
	Name       string
	CurrentBid float64
	BidHistory []Bid
}

type Bid struct {
	ItemID int
	Bidder string
	Email  string
	Amount float64
}

func sendEmail(recipient, subject, body string) error {
	e := email.NewEmail()
	e.From = "your_email@example.com"
	e.To = []string{recipient}
	e.Subject = subject
	e.HTML = []byte(body)

	smtpAddr := "smtp.example.com:587"
	smtpAuth := smtp.PlainAuth("", "your_email@example.com", "your_password", "smtp.example.com")

	err := e.Send(smtpAddr, smtpAuth)
	if err != nil {
		return err
	}

	return nil
}

func handleBid(bid Bid, items map[int]*AuctionItem) {
	item, ok := items[bid.ItemID]
	if !ok {
		log.Printf("Item not found: %d", bid.ItemID)
		return
	}

	item.CurrentBid = bid.Amount
	item.BidHistory = append(item.BidHistory, bid)

	if len(item.BidHistory) > 1 {
		previousBid := item.BidHistory[len(item.BidHistory)-2]
		subject := fmt.Sprintf("You've been outbid on %s", item.Name)
		body := fmt.Sprintf("Your bid of %.2f has been outbid. The current highest bid is %.2f.", previousBid.Amount, bid.Amount)
		err := sendEmail(previousBid.Email, subject, body)
		if err != nil {
			log.Printf("Error sending email to %s: %v", previousBid.Email, err)
		}
	}
}

func main() {
	items := map[int]*AuctionItem{
		1: {ID: 1, Name: "Antique Clock", CurrentBid: 100.00},
		2: {ID: 2, Name: "Rare Painting", CurrentBid: 500.00},
	}

	http.HandleFunc("/bid", func(w http.ResponseWriter, r *http.Request) {
		var bid Bid
		err := json.NewDecoder(r.Body).Decode(&bid)
		if err != nil {
			http.Error(w, "Invalid bid data", http.StatusBadRequest)
			return
		}

		go handleBid(bid, items)
		w.WriteHeader(http.StatusAccepted)
	})

	log.Println("Server listening on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
