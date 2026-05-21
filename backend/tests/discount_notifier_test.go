package tests

import (
	"testing"

	"medieval-store/config"
	"medieval-store/models"
	"medieval-store/services"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ==========================================
// FindWishlistUsersForProduct (A4 testable join)
// ==========================================

func TestFindWishlistUsersForProduct_ReturnsOnlyWishlistMembers(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	productHex := primitive.NewObjectID().Hex()
	otherProductHex := primitive.NewObjectID().Hex()

	// 3 users; user 1 has the target wishlisted, user 2 has it AND another,
	// user 3 has only the unrelated product.
	config.DB.Create(&models.User{ID: 1, Name: "Arthur", Email: "arthur@camelot.com"})
	config.DB.Create(&models.User{ID: 2, Name: "Lancelot", Email: "lance@knight.com"})
	config.DB.Create(&models.User{ID: 3, Name: "Galahad", Email: "gala@knight.com"})

	config.DB.Create(&models.WishlistItem{UserID: 1, ProductID: productHex})
	config.DB.Create(&models.WishlistItem{UserID: 2, ProductID: productHex})
	config.DB.Create(&models.WishlistItem{UserID: 2, ProductID: otherProductHex})
	config.DB.Create(&models.WishlistItem{UserID: 3, ProductID: otherProductHex})

	users, err := services.FindWishlistUsersForProduct(productHex)
	assert.NoError(t, err)
	assert.Len(t, users, 2)

	emails := []string{}
	for _, u := range users {
		emails = append(emails, u.Email)
	}
	assert.Contains(t, emails, "arthur@camelot.com")
	assert.Contains(t, emails, "lance@knight.com")
	assert.NotContains(t, emails, "gala@knight.com")
}

func TestFindWishlistUsersForProduct_EmptyWhenNoneMatch(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	productHex := primitive.NewObjectID().Hex()
	config.DB.Create(&models.User{ID: 1, Name: "Solo", Email: "solo@me.com"})
	// No wishlist row for this user

	users, err := services.FindWishlistUsersForProduct(productHex)
	assert.NoError(t, err)
	assert.Empty(t, users)
}

func TestFindWishlistUsersForProduct_OneRowPerUser(t *testing.T) {
	// A user can wishlist a product only once (composite primary key on
	// WishlistItem), so the join should never duplicate a user in the result.
	setupTestDB()
	defer clearTestDB()

	productHex := primitive.NewObjectID().Hex()
	config.DB.Create(&models.User{ID: 1, Name: "Arthur", Email: "arthur@camelot.com"})
	config.DB.Create(&models.WishlistItem{UserID: 1, ProductID: productHex})

	users, err := services.FindWishlistUsersForProduct(productHex)
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "arthur@camelot.com", users[0].Email)
}
