package main

import (
	"errors"
	"log"
	"net/http"
	"os"

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

// DB is the shared GORM handle used by the handlers.
var DB *gorm.DB

// initDatabase connects to PostgreSQL, migrates the schema and seeds initial data.
// The connection string (dsn) comes from DATABASE_DSN in .env.
//
// TODO(1): Handle the AutoMigrate error instead of ignoring it.
func initDatabase() {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("DATABASE_DSN is not set. Copy .env.example to .env and fill it in.")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database! ", err)
	}

	DB.AutoMigrate(&Item{})
	seedDatabase()
}

// seedDatabase seeds the database with 20 default items when the table is empty.
func seedDatabase() {
	var count int64
	DB.Model(&Item{}).Count(&count)
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

	if err := DB.Create(&items).Error; err != nil {
		log.Fatal("Failed to seed database: ", err)
	}
	log.Println("Database seeded with 20 sample items.")
}

// ============================================================
// HANDLERS
// One function per endpoint (rubric). All use GORM via DB and return JSON.
// ============================================================

// findItem loads an item by ID. On failure it writes the error response
// (404 if not found, 500 for any other database error) and returns false.
func findItem(c *gin.Context, id string) (Item, bool) {
	var item Item
	if err := DB.First(&item, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch item"})
		}
		return item, false
	}
	return item, true
}

// getItems handles GET /inventory
//
// TODO(6): Pagination: read ?limit= and ?offset= (defaults and a max limit), apply Limit/Offset.
// TODO(6): Sorting: read ?sort_by= (name|stock|price, whitelist it!) and ?order= (asc|desc), apply Order().
// TODO(6): Filtering: ?name= (partial, case-insensitive, e.g. ILIKE) and ?min_stock= (stock >= n).
// TODO(6): Return 400 for invalid query params (non-numeric limit, unknown sort field).
func getItems(c *gin.Context) {
	var items []Item
	if err := DB.Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch items"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
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
	if err := DB.Create(&newItem).Error; err != nil {
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
	if err := DB.Save(&item).Error; err != nil {
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
	if err := DB.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete item"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Item deleted successfully"})
}

// ============================================================
// MIDDLEWARE
// ============================================================

// rateLimiter limits requests per client.
//
// TODO(5): Run `go get golang.org/x/time/rate`, then create rate.NewLimiter(1, 5): 1 req/sec, burst of 5.
// TODO(5): Decide scope: one global limiter (simplest) or one per client IP (map + mutex, c.ClientIP()).
// TODO(5): If !limiter.Allow(), c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"}).
// TODO(7): Test by firing >5 quick requests (curl loop) and confirm 429s, then recovery.
func rateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // pass-through until implemented
	}
}

// ============================================================
// ROUTES
// ============================================================

// setupRouter builds the Gin engine with middleware and all routes.
//
// TODO(5): Confirm rateLimiter is applied with router.Use() before the routes.
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
