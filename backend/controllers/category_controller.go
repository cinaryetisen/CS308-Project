package controllers

import (
	"context"
	"net/http"
	"time"

	"medieval-store/config"
	"medieval-store/errs"
	"medieval-store/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetCategories returns every category, sorted alphabetically by name
func GetCategories(c *gin.Context) {
	collection := config.MongoClient.Database(config.MongoDBName).Collection("categories")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}
	defer cursor.Close(ctx)

	var categories []models.Category
	if err := cursor.All(ctx, &categories); err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	if categories == nil {
		categories = []models.Category{}
	}

	c.JSON(http.StatusOK, categories)
}

func CreateCategory(c *gin.Context) {
	var input struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		errs.AbortWithDetail(c, errs.InvalidJSON, err.Error())
		return
	}

	category := models.Category{
		Name:      input.Name,
		CreatedAt: time.Now(),
	}

	collection := config.MongoClient.Database(config.MongoDBName).Collection("categories")
	result, err := collection.InsertOne(context.Background(), category)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			errs.Abort(c, errs.CategoryDuplicate)
			return
		}
		errs.Abort(c, errs.InternalError)
		return
	}

	category.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, category)
}

func DeleteCategory(c *gin.Context) {
	idParam := c.Param("id")

	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		errs.Abort(c, errs.CategoryInvalidID)
		return
	}

	collection := config.MongoClient.Database(config.MongoDBName).Collection("categories")
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": objID})
	if err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}
	if result.DeletedCount == 0 {
		errs.Abort(c, errs.CategoryNotFound)
		return
	}

	c.Status(http.StatusNoContent)
}
