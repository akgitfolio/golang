package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/boltdb/bolt"
	"github.com/zserge/lorca"
)

const (
	dbFile           = "clipboard.db"
	historyBucket    = "History"
	favoritesBucket  = "Favorites"
	categoriesBucket = "Categories"
)

type ClipItem struct {
	Text      string
	Timestamp time.Time
}

func main() {
	ui, err := lorca.New("", "", 800, 600)
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(historyBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = tx.CreateBucketIfNotExists([]byte(favoritesBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = tx.CreateBucketIfNotExists([]byte(categoriesBucket))
		return err
	})
	if err != nil {
		log.Fatal(err)
	}

	go monitorClipboard(db)

	ui.Bind("getHistory", func() []ClipItem {
		return getClipItems(db, historyBucket)
	})
	ui.Bind("getFavorites", func() []ClipItem {
		return getClipItems(db, favoritesBucket)
	})
	ui.Bind("addFavorite", func(text string) {
		addClipItem(db, favoritesBucket, text)
	})
	ui.Bind("search", func(query string) []ClipItem {
		return searchClipItems(db, query)
	})

	ui.Load("data:text/html," + url.PathEscape(indexHTML))

	<-ui.Done()
}

func monitorClipboard(db *bolt.DB) {
	var lastText string
	for {
		text, err := clipboard.ReadAll()
		if err != nil {
			log.Println("Failed to read clipboard:", err)
			time.Sleep(time.Second)
			continue
		}

		if text != lastText && text != "" {
			lastText = text
			addClipItem(db, historyBucket, text)
		}

		time.Sleep(time.Second)
	}
}

func addClipItem(db *bolt.DB, bucket string, text string) {
	item := ClipItem{
		Text:      text,
		Timestamp: time.Now(),
	}

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		buf := new(bytes.Buffer)
		enc := gob.NewEncoder(buf)
		err := enc.Encode(item)
		if err != nil {
			return err
		}
		return b.Put([]byte(item.Timestamp.Format(time.RFC3339Nano)), buf.Bytes())
	})
}

func getClipItems(db *bolt.DB, bucket string) []ClipItem {
	var items []ClipItem
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		b.ForEach(func(_, v []byte) error {
			var item ClipItem
			buf := bytes.NewBuffer(v)
			dec := gob.NewDecoder(buf)
			if err := dec.Decode(&item); err == nil {
				items = append(items, item)
			}
			return nil
		})
		return nil
	})
	return items
}

func searchClipItems(db *bolt.DB, query string) []ClipItem {
	var items []ClipItem
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(historyBucket))
		b.ForEach(func(_, v []byte) error {
			var item ClipItem
			buf := bytes.NewBuffer(v)
			dec := gob.NewDecoder(buf)
			if err := dec.Decode(&item); err == nil && strings.Contains(item.Text, query) {
				items = append(items, item)
			}
			return nil
		})
		return nil
	})
	return items
}

const indexHTML = `
<!DOCTYPE html>
<html>
<head>
	<title>Clipboard Manager</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 20px; }
		.clip-item { margin-bottom: 10px; padding: 10px; border: 1px solid #ccc; }
		.clip-item .timestamp { color: #999; font-size: 12px; }
	</style>
</head>
<body>
	<h1>Clipboard Manager</h1>
	<h2>History</h2>
	<div id="history"></div>
	<h2>Favorites</h2>
	<div id="favorites"></div>
	<h2>Search</h2>
	<input type="text" id="searchInput" oninput="search()">
	<div id="searchResults"></div>
	<script>
		async function loadHistory() {
			let history = await getHistory();
			let historyDiv = document.getElementById('history');
			historyDiv.innerHTML = '';
			history.forEach(item => {
				historyDiv.innerHTML += '<div class="clip-item">' + item.Text + '<div class="timestamp">' + item.Timestamp + '</div></div>';
			});
		}

		async function loadFavorites() {
			let favorites = await getFavorites();
			let favoritesDiv = document.getElementById('favorites');
			favoritesDiv.innerHTML = '';
			favorites.forEach(item => {
				favoritesDiv.innerHTML += '<div class="clip-item">' + item.Text + '<div class="timestamp">' + item.Timestamp + '</div></div>';
			});
		}

		async function search() {
			let query = document.getElementById('searchInput').value;
			let results = await search(query);
			let searchResultsDiv = document.getElementById('searchResults');
			searchResultsDiv.innerHTML = '';
			results.forEach(item => {
				searchResultsDiv.innerHTML += '<div class="clip-item">' + item.Text + '<div class="timestamp">' + item.Timestamp + '</div></div>';
			});
		}

		window.onload = function() {
			loadHistory();
			loadFavorites();
		}
	</script>
</body>
</html>
`
