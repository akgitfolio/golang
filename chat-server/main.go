package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type PlaylistItem struct {
	SongName string `json:"song_name"`
	Votes    int    `json:"votes"`
}

type Message struct {
	Type    string       `json:"type"`
	Payload PlaylistItem `json:"payload"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins
}

var (
	playlist []PlaylistItem
	clients  = make(map[*websocket.Conn]bool)
	mu       sync.Mutex // Mutex for playlist and clients map
)

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	mu.Lock()
	clients[conn] = true
	mu.Unlock()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		var msg Message
		err = json.Unmarshal(message, &msg)
		if err != nil {
			log.Println("Error unmarshalling message:", err)
			continue
		}

		switch msg.Type {
		case "add":
			addSong(msg.Payload)
		case "vote":
			voteSong(msg.Payload)
		}

		broadcastPlaylist()
	}

	mu.Lock()
	delete(clients, conn)
	mu.Unlock()
}

func addSong(song PlaylistItem) {
	mu.Lock()
	defer mu.Unlock()
	playlist = append(playlist, song)
}

func voteSong(song PlaylistItem) {
	mu.Lock()
	defer mu.Unlock()
	for i, item := range playlist {
		if item.SongName == song.SongName {
			playlist[i].Votes += song.Votes
			break
		}
	}
}

func broadcastPlaylist() {
	mu.Lock()
	defer mu.Unlock()
	playlistJSON, err := json.Marshal(playlist)
	if err != nil {
		log.Println("Error marshalling playlist:", err)
		return
	}
	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, playlistJSON)
		if err != nil {
			log.Println("Error broadcasting to client:", err)
			client.Close()
			delete(clients, client)
		}
	}
}
