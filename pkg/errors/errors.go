package errs

import "errors"

var ErrNotFound = errors.New("not found")
var ErrDuplicateSub = errors.New("ERROR: duplicate key value violates unique constraint \"subscriber_pkey\" (SQLSTATE 23505)")
