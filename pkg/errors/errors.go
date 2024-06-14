package errs

import "errors"

var ErrNotFound = errors.New("not found")
var ErrDuplicateSub = errors.New("ERROR: duplicate key value violates unique constraint \"subscriber_pkey\" (SQLSTATE 23505)")

var ErrMethodNotAllowed = "Method not found"
var ErrInternalServer = "Internal server error"
var ErrBadRequest = "Bad request"
var ErrUnauthorized = "Unauthorized"
var ErrNotFoundString = "Not found"
var ErrNoBirthEmailLogin = "Not have birthday, email, password or login"
var ErrAlreadyExists = "Already exists"
