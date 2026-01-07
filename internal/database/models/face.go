package models

import (
	"errors"
	"time"
)

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

// MatchResult represents an identification result
type MatchResult struct {
	UserID     string
	User       *User
	FaceID     string
	Confidence float64
	Matched    bool
}
