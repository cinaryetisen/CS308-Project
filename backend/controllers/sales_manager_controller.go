package controllers

import (
	"context"
	"net/http"
	"sort"
	"time"

	"medieval-store/config"
	"medieval-store/errs"
	"medieval-store/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Data structures for the JSON response
type DailyStat struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Profit  float64 `json:"profit"`
}

type RevenueResponse struct {
	TotalRevenue float64     `json:"total_revenue"`
	TotalCost    float64     `json:"total_cost"`
	Profit       float64     `json:"profit"`
	Daily        []DailyStat `json:"daily"`
}

// GetRevenue calculates cross-database profit metrics for the Sales Manager
func GetRevenue(c *gin.Context) {
	// 1. Parse Date Range from URL Query Parameters
	fromStr := c.Query("from")
	toStr := c.Query("to")

	fromTime, errFrom := time.Parse("2006-01-02", fromStr)
	toTime, errTo := time.Parse("2006-01-02", toStr)

	if errFrom != nil || errTo != nil {
		errs.Abort(c, errs.InvalidDateFormat)
		return
	}

	// Adjust 'toTime' to include the entire final day (up to 23:59:59)
	toTime = toTime.Add(24 * time.Hour).Add(-time.Nanosecond)

	// 2. Query PostgreSQL for valid orders within the timeframe
	var orders []models.Order
	// We only count orders that are not cancelled or returned
	validStatuses := []string{"processing", "in-transit", "delivered"}

	if err := config.DB.Preload("Items").
		Where("created_at >= ? AND created_at <= ?", fromTime, toTime).
		Where("status IN ?", validStatuses).
		Find(&orders).Error; err != nil {
		errs.Abort(c, errs.InternalError)
		return
	}

	// 3. Extract unique Product IDs to query MongoDB efficiently
	uniqueProductIDs := make(map[string]primitive.ObjectID)
	for _, order := range orders {
		for _, item := range order.Items {
			if oid, err := primitive.ObjectIDFromHex(item.ProductID); err == nil {
				uniqueProductIDs[item.ProductID] = oid
			}
		}
	}

	// Convert map to a slice of ObjectIDs for the Mongo query
	var objectIDs []primitive.ObjectID
	for _, oid := range uniqueProductIDs {
		objectIDs = append(objectIDs, oid)
	}

	// 4. Query MongoDB for Product Costs
	costMap := make(map[string]float64)
	if len(objectIDs) > 0 {
		collection := config.MongoClient.Database("medieval_store").Collection("products")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := collection.Find(ctx, bson.M{"_id": bson.M{"$in": objectIDs}})
		if err == nil {
			var products []models.Product
			if err = cursor.All(ctx, &products); err == nil {
				// Map the ID back to the cost for fast lookups
				for _, p := range products {
					costMap[p.ID.Hex()] = p.Cost
				}
			}
		}
		// If MongoDB fails, costMap stays empty: revenue is still returned,
		// profit will show as equal to revenue (cost treated as 0).
	}

	// 5. Calculate Revenue, Cost, and Profit
	var response RevenueResponse
	dailyAggregator := make(map[string]*DailyStat)

	for _, order := range orders {
		dateStr := order.CreatedAt.Format("2006-01-02")

		// Initialize the daily bucket if it doesn't exist yet
		if _, exists := dailyAggregator[dateStr]; !exists {
			dailyAggregator[dateStr] = &DailyStat{Date: dateStr}
		}

		for _, item := range order.Items {
			// Revenue is what the customer actually paid at the time (stored in PG)
			revenue := item.Price * float64(item.Quantity)

			// Cost is what the store pays (fetched from Mongo)
			cost := costMap[item.ProductID] * float64(item.Quantity)

			profit := revenue - cost

			// Add to overall totals
			response.TotalRevenue += revenue
			response.TotalCost += cost
			response.Profit += profit

			// Add to daily totals
			dailyAggregator[dateStr].Revenue += revenue
			dailyAggregator[dateStr].Profit += profit
		}
	}

	// 6. Format and sort the Daily array for the frontend charts
	for _, stat := range dailyAggregator {
		response.Daily = append(response.Daily, *stat)
	}

	// Sort the daily stats chronologically so the frontend chart looks correct
	sort.Slice(response.Daily, func(i, j int) bool {
		return response.Daily[i].Date < response.Daily[j].Date
	})

	// Return data to frontend
	c.JSON(http.StatusOK, response)
}

// Returns all orders (with items) created within a date range.
func GetInvoicesByDateRange(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")

	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "'from' and 'to' query parameters are required (YYYY-MM-DD)"})
		return
	}

	fromTime, err := time.Parse("2006-01-02", from)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'from' date. Use YYYY-MM-DD"})
		return
	}
	toTime, err := time.Parse("2006-01-02", to)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'to' date. Use YYYY-MM-DD"})
		return
	}

	// Include the entire final day.
	toTime = toTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	var orders []models.Order
	if err := config.DB.Preload("Items").
		Where("created_at BETWEEN ? AND ?", fromTime, toTime).
		Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch invoices"})
		return
	}

	c.JSON(http.StatusOK, orders)
}
