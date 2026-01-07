package models

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

// DefaultSettings returns default settings
func DefaultSettings() *Settings {
	return &Settings{
		ID:                 1,
		MatchThreshold:     0.6,
		MaxFacesPerUser:    10,
		EmbeddingDimension: 128,
	}
}
