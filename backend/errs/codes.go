package errs

type Code string

const (
	// Auth Errors
	AuthInvalidCreds Code = "AUTH_INVALID_CREDENTIALS"
	AuthUserExists   Code = "AUTH_USER_ALREADY_EXISTS"
	AuthWeakPassword Code = "AUTH_WEAK_PASSWORD"
	AuthInvalidEmail Code = "AUTH_INVALID_EMAIL"

	// Fallback
	InternalError Code = "INTERNAL_SERVER_ERROR"
)
