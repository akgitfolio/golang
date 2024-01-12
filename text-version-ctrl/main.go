package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gorilla/websocket"
)

// Embedding static files
//
//go:embed static/*
var content embed.FS

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Accepting all requests
	},
}

type Server struct {
	clients       map[*websocket.Conn]bool
	handleMessage func(message []byte)
}

func (server *Server) echo(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	server.clients[conn] = true
	defer func() {
		delete(server.clients, conn)
		conn.Close()
	}()
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil || mt == websocket.CloseMessage {
			break
		}
		go server.handleMessage(message)
	}
}

func (server *Server) WriteMessage(message []byte) {
	for conn := range server.clients {
		conn.WriteMessage(websocket.TextMessage, message)
	}
}

func StartServer(handleMessage func(message []byte)) *Server {
	server := &Server{
		clients:       make(map[*websocket.Conn]bool),
		handleMessage: handleMessage,
	}
	http.HandleFunc("/ws", server.echo)
	go http.ListenAndServe(":8080", nil)
	return server
}

// Initialize a new repository
func initRepo(path string) (*git.Repository, error) {
	r, err := git.PlainInit(path, false)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Clone an existing repository
func cloneRepo(url, path string) (*git.Repository, error) {
	r, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Create a new branch
func createBranch(r *git.Repository, branchName string) error {
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	ref := plumbing.NewBranchReferenceName(branchName)
	err = w.Checkout(&git.CheckoutOptions{
		Branch: ref,
		Create: true,
	})
	if err != nil {
		return err
	}
	return nil
}

// Merge branches
func mergeBranches(r *git.Repository, sourceBranch, targetBranch string) error {
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(targetBranch),
	})
	if err != nil {
		return err
	}
	err = w.Merge(&git.MergeOptions{
		Commit:  true,
		Message: "Merging " + sourceBranch + " into " + targetBranch,
	})
	if err != nil {
		return err
	}
	return nil
}

// Commit changes to the repository
func commitChanges(r *git.Repository, message string) error {
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	_, err = w.Add(".")
	if err != nil {
		return err
	}
	_, err = w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Author Name",
			Email: "author@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

// Get the commit history
func getHistory(r *git.Repository) ([]*object.Commit, error) {
	ref, err := r.Head()
	if err != nil {
		return nil, err
	}
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, err
	}
	var commits []*object.Commit
	err = cIter.ForEach(func(c *object.Commit) error {
		commits = append(commits, c)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return commits, nil
}

func main() {
	// Initialize the repository
	repo, err := initRepo("/path/to/repo")
	if err != nil {
		log.Fatal(err)
	}

	// Start WebSocket server
	server := StartServer(func(message []byte) {
		fmt.Println("Received message:", string(message))
	})

	// Example usage of repository functions
	err = createBranch(repo, "new-branch")
	if err != nil {
		log.Fatal(err)
	}

	err = commitChanges(repo, "Initial commit")
	if err != nil {
		log.Fatal(err)
	}

	commits, err := getHistory(repo)
	if err != nil {
		log.Fatal(err)
	}

	for _, commit := range commits {
		fmt.Println(commit)
	}

	// Serve embedded static files
	http.Handle("/", http.FileServer(http.FS(content)))
	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
