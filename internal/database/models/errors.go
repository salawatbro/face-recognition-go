package models

import "errors"

// Database errors
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrFaceNotDetected   = errors.New("no face detected in image")
	ErrMultipleFaces     = errors.New("multiple faces detected, expected one")
	ErrNoMatch           = errors.New("no matching user found")
	ErrInvalidImage      = errors.New("invalid image format")
	ErrDatabaseCorrupt   = errors.New("database file is corrupted")
	ErrMaxFacesReached   = errors.New("maximum faces per user reached")
	ErrEmptyName         = errors.New("user name cannot be empty")
	ErrInvalidID         = errors.New("invalid user or face ID")
)
