package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ============================================================
// MODEL
// ============================================================

// Item is a single inventory record, mapped to the "items" table by GORM.
// The ID is a UUID generated in createItem; binding tags validate request bodies.
type Item struct {
	ID    string  `json:"id" gorm:"primaryKey"`
	Name  string  `json:"name" gorm:"not null" binding:"required"`
	Stock int     `json:"stock" gorm:"not null" binding:"gte=0"`
	Price float64 `json:"price" gorm:"not null" binding:"gte=0"`
}

// ============================================================
// DATABASE
// ============================================================

// db is the shared GORM handle used by the handlers.
var db *gorm.DB

// initDatabase connects to PostgreSQL, migrates the schema and seeds initial data.
// The connection string (dsn) comes from DATABASE_DSN in .env.
func initDatabase() {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("DATABASE_DSN is not set. Copy .env.example to .env and fill it in.")
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database! ", err)
	}

	if err := db.AutoMigrate(&Item{}); err != nil {
		log.Fatal("Failed to migrate the database! ", err)
	}
	seedDatabase()
}

// seedDatabase seeds the database with 20 default items when the table is empty.
func seedDatabase() {
	var count int64
	db.Model(&Item{}).Count(&count)
	if count > 0 {
		log.Println("Database already contains data, skipping seeding.")
		return
	}

	items := []Item{
		{ID: uuid.New().String(), Name: "Laptop", Stock: 10, Price: 999.99},
		{ID: uuid.New().String(), Name: "Smartphone", Stock: 20, Price: 699.99},
		{ID: uuid.New().String(), Name: "Headphones", Stock: 15, Price: 199.99},
		{ID: uuid.New().String(), Name: "Keyboard", Stock: 25, Price: 89.99},
		{ID: uuid.New().String(), Name: "Mouse", Stock: 30, Price: 49.99},
		{ID: uuid.New().String(), Name: "Monitor", Stock: 12, Price: 299.99},
		{ID: uuid.New().String(), Name: "Webcam", Stock: 18, Price: 79.99},
		{ID: uuid.New().String(), Name: "Printer", Stock: 7, Price: 149.99},
		{ID: uuid.New().String(), Name: "Tablet", Stock: 5, Price: 399.99},
		{ID: uuid.New().String(), Name: "Smartwatch", Stock: 14, Price: 249.99},
		{ID: uuid.New().String(), Name: "External Hard Drive", Stock: 8, Price: 119.99},
		{ID: uuid.New().String(), Name: "USB Flash Drive", Stock: 50, Price: 19.99},
		{ID: uuid.New().String(), Name: "Router", Stock: 6, Price: 89.99},
		{ID: uuid.New().String(), Name: "Projector", Stock: 3, Price: 499.99},
		{ID: uuid.New().String(), Name: "Bluetooth Speaker", Stock: 22, Price: 129.99},
		{ID: uuid.New().String(), Name: "Gaming Console", Stock: 11, Price: 499.99},
		{ID: uuid.New().String(), Name: "Camera", Stock: 4, Price: 599.99},
		{ID: uuid.New().String(), Name: "Fitness Tracker", Stock: 16, Price: 99.99},
		{ID: uuid.New().String(), Name: "Drone", Stock: 2, Price: 899.99},
		{ID: uuid.New().String(), Name: "VR Headset", Stock: 9, Price: 399.99},
	}

	if err := db.Create(&items).Error; err != nil {
		log.Fatal("Failed to seed database: ", err)
	}
	log.Println("Database seeded with 20 sample items.")
}

// ============================================================
// HANDLERS
// One function per endpoint (rubric). All use GORM via db and return JSON.
// ============================================================

// findItem loads an item by ID. On failure it writes the error response
// (404 if not found, 500 for any other database error) and returns false.
func findItem(c *gin.Context, id string) (Item, bool) {
	var item Item
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch item"})
		}
		return item, false
	}
	return item, true
}

// maxLimit caps how many items one page can return.
const maxLimit = 100

// sortableFields are the columns clients may sort by.
var sortableFields = map[string]bool{"name": true, "stock": true, "price": true}

// getItems handles GET /inventory
// Query params: limit (1-100, default 10), offset, sort_by (name|stock|price),
// order (asc|desc, case-insensitive), name (partial match), min_stock.
// Invalid values return 400.
func getItems(c *gin.Context) {
	limitParam := c.DefaultQuery("limit", "10")
	offsetParam := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit <= 0 || limit > maxLimit {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter; use 1 to " + strconv.Itoa(maxLimit)})
		return
	}

	offset, err := strconv.Atoi(offsetParam)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset parameter"})
		return
	}

	query := db.Model(&Item{})

	if name := c.Query("name"); name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}
	if raw := c.Query("min_stock"); raw != "" {
		minStock, err := strconv.Atoi(raw)
		if err != nil || minStock < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid min_stock parameter"})
			return
		}
		query = query.Where("stock >= ?", minStock)
	}

	// sort_by is whitelisted: it is concatenated into SQL, so it must never be raw user input.
	sortBy := c.DefaultQuery("sort_by", "name")
	if !sortableFields[sortBy] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sort_by value. Use 'name', 'stock' or 'price'"})
		return
	}
	order := strings.ToUpper(c.DefaultQuery("order", "ASC"))
	if order != "ASC" && order != "DESC" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order value. Use 'asc' or 'desc'"})
		return
	}
	// "id" as a tiebreaker keeps page boundaries stable when sort values are equal.
	query = query.Order(sortBy + " " + order).Order("id")

	var items []Item
	if err := query.Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch items"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"metadata": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

// getItem handles GET /inventory/:id (200, or 404 if missing)
func getItem(c *gin.Context) {
	item, ok := findItem(c, c.Param("id"))
	if !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

// createItem handles POST /inventory (201, or 400 on an invalid body)
func createItem(c *gin.Context) {
	var newItem Item
	if err := c.ShouldBindJSON(&newItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newItem.ID = uuid.New().String()
	if err := db.Create(&newItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create item"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": newItem})
}

// updateItem handles PUT /inventory/:id (200, 404 if missing, 400 on an invalid body)
func updateItem(c *gin.Context) {
	item, ok := findItem(c, c.Param("id"))
	if !ok {
		return
	}

	// Bind into a separate struct so the stored row isn't overwritten and the ID can't change.
	var input Item
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item.Name, item.Stock, item.Price = input.Name, input.Stock, input.Price
	if err := db.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update item"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

// deleteItem handles DELETE /inventory/:id (200, or 404 if missing)
func deleteItem(c *gin.Context) {
	item, ok := findItem(c, c.Param("id"))
	if !ok {
		return
	}
	if err := db.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete item"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Item deleted successfully"})
}

// ============================================================
// MIDDLEWARE
// ============================================================

// TokenBucket is a token bucket rate limiter: it holds up to `capacity` tokens
// (the burst size) and gains one token every `rate`. Each request spends one token.
type TokenBucket struct {
	capacity   int
	tokens     int
	rate       time.Duration
	lastFilled time.Time
	mu         sync.Mutex
}

// NewTokenBucket returns a full bucket.
func NewTokenBucket(capacity int, rate time.Duration) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		rate:       rate,
		lastFilled: time.Now(),
	}
}

// Allow refills the bucket for the time elapsed, then spends a token if one is available.
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	addTokens := int(now.Sub(tb.lastFilled) / tb.rate)

	if addTokens > 0 {
		tb.tokens += addTokens
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		// Advance only by the time actually credited so partial intervals aren't lost.
		tb.lastFilled = tb.lastFilled.Add(time.Duration(addTokens) * tb.rate)
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

// rateLimiter limits requests with one shared bucket: burst of 5, then 1 request per second.
func rateLimiter() gin.HandlerFunc {
	limiter := NewTokenBucket(5, time.Second)
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

// ============================================================
// ROUTES
// ============================================================

// setupRouter builds the Gin engine with middleware and all routes.
//
// TODO(8, optional): Swagger UI route + swag annotations on handlers, then run `swag init`.
// TODO(8, optional): Cache GET /inventory/:id in an in-memory map (RWMutex + TTL); invalidate on PUT/DELETE.
func setupRouter() *gin.Engine {
	router := gin.Default()
	router.Use(rateLimiter())

	inventory := router.Group("/inventory")
	{
		inventory.GET("", getItems)
		inventory.POST("", createItem)
		inventory.GET("/:id", getItem)
		inventory.PUT("/:id", updateItem)
		inventory.DELETE("/:id", deleteItem)
	}

	return router
}

// ============================================================
// MAIN
// ============================================================

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables only")
	}

	initDatabase()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := setupRouter()
	log.Printf("Inventory API listening on http://localhost:%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
