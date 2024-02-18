package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/uber-go/ratelimit"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int64  `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
}

type Auction struct {
	ID          int64     `db:"id"`
	SellerID    int64     `db:"seller_id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	StartingBid float64   `db:"starting_bid"`
	CurrentBid  float64   `db:"current_bid"`
	EndTime     time.Time `db:"end_time"`
}

type Bid struct {
	ID        int64     `db:"id"`
	AuctionID int64     `db:"auction_id"`
	UserID    int64     `db:"user_id"`
	Amount    float64   `db:"amount"`
	Timestamp time.Time `db:"timestamp"`
}

func main() {
	db, err := sql.Open("postgres", "user=auction_user password=auction_password dbname=auction_db sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	globalLimiter := ratelimit.New(100)
	userLimiter := ratelimit.New(10)
	bidLimiter := ratelimit.New(5)

	router := gin.Default()
	router.Use(ErrorHandlerMiddleware())
	router.Use(RateLimitMiddleware(globalLimiter))

	router.POST("/users", func(c *gin.Context) {
		user := User{}
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !userLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		user.Password = string(hashedPassword)

		_, err = db.ExecContext(c, "INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, user.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
	})

	router.POST("/login", func(c *gin.Context) {
	})

	router.POST("/auctions", func(c *gin.Context) {
	})

	router.POST("/auctions/:auctionID/bids", func(c *gin.Context) {
		bid := Bid{}
		if err := c.BindJSON(&bid); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if !bidLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}

		_, err = db.ExecContext(c, "INSERT INTO bids (auction_id, user_id, amount) VALUES ($1, $2, $3)", bid.AuctionID, bid.UserID, bid.Amount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		_, err = db.ExecContext(c, "UPDATE auctions SET current_bid = $1 WHERE id = $2", bid.Amount, bid.AuctionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Bid placed successfully"})
	})

	router.Run(":8080")
}

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}

func RateLimitMiddleware(rl ratelimit.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.Allow() {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		c.Next()
	}
}
