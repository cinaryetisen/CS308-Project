package services

import (
	"log"

	"medieval-store/config"
	"medieval-store/models"
)

// FindWishlistUsersForProduct returns every user who has the product in their wishlist
func FindWishlistUsersForProduct(productID string) ([]models.User, error) {
	var users []models.User
	err := config.DB.
		Joins("JOIN wishlist_items ON wishlist_items.user_id = users.id").
		Where("wishlist_items.product_id = ?", productID).Find(&users).Error
	return users, err
}

// NotifyWishlistOfDiscount emails every user whose wishlist contains the product that a discount has just been applied.
func NotifyWishlistOfDiscount(product models.Product, newDiscount float64) {
	users, err := FindWishlistUsersForProduct(product.ID.Hex())
	if err != nil {
		log.Printf("discount notifier: failed to look up wishlist users for %s: %v", product.ID.Hex(), err)
		return
	}

	for _, u := range users {
		if err := SendDiscountNotificationEmail(u.Email, u.Name, product.Name, product.Price, newDiscount); err != nil {
			log.Printf("discount notifier: send to %s failed: %v", u.Email, err)
		}
	}
}
