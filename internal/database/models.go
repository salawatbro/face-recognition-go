package database

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Custom types for GORM JSON handling

// Metadata is a custom type for storing JSON metadata
type Metadata map[string]interface{}

// Scan implements sql.Scanner interface
func (m *Metadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(map[string]interface{})
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("invalid type for Metadata")
	}

	if len(bytes) == 0 {
		*m = make(map[string]interface{})
		return nil
	}

	return json.Unmarshal(bytes, m)
}

// Value implements driver.Valuer interface
func (m Metadata) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	return json.Marshal(m)
}

// Embedding is a custom type for storing float32 arrays as JSON
type Embedding []float32

// Scan implements sql.Scanner interface
func (e *Embedding) Scan(value interface{}) error {
	if value == nil {
		*e = []float32{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("invalid type for Embedding")
	}

	if len(bytes) == 0 {
		*e = []float32{}
		return nil
	}

	return json.Unmarshal(bytes, e)
}

// Value implements driver.Valuer interface
func (e Embedding) Value() (driver.Value, error) {
	if e == nil {
		return "[]", nil
	}
	return json.Marshal(e)
}

// User represents a registered user in the system
type User struct {
	ID        string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Email     string    `gorm:"type:varchar(255)" json:"email,omitempty"`
	Phone     string    `gorm:"type:varchar(50)" json:"phone,omitempty"`
	Metadata  Metadata  `gorm:"type:text" json:"metadata,omitempty"`
	Faces     []Face    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"faces"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// Face represents a face image and its embedding
type Face struct {
	ID           string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID       string    `gorm:"type:varchar(36);not null;index" json:"user_id"`
	Filename     string    `gorm:"type:varchar(255);not null" json:"filename"`
	Embedding    Embedding `gorm:"type:text;not null" json:"embedding"`
	QualityScore float64   `gorm:"type:real;not null;default:0" json:"quality_score"`
	EnrolledAt   time.Time `gorm:"not null" json:"enrolled_at"`
}

// TableName specifies the table name for Face
func (Face) TableName() string {
	return "faces"
}

// Settings stores global configuration
type Settings struct {
	ID                 int     `gorm:"primaryKey" json:"id"`
	MatchThreshold     float64 `gorm:"type:real;not null;default:0.6" json:"match_threshold"`
	MaxFacesPerUser    int     `gorm:"not null;default:10" json:"max_faces_per_user"`
	EmbeddingDimension int     `gorm:"not null;default:128" json:"embedding_dimension"`
}

// TableName specifies the table name for Settings
func (Settings) TableName() string {
	return "settings"
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

// DefaultSettings returns default settings
func DefaultSettings() *Settings {
	return &Settings{
		ID:                 1,
		MatchThreshold:     0.6,
		MaxFacesPerUser:    10,
		EmbeddingDimension: 128,
	}
}
