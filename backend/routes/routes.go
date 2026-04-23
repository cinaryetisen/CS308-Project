package routes

import (
	"medieval-store/controllers"
	"medieval-store/security"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	//CORS setup
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Note: For production, you'll want to lock this down to your exact frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := router.Group("/api")

	api.POST("/signup", controllers.Signup)
	api.POST("/login", controllers.Login)
	api.GET("/products", controllers.GetProducts)
	api.GET("/products/:id", controllers.GetProduct)

	protected := router.Group("/api")
	protected.Use(security.AuthMiddleware())

	//Cart routes
	protected.GET("/cart", controllers.GetCart)
	protected.DELETE("/cart", controllers.ClearCart)
	protected.DELETE("/cart/:id", controllers.RemoveFromCart)
	protected.PATCH("/cart/item", controllers.AddToCart)
	protected.POST("/cart/merge", controllers.MergeCarts)

	//User Profile routes
	protected.GET("/users/me", controllers.GetProfile)
	protected.PATCH("/users/me", controllers.UpdateProfile)

	// Checkout & Orders
	protected.POST("/checkout", controllers.Checkout)
	protected.GET("/orders/me", controllers.GetMyOrders)

	//Product manager routes
	product_manager := router.Group("/api", security.AuthMiddleware(), security.Authorize("product_manager"))
	product_manager.GET("/deliveries", controllers.GetDeliveryList)
	product_manager.PATCH("/deliveries/:id/status", controllers.UpdateOrderStatus)

	return router
}
