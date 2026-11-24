package apperr

import "errors"

var (
	// ErrNotFound means the requested resource doesn't exist.
	ErrNotFound = errors.New("not found")
	// ErrConflict means the request clashes with the current state, such as
	// creating something that already exists.
	ErrConflict = errors.New("conflict")
)

type ValidationError struct {
	Fields map[string]string
}
