package models

import (
	"errors"
	"time"
)

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
