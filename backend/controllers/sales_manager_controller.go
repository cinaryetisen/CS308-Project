package controllers

import (
	"context"
	"net/http"
	"sort"
	"time"

	"medieval-store/config"
	"medieval-store/errs"
	"medieval-store/models"
	"medieval-store/services"

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
		collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
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

	// ?format=pdf streams every invoice in range as ONE multi-page document —
	// the "save all invoices" export for the sales manager.
	if c.Query("format") == "pdf" {
		bundles, err := buildInvoiceBundles(orders)
		if err != nil {
			errs.Abort(c, errs.InternalError)
			return
		}
		pdfBytes, err := services.GenerateBulkInvoicePDF(bundles)
		if err != nil {
			errs.Abort(c, errs.InternalError)
			return
		}
		c.Header("Content-Disposition", `attachment; filename="invoices_`+from+`_`+to+`.pdf"`)
		c.Data(http.StatusOK, "application/pdf", pdfBytes)
		return
	}

	c.JSON(http.StatusOK, orders)
}

// buildInvoiceBundles resolves the customers and product names for a set of
// orders so the PDF renderer has everything it needs.
func buildInvoiceBundles(orders []models.Order) ([]services.InvoiceBundle, error) {
	// Collect unique customer IDs and product ObjectIDs across all orders.
	customerIDs := make(map[uint]bool)
	productIDs := make(map[string]primitive.ObjectID)
	for _, o := range orders {
		customerIDs[o.CustomerID] = true
		for _, item := range o.Items {
			if oid, err := primitive.ObjectIDFromHex(item.ProductID); err == nil {
				productIDs[item.ProductID] = oid
			}
		}
	}

	// Fetch all customers in one query.
	ids := make([]uint, 0, len(customerIDs))
	for id := range customerIDs {
		ids = append(ids, id)
	}
	userMap := make(map[uint]models.User)
	if len(ids) > 0 {
		var users []models.User
		if err := config.DB.Where("id IN ?", ids).Find(&users).Error; err != nil {
			return nil, err
		}
		for _, u := range users {
			userMap[u.ID] = u
		}
	}

	// Fetch all referenced products in one query.
	productMap := make(map[string]models.Product)
	if len(productIDs) > 0 {
		var objIDs []primitive.ObjectID
		for _, oid := range productIDs {
			objIDs = append(objIDs, oid)
		}
		collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cursor, err := collection.Find(ctx, bson.M{"_id": bson.M{"$in": objIDs}})
		if err == nil {
			var products []models.Product
			if err := cursor.All(ctx, &products); err == nil {
				for _, p := range products {
					productMap[p.ID.Hex()] = p
				}
			}
		}
		// Product lookup failures degrade to "Unknown Artifact" names, not errors.
	}

	bundles := make([]services.InvoiceBundle, 0, len(orders))
	for _, o := range orders {
		bundles = append(bundles, services.InvoiceBundle{
			User:       userMap[o.CustomerID],
			Order:      o,
			Items:      o.Items,
			ProductMap: productMap,
		})
	}
	return bundles, nil
}
