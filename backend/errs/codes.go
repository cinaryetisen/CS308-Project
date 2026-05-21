package errs

type Code string

const (
	// Auth / JWT
	AuthInvalidCreds    Code = "AUTH_INVALID_CREDENTIALS"
	AuthUserExists      Code = "AUTH_USER_ALREADY_EXISTS"
	AuthWeakPassword    Code = "AUTH_WEAK_PASSWORD"
	AuthInvalidEmail    Code = "AUTH_INVALID_EMAIL"
	AuthMissingHeader   Code = "AUTH_MISSING_HEADER"
	AuthInvalidHeader   Code = "AUTH_INVALID_HEADER"
	AuthInvalidToken    Code = "AUTH_INVALID_TOKEN"
	AuthClaimsParseFail Code = "AUTH_CLAIMS_PARSE_FAIL"

	// User
	UserNotFound     Code = "USER_NOT_FOUND"
	UserUnauthorized Code = "USER_UNAUTHORIZED"
	UserForbidden    Code = "USER_FORBIDDEN"

	// Product
	ProductNotFound   Code = "PRODUCT_NOT_FOUND"
	ProductInvalidID  Code = "PRODUCT_INVALID_ID"
	ProductOutOfStock Code = "PRODUCT_OUT_OF_STOCK"

	// Order
	OrderNotFound       Code = "ORDER_NOT_FOUND"
	OrderForbidden      Code = "ORDER_FORBIDDEN"
	OrderInvalidStatus  Code = "ORDER_INVALID_STATUS"
	OrderNotCancellable Code = "ORDER_NOT_CANCELLABLE"

	// Cart
	CartItemNotFound Code = "CART_ITEM_NOT_FOUND"

	// Wishlist
	WishlistItemNotFound Code = "WISHLIST_ITEM_NOT_FOUND"

	// Refund
	RefundWindowExpired   Code = "REFUND_WINDOW_EXPIRED"
	RefundIneligibleOrder Code = "REFUND_INELIGIBLE_ORDER"
	RefundAlreadyExists   Code = "REFUND_ALREADY_EXISTS"
	RefundAlreadyResolved Code = "REFUND_ALREADY_RESOLVED"
	RefundNotFound        Code = "REFUND_NOT_FOUND"
	RefundItemMismatch    Code = "REFUND_ITEM_MISMATCH"

	// Review
	ReviewUnauthorized  Code = "REVIEW_UNAUTHORIZED"
	ReviewInvalidAction Code = "REVIEW_INVALID_ACTION"
	ReviewNotFound      Code = "REVIEW_NOT_FOUND"

	// Rating
	RatingUnauthorized Code = "RATING_UNAUTHORIZED"

	// Category
	CategoryDuplicate Code = "CATEGORY_DUPLICATE"
	CategoryNotFound  Code = "CATEGORY_NOT_FOUND"
	CategoryInvalidID Code = "CATEGORY_INVALID_ID"

	// Validation
	InvalidJSON       Code = "INVALID_JSON"
	InvalidDateFormat Code = "INVALID_DATE_FORMAT"

	// Fallback
	InternalError Code = "INTERNAL_SERVER_ERROR"
)
