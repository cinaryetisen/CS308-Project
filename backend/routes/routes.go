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

	router.Static("/images", "./images")

	api := router.Group("/api")

	api.POST("/signup", controllers.Signup)
	api.POST("/login", controllers.Login)
	api.GET("/products", controllers.GetProducts)
	api.GET("/products/:id", controllers.GetProduct)
	api.GET("/products/:id/reviews", controllers.GetProductReviews)
	api.GET("/products/:id/ratings", controllers.GetProductRatings)

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
	protected.GET("/orders/:id/invoice", controllers.DownloadInvoice)

	//Post review
	protected.POST("/reviews", controllers.CreateReview)
	protected.POST("/ratings", controllers.CreateRating)
	protected.GET("/me/ratings/:productId", controllers.GetMyRatingForProduct)

	//Product manager routes
	product_manager := router.Group("/api", security.AuthMiddleware(), security.Authorize("product_manager"))
	product_manager.GET("/deliveries", controllers.GetDeliveryList)
	product_manager.PATCH("/deliveries/:id/status", controllers.UpdateOrderStatus)
	product_manager.GET("/reviews/pending", controllers.GetPendingReviews)
	product_manager.PATCH("/reviews/:id/moderate", controllers.ModerateReview)

	//Sales manager routes
	sales_manager := router.Group("/api/admin", security.AuthMiddleware(), security.Authorize("sales_manager"))
	sales_manager.PATCH("/products/:id/price", controllers.UpdateProductPrice)
	sales_manager.PATCH("/products/:id/discount", controllers.SetProductDiscount)
	sales_manager.GET("/refunds", controllers.GetRefundRequests)
	sales_manager.PATCH("/refunds/:id", controllers.ResolveRefund)

	//Customer refund routes
	protected.POST("/orders/:id/refund", controllers.RequestRefund)
	protected.GET("/orders/me/refunds", controllers.GetMyRefunds)

	return router
}
