package routes

import (
	"github.com/gin-gonic/gin"

	"inventory-api/handlers"
	"inventory-api/middleware"
)

// Setup builds the Gin engine with middleware and all routes.
//
// TODO(5): Confirm RateLimiter is applied with router.Use() before the routes.
// TODO(8, optional): Swagger UI route, e.g. router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler)).
// TODO(8, optional): Add swag annotations to each handler, then run `swag init`.
// TODO(8, optional): Cache GET /inventory/:id responses in an in-memory map (RWMutex + TTL); invalidate on PUT/DELETE.
func Setup() *gin.Engine {
	router := gin.Default()
	router.Use(middleware.RateLimiter())

	inventory := router.Group("/inventory")
	{
		inventory.GET("", handlers.GetItems)
		inventory.POST("", handlers.CreateItem)
		inventory.GET("/:id", handlers.GetItem)
		inventory.PUT("/:id", handlers.UpdateItem)
		inventory.DELETE("/:id", handlers.DeleteItem)
	}

	return router
}
