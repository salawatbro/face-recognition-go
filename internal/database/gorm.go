package database

import (
	"embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// GormDatabase implements Database interface using GORM
type GormDatabase struct {
	db     *gorm.DB
	dbType DatabaseType
}

// NewSQLiteDatabase creates a new SQLite database instance using GORM
func NewSQLiteDatabase(filePath string) (*GormDatabase, error) {
	db, err := gorm.Open(sqlite.Open(filePath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Enable foreign keys for SQLite
	db.Exec("PRAGMA foreign_keys = ON")

	gdb := &GormDatabase{db: db, dbType: DatabaseTypeSQLite}

	if err := gdb.runMigrations(filePath); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Ensure default settings exist
	if err := gdb.ensureDefaultSettings(); err != nil {
		return nil, fmt.Errorf("failed to create default settings: %w", err)
	}

	return gdb, nil
}

// NewPostgresDatabase creates a new PostgreSQL database instance using GORM
func NewPostgresDatabase(dsn string) (*GormDatabase, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres database: %w", err)
	}

	gdb := &GormDatabase{db: db, dbType: DatabaseTypePostgres}

	if err := gdb.runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Ensure default settings exist
	if err := gdb.ensureDefaultSettings(); err != nil {
		return nil, fmt.Errorf("failed to create default settings: %w", err)
	}

	return gdb, nil
}

// runMigrations runs database migrations
func (g *GormDatabase) runMigrations(connectionString string) error {
	d, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	var dbURL string
	if g.dbType == DatabaseTypeSQLite {
		dbURL = fmt.Sprintf("sqlite://%s", connectionString)
	} else {
		dbURL = connectionString
		if !strings.HasPrefix(dbURL, "postgres://") {
			dbURL = "postgres://" + dbURL
		}
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dbURL)
	if err != nil {
		// If migrations fail, fall back to auto-migrate
		return g.autoMigrate()
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		// Fall back to auto-migrate if migrations fail
		return g.autoMigrate()
	}

	return nil
}

// autoMigrate uses GORM's auto-migration as fallback
func (g *GormDatabase) autoMigrate() error {
	return g.db.AutoMigrate(&User{}, &Face{}, &Settings{})
}

// ensureDefaultSettings creates default settings if not exists
func (g *GormDatabase) ensureDefaultSettings() error {
	var count int64
	g.db.Model(&Settings{}).Count(&count)
	if count == 0 {
		return g.db.Create(DefaultSettings()).Error
	}
	return nil
}

// CreateUser adds a new user to the database
func (g *GormDatabase) CreateUser(user *User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	if err := user.Validate(); err != nil {
		return err
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	if user.Faces == nil {
		user.Faces = []Face{}
	}
	if user.Metadata == nil {
		user.Metadata = make(Metadata)
	}

	result := g.db.Create(user)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "UNIQUE") ||
			strings.Contains(result.Error.Error(), "duplicate") {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", result.Error)
	}

	return nil
}

// GetUser retrieves a user by ID
func (g *GormDatabase) GetUser(id string) (*User, error) {
	var user User
	result := g.db.Preload("Faces").First(&user, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", result.Error)
	}
	return &user, nil
}

// GetUserByName retrieves a user by name
func (g *GormDatabase) GetUserByName(name string) (*User, error) {
	var user User
	result := g.db.Preload("Faces").First(&user, "name = ?", name)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by name: %w", result.Error)
	}
	return &user, nil
}

// UpdateUser updates an existing user
func (g *GormDatabase) UpdateUser(user *User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	user.UpdatedAt = time.Now()

	result := g.db.Model(user).Updates(map[string]interface{}{
		"name":       user.Name,
		"email":      user.Email,
		"phone":      user.Phone,
		"metadata":   user.Metadata,
		"updated_at": user.UpdatedAt,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// DeleteUser removes a user from the database
func (g *GormDatabase) DeleteUser(id string) error {
	result := g.db.Delete(&User{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// ListUsers returns all users in the database
func (g *GormDatabase) ListUsers() ([]User, error) {
	var users []User
	result := g.db.Preload("Faces").Order("created_at DESC").Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list users: %w", result.Error)
	}

	if users == nil {
		users = []User{}
	}

	return users, nil
}

// AddFace adds a face to a user
func (g *GormDatabase) AddFace(userID string, face *Face) error {
	// Check if user exists
	var user User
	if err := g.db.First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Check max faces
	settings, err := g.GetSettings()
	if err != nil {
		return err
	}

	var faceCount int64
	g.db.Model(&Face{}).Where("user_id = ?", userID).Count(&faceCount)
	if int(faceCount) >= settings.MaxFacesPerUser {
		return ErrMaxFacesReached
	}

	if face.ID == "" {
		face.ID = uuid.New().String()
	}

	if err := face.Validate(); err != nil {
		return err
	}

	face.UserID = userID
	face.EnrolledAt = time.Now()

	if err := g.db.Create(face).Error; err != nil {
		return fmt.Errorf("failed to add face: %w", err)
	}

	// Update user's updated_at
	g.db.Model(&User{}).Where("id = ?", userID).Update("updated_at", time.Now())

	return nil
}

// RemoveFace removes a face from a user
func (g *GormDatabase) RemoveFace(userID, faceID string) error {
	result := g.db.Where("id = ? AND user_id = ?", faceID, userID).Delete(&Face{})
	if result.Error != nil {
		return fmt.Errorf("failed to remove face: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("face with ID %s not found", faceID)
	}

	// Update user's updated_at
	g.db.Model(&User{}).Where("id = ?", userID).Update("updated_at", time.Now())

	return nil
}

// GetAllEmbeddings returns a map of userID to faces for matching
func (g *GormDatabase) GetAllEmbeddings() (map[string][]Face, error) {
	var faces []Face
	result := g.db.Find(&faces)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get embeddings: %w", result.Error)
	}

	embeddings := make(map[string][]Face)
	for _, face := range faces {
		embeddings[face.UserID] = append(embeddings[face.UserID], face)
	}

	return embeddings, nil
}

// GetSettings returns the current settings
func (g *GormDatabase) GetSettings() (*Settings, error) {
	var settings Settings
	result := g.db.First(&settings, "id = ?", 1)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Create default settings
			settings = *DefaultSettings()
			if err := g.db.Create(&settings).Error; err != nil {
				return nil, fmt.Errorf("failed to create default settings: %w", err)
			}
			return &settings, nil
		}
		return nil, fmt.Errorf("failed to get settings: %w", result.Error)
	}
	return &settings, nil
}

// UpdateSettings updates the database settings
func (g *GormDatabase) UpdateSettings(settings *Settings) error {
	settings.ID = 1
	result := g.db.Save(settings)
	if result.Error != nil {
		return fmt.Errorf("failed to update settings: %w", result.Error)
	}
	return nil
}

// Close closes the database connection
func (g *GormDatabase) Close() error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB returns the underlying GORM database (for advanced usage)
func (g *GormDatabase) GetDB() *gorm.DB {
	return g.db
}
