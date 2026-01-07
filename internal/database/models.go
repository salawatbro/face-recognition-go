package database

import (
	"errors"
	"time"
)

// User represents a registered user in the system
type User struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Email     string                 `json:"email,omitempty"`
	Phone     string                 `json:"phone,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Faces     []Face                 `json:"faces"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// Face represents a face image and its embedding
type Face struct {
	ID           string    `json:"id"`
	Filename     string    `json:"filename"`
	Embedding    []float32 `json:"embedding"`
	EnrolledAt   time.Time `json:"enrolled_at"`
	QualityScore float64   `json:"quality_score"`
}

// Database represents the entire database structure
type Database struct {
	Version  string   `json:"version"`
	Users    []User   `json:"users"`
	Settings Settings `json:"settings"`
}

// Settings stores global configuration
type Settings struct {
	MatchThreshold     float64 `json:"match_threshold"`
	MaxFacesPerUser    int     `json:"max_faces_per_user"`
	EmbeddingDimension int     `json:"embedding_dimension"`
}

// MatchResult represents an identification result
type MatchResult struct {
	UserID     string
	User       *User
	FaceID     string
	Confidence float64
	Matched    bool
}

// Validation errors
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

// Validate checks if the User struct has valid data
func (u *User) Validate() error {
	if u.ID == "" {
		return ErrInvalidID
	}
	if u.Name == "" {
		return ErrEmptyName
	}
	if len(u.Name) > 100 {
		return errors.New("name exceeds maximum length of 100 characters")
	}
	return nil
}

// Validate checks if the Face struct has valid data
func (f *Face) Validate() error {
	if f.ID == "" {
		return ErrInvalidID
	}
	if f.Filename == "" {
		return errors.New("filename cannot be empty")
	}
	if len(f.Embedding) == 0 {
		return errors.New("embedding cannot be empty")
	}
	if f.QualityScore < 0 || f.QualityScore > 1 {
		return errors.New("quality score must be between 0 and 1")
	}
	return nil
}

// NewDatabase creates a new database with default settings
func NewDatabase() *Database {
	return &Database{
		Version: "1.0",
		Users:   []User{},
		Settings: Settings{
			MatchThreshold:     0.6,
			MaxFacesPerUser:    10,
			EmbeddingDimension: 128,
		},
	}
}
