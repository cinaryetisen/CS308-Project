package errs

import "net/http"

type errorMeta struct {
	status  int
	message string
}

var registry = map[Code]errorMeta{
	// Auth / JWT
	AuthInvalidCreds:    {http.StatusUnauthorized, "Invalid email or password."},
	AuthUserExists:      {http.StatusConflict, "An account with this email already exists."},
	AuthWeakPassword:    {http.StatusBadRequest, "Password does not meet security requirements."},
	AuthInvalidEmail:    {http.StatusBadRequest, "The email address provided is not valid."},
	AuthMissingHeader:   {http.StatusUnauthorized, "Authorization header is missing."},
	AuthInvalidHeader:   {http.StatusUnauthorized, "Invalid Authorization header format."},
	AuthInvalidToken:    {http.StatusUnauthorized, "Invalid or expired token."},
	AuthClaimsParseFail: {http.StatusUnauthorized, "Failed to parse token claims."},

	// User
	UserNotFound:     {http.StatusNotFound, "User not found."},
	UserUnauthorized: {http.StatusUnauthorized, "User is not authenticated."},
	UserForbidden:    {http.StatusForbidden, "Access denied."},

	// Product
	ProductNotFound:   {http.StatusNotFound, "Product not found."},
	ProductInvalidID:  {http.StatusBadRequest, "Invalid product ID format."},
	ProductOutOfStock: {http.StatusConflict, "One or more items are out of stock."},

	// Order
	OrderNotFound:       {http.StatusNotFound, "Order not found."},
	OrderForbidden:      {http.StatusForbidden, "This order does not belong to you."},
	OrderInvalidStatus:  {http.StatusBadRequest, "Invalid order status value."},
	OrderNotCancellable: {http.StatusBadRequest, "Only orders in processing state can be cancelled."},

	// Cart
	CartItemNotFound: {http.StatusNotFound, "Item not found in cart."},

	// Wishlist
	WishlistItemNotFound: {http.StatusNotFound, "Item not found in wishlist."},

	// Refund
	RefundWindowExpired:   {http.StatusBadRequest, "The 30-day refund window for this order has passed."},
	RefundIneligibleOrder: {http.StatusBadRequest, "Only delivered orders are eligible for a refund."},
	RefundAlreadyExists:   {http.StatusConflict, "A refund request already exists for this item."},
	RefundAlreadyResolved: {http.StatusConflict, "This refund has already been resolved."},
	RefundNotFound:        {http.StatusNotFound, "Refund request not found."},
	RefundItemMismatch:    {http.StatusBadRequest, "Order item does not belong to this order."},

	// Review
	ReviewUnauthorized:  {http.StatusForbidden, "You can only review products that have been delivered to you."},
	ReviewInvalidAction: {http.StatusBadRequest, "Action must be 'approve' or 'reject'."},
	ReviewNotFound:      {http.StatusNotFound, "Review not found."},

	// Rating
	RatingUnauthorized: {http.StatusForbidden, "You can only rate products that have been delivered to you."},

	// Category
	CategoryDuplicate: {http.StatusConflict, "A category with that name already exists."},
	CategoryNotFound:  {http.StatusNotFound, "Category not found."},
	CategoryInvalidID: {http.StatusBadRequest, "Invalid category ID format."},

	// Validation
	InvalidJSON:       {http.StatusBadRequest, "Request body could not be parsed."},
	InvalidDateFormat: {http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD."},

	// Fallback
	InternalError: {http.StatusInternalServerError, "An unexpected error occurred."},
}
