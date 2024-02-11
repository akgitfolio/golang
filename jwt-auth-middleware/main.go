package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int64  `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
}

type File struct {
	ID       int64     `db:"id"`
	Name     string    `db:"name"`
	UserID   int64     `db:"user_id"`
	Content  []byte    `db:"content"`
	Uploaded time.Time `db:"uploaded"`
}

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.StandardClaims
}

func main() {
	db, err := sql.Open("postgres", "postgres://user:password@localhost/filesharing?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	secretKey := "your-secret-key"

	r := mux.NewRouter()
	r.Use(authMiddleware(secretKey))

	r.HandleFunc("/register", registerUser(db)).Methods("POST")
	r.HandleFunc("/login", loginUser(db, secretKey)).Methods("POST")

	r.HandleFunc("/files", uploadFile(db)).Methods("POST")
	r.HandleFunc("/files/{id}", downloadFile(db)).Methods("GET")
	r.HandleFunc("/files/{id}", deleteFile(db)).Methods("DELETE")

	log.Println("Server listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func authMiddleware(secretKey string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("Authorization")
			if tokenString == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(secretKey), nil
			})
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			if claims, ok := token.Claims.(*Claims); ok && token.Valid {
				ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			}
		})
	}
}

func registerUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}
		user.Password = string(hashedPassword)
		if _, err := db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, user.Password); err != nil {
			http.Error(w, "Error registering user", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func loginUser(db *sql.DB, secretKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds User
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		var storedUser User
		if err := db.QueryRow("SELECT id, password FROM users WHERE username = $1", creds.Username).Scan(&storedUser.ID, &storedUser.Password); err != nil {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(creds.Password)); err != nil {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}
		expirationTime := time.Now().Add(24 * time.Hour)
		claims := &Claims{
			UserID: storedUser.ID,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secretKey))
		if err != nil {
			http.Error(w, "Error generating token", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	}
}

func uploadFile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int64)
		var file File
		file.UserID = userID
		file.Uploaded = time.Now()
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "File too large", http.StatusBadRequest)
			return
		}
		fileHeader, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Invalid file", http.StatusBadRequest)
			return
		}
		defer fileHeader.Close()
		file.Content, err = io.ReadAll(fileHeader)
		if err != nil {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}
		if _, err := db.Exec("INSERT INTO files (name, user_id, content, uploaded) VALUES ($1, $2, $3, $4)", file.Name, file.UserID, file.Content, file.Uploaded); err != nil {
			http.Error(w, "Error uploading file", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func downloadFile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int64)
		fileID := mux.Vars(r)["id"]
		var file File
		if err := db.QueryRow("SELECT id, name, user_id, content, uploaded FROM files WHERE id = $1 AND user_id = $2", fileID, userID).Scan(&file.ID, &file.Name, &file.UserID, &file.Content, &file.Uploaded); err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Name))
		w.Write(file.Content)
	}
}

func deleteFile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int64)
		fileID := mux.Vars(r)["id"]
		if _, err := db.Exec("DELETE FROM files WHERE id = $1 AND user_id = $2", fileID, userID); err != nil {
			http.Error(w, "Error deleting file", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
