package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/oschwald/geoip2-golang"
)

type URLMapping struct {
	OriginalURL string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	ClickCount  int
	Geolocation map[string]int
}

var (
	db         *bolt.DB
	geoDB      *geoip2.Reader
	baseURL    = "http://localhost:8080/"
	bucketName = []byte("urlMappings")
)

func main() {
	var err error
	db, err = bolt.Open("urlshortener.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	if err != nil {
		log.Fatal(err)
	}

	geoDB, err = geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer geoDB.Close()

	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/shorten", shortenHandler).Methods("POST")
	r.HandleFunc("/{alias}", redirectHandler).Methods("GET")

	http.Handle("/", r)
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, nil)
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	originalURL := r.FormValue("url")
	customAlias := r.FormValue("alias")
	expiration := r.FormValue("expiration")

	if originalURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	alias := customAlias
	if alias == "" {
		alias = generateAlias()
	}

	var expiresAt time.Time
	if expiration != "" {
		duration, err := time.ParseDuration(expiration)
		if err != nil {
			http.Error(w, "Invalid expiration format", http.StatusBadRequest)
			return
		}
		expiresAt = time.Now().Add(duration)
	}

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b.Get([]byte(alias)) != nil {
			return fmt.Errorf("alias already exists")
		}

		mapping := URLMapping{
			OriginalURL: originalURL,
			CreatedAt:   time.Now(),
			ExpiresAt:   expiresAt,
			ClickCount:  0,
			Geolocation: make(map[string]int),
		}

		data, err := json.Marshal(mapping)
		if err != nil {
			return err
		}

		return b.Put([]byte(alias), data)
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shortenedURL := baseURL + alias
	fmt.Fprintf(w, "Shortened URL: %s\n", shortenedURL)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alias := vars["alias"]

	var mapping URLMapping

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		data := b.Get([]byte(alias))
		if data == nil {
			return fmt.Errorf("alias not found")
		}

		return json.Unmarshal(data, &mapping)
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if !mapping.ExpiresAt.IsZero() && time.Now().After(mapping.ExpiresAt) {
		http.Error(w, "URL has expired", http.StatusGone)
		return
	}

	updateClickCount(alias)
	updateGeolocation(alias, r.RemoteAddr)

	http.Redirect(w, r, mapping.OriginalURL, http.StatusFound)
}

func generateAlias() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func updateClickCount(alias string) {
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		data := b.Get([]byte(alias))
		if data == nil {
			return fmt.Errorf("alias not found")
		}

		var mapping URLMapping
		json.Unmarshal(data, &mapping)
		mapping.ClickCount++

		newData, err := json.Marshal(mapping)
		if err != nil {
			return err
		}

		return b.Put([]byte(alias), newData)
	})
}

func updateGeolocation(alias string, remoteAddr string) {
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return
	}

	record, err := geoDB.City(net.ParseIP(ip))
	if err != nil {
		return
	}

	country := record.Country.IsoCode

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		data := b.Get([]byte(alias))
		if data == nil {
			return fmt.Errorf("alias not found")
		}

		var mapping URLMapping
		json.Unmarshal(data, &mapping)
		mapping.Geolocation[country]++

		newData, err := json.Marshal(mapping)
		if err != nil {
			return err
		}

		return b.Put([]byte(alias), newData)
	})
}
