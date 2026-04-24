package errs

import "net/http"

type errorMeta struct {
	status  int
	message string
}

var registry = map[Code]errorMeta{
	AuthInvalidCreds: {http.StatusUnauthorized, "Invalid email or password."},
	AuthUserExists:   {http.StatusConflict, "An account with this email already exists."},
	AuthWeakPassword: {http.StatusBadRequest, "Password does not meet security requirements."},
	AuthInvalidEmail: {http.StatusBadRequest, "The email address provided is not valid."},
	InternalError:    {http.StatusInternalServerError, "An unexpected error occurred."},
}
