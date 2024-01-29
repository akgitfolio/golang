package main

import (
	"crypto/subtle"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var (
	dir      string
	username string
	password string
)

func main() {
	flag.StringVar(&dir, "dir", ".", "Directory to serve files from")
	flag.StringVar(&username, "username", "admin", "Username for basic authentication")
	flag.StringVar(&password, "password", "password", "Password for basic authentication")
	flag.Parse()

	http.HandleFunc("/", basicAuth(fileHandler))
	http.HandleFunc("/upload", basicAuth(uploadHandler))

	log.Printf("Serving files from directory: %s\n", dir)
	log.Printf("Starting server on :8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Join(dir, r.URL.Path)
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if info.IsDir() {
		serveDirectory(w, r, filePath)
	} else {
		http.ServeFile(w, r, filePath)
	}
}

func serveDirectory(w http.ResponseWriter, r *http.Request, dirPath string) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<h1>Directory listing for %s</h1>", r.URL.Path)
	fmt.Fprintf(w, "<ul>")
	for _, file := range files {
		name := file.Name()
		if file.IsDir() {
			name += "/"
		}
		fmt.Fprintf(w, `<li><a href="%s">%s</a></li>`, filepath.Join(r.URL.Path, name), name)
	}
	fmt.Fprintf(w, "</ul>")
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<form method="POST" enctype="multipart/form-data">
                        <input type="file" name="file">
                        <input type="submit" value="Upload">
                        </form>`)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	destPath := filepath.Join(dir, header.Filename)
	destFile, err := os.Create(destPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "File uploaded successfully: %s\n", header.Filename)
}
