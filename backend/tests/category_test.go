package tests

import (
	"context"
	"testing"

	"medieval-store/config"
	"medieval-store/models"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

// ==========================================
// Category model — round-trip + uniqueness (B5)
// ==========================================

func TestCategory_RoundTrip(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("categories")

	collection := config.MongoClient.Database(config.MongoDBName).Collection("categories")
	_, err := collection.InsertOne(context.Background(), models.Category{
		Name: "Weapons",
	})
	assert.NoError(t, err)

	var fetched models.Category
	err = collection.FindOne(context.Background(), bson.M{"name": "Weapons"}).Decode(&fetched)
	assert.NoError(t, err)
	assert.Equal(t, "Weapons", fetched.Name)
	assert.False(t, fetched.ID.IsZero(), "Mongo should populate _id on insert")
}

func TestCategory_NameMustBeUnique(t *testing.T) {
	// The unique index in config/mongo.go must reject duplicate category names.
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("categories")

	collection := config.MongoClient.Database(config.MongoDBName).Collection("categories")

	_, err := collection.InsertOne(context.Background(), models.Category{Name: "Spells"})
	assert.NoError(t, err)

	// Second insert with the same name must fail.
	_, err = collection.InsertOne(context.Background(), models.Category{Name: "Spells"})
	assert.Error(t, err)
}
