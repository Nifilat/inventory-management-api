package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Each handler is a separate function (rubric). All use GORM via database.DB and return JSON.
// Every stub currently answers 501 so the server runs end to end; replace each body.

// GetItems handles GET /inventory
//
// TODO(4): Fetch all items with database.DB.Find(&items); return 200 + JSON.
// TODO(6): Pagination: read ?limit= and ?offset= (sensible defaults and a max limit), apply Limit/Offset.
// TODO(6): Sorting: read ?sort_by= (name|stock|price, whitelist it!) and ?order= (asc|desc), apply Order().
// TODO(6): Filtering: ?name= (partial, case-insensitive match, e.g. ILIKE) and ?min_stock= (stock >= n).
// TODO(6): Return 400 for invalid query params (non-numeric limit, unknown sort field).
func GetItems(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// GetItem handles GET /inventory/:id
//
// TODO(4): Look up by c.Param("id") with First(); 200 + item, or 404 + {"error": "item not found"}.
func GetItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// CreateItem handles POST /inventory
//
// TODO(4): Bind JSON body (ShouldBindJSON), return 400 on invalid/missing fields.
// TODO(4): Assign a fresh UUID, database.DB.Create(&item), return 201 + created item.
func CreateItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// UpdateItem handles PUT /inventory/:id
//
// TODO(4): 404 if the item doesn't exist; 400 on bad body; otherwise update fields, Save(), return 200 + item.
func UpdateItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// DeleteItem handles DELETE /inventory/:id
//
// TODO(4): 404 if the item doesn't exist; otherwise Delete() and return 200 with a message (or 204).
func DeleteItem(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
