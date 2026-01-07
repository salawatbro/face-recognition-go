package database

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// JSONDatabase implements a thread-safe JSON file-based database
type JSONDatabase struct {
	filePath string
	db       *Database
	mutex    sync.RWMutex
}

// NewJSONDatabase creates a new JSON database instance
func NewJSONDatabase(filePath string) (*JSONDatabase, error) {
	jdb := &JSONDatabase{
		filePath: filePath,
		db:       NewDatabase(),
	}

	if loadErr := jdb.Load(); loadErr != nil {
		if os.IsNotExist(loadErr) {
			if err := jdb.Save(); err != nil {
				return nil, fmt.Errorf("failed to create database file: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load database: %w", loadErr)
		}
	}

	return jdb, nil
}

// Load reads the database from disk
func (j *JSONDatabase) Load() error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	data, err := os.ReadFile(j.filePath)
	if err != nil {
		return err
	}

	db := NewDatabase()
	if err := json.Unmarshal(data, db); err != nil {
		return ErrDatabaseCorrupt
	}

	j.db = db
	return nil
}

// Save writes the database to disk with backup
func (j *JSONDatabase) Save() error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	backupPath := j.filePath + ".backup"
	if _, err := os.Stat(j.filePath); err == nil {
		if err := os.Rename(j.filePath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	data, err := json.MarshalIndent(j.db, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal database: %w", err)
	}

	if err := os.WriteFile(j.filePath, data, 0o600); err != nil {
		if _, statErr := os.Stat(backupPath); statErr == nil {
			_ = os.Rename(backupPath, j.filePath)
		}
		return fmt.Errorf("failed to write database: %w", err)
	}

	_ = os.Remove(backupPath)
	return nil
}

// CreateUser adds a new user to the database
func (j *JSONDatabase) CreateUser(user *User) error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	if err := user.Validate(); err != nil {
		return err
	}

	for i := range j.db.Users {
		if j.db.Users[i].ID == user.ID {
			return ErrUserAlreadyExists
		}
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	if user.Faces == nil {
		user.Faces = []Face{}
	}

	if user.Metadata == nil {
		user.Metadata = make(map[string]interface{})
	}

	j.db.Users = append(j.db.Users, *user)
	return j.saveInternal()
}

// GetUser retrieves a user by ID
func (j *JSONDatabase) GetUser(id string) (*User, error) {
	j.mutex.RLock()
	defer j.mutex.RUnlock()

	for i := range j.db.Users {
		if j.db.Users[i].ID == id {
			user := j.db.Users[i]
			return &user, nil
		}
	}

	return nil, ErrUserNotFound
}

// GetUserByName retrieves a user by name (case-sensitive)
func (j *JSONDatabase) GetUserByName(name string) (*User, error) {
	j.mutex.RLock()
	defer j.mutex.RUnlock()

	for i := range j.db.Users {
		if j.db.Users[i].Name == name {
			user := j.db.Users[i]
			return &user, nil
		}
	}

	return nil, ErrUserNotFound
}

// UpdateUser updates an existing user
func (j *JSONDatabase) UpdateUser(user *User) error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	if err := user.Validate(); err != nil {
		return err
	}

	for i := range j.db.Users {
		if j.db.Users[i].ID == user.ID {
			user.UpdatedAt = time.Now()
			user.CreatedAt = j.db.Users[i].CreatedAt
			j.db.Users[i] = *user
			return j.saveInternal()
		}
	}

	return ErrUserNotFound
}

// DeleteUser removes a user from the database
func (j *JSONDatabase) DeleteUser(id string) error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	for i := range j.db.Users {
		if j.db.Users[i].ID == id {
			j.db.Users = append(j.db.Users[:i], j.db.Users[i+1:]...)
			return j.saveInternal()
		}
	}

	return ErrUserNotFound
}

// ListUsers returns all users in the database
func (j *JSONDatabase) ListUsers() ([]User, error) {
	j.mutex.RLock()
	defer j.mutex.RUnlock()

	users := make([]User, len(j.db.Users))
	copy(users, j.db.Users)
	return users, nil
}

// AddFace adds a face to a user
func (j *JSONDatabase) AddFace(userID string, face *Face) error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	if err := face.Validate(); err != nil {
		return err
	}

	for i := range j.db.Users {
		if j.db.Users[i].ID != userID {
			continue
		}
		if len(j.db.Users[i].Faces) >= j.db.Settings.MaxFacesPerUser {
			return ErrMaxFacesReached
		}

		if face.ID == "" {
			face.ID = uuid.New().String()
		}

		face.EnrolledAt = time.Now()
		j.db.Users[i].Faces = append(j.db.Users[i].Faces, *face)
		j.db.Users[i].UpdatedAt = time.Now()
		return j.saveInternal()
	}

	return ErrUserNotFound
}

// RemoveFace removes a face from a user
func (j *JSONDatabase) RemoveFace(userID, faceID string) error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	for i := range j.db.Users {
		if j.db.Users[i].ID == userID {
			for k := range j.db.Users[i].Faces {
				if j.db.Users[i].Faces[k].ID == faceID {
					j.db.Users[i].Faces = append(
						j.db.Users[i].Faces[:k],
						j.db.Users[i].Faces[k+1:]...,
					)
					j.db.Users[i].UpdatedAt = time.Now()
					return j.saveInternal()
				}
			}
			return fmt.Errorf("face with ID %s not found", faceID)
		}
	}

	return ErrUserNotFound
}

// GetAllEmbeddings returns a map of userID to faces for matching
func (j *JSONDatabase) GetAllEmbeddings() (map[string][]Face, error) {
	j.mutex.RLock()
	defer j.mutex.RUnlock()

	embeddings := make(map[string][]Face)
	for i := range j.db.Users {
		if len(j.db.Users[i].Faces) > 0 {
			embeddings[j.db.Users[i].ID] = j.db.Users[i].Faces
		}
	}

	return embeddings, nil
}

// GetSettings returns the current settings
func (j *JSONDatabase) GetSettings() (*Settings, error) {
	j.mutex.RLock()
	defer j.mutex.RUnlock()

	settings := j.db.Settings
	return &settings, nil
}

// UpdateSettings updates the database settings
func (j *JSONDatabase) UpdateSettings(settings *Settings) error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	j.db.Settings = *settings
	return j.saveInternal()
}

// saveInternal saves without acquiring the lock (must be called with lock held)
func (j *JSONDatabase) saveInternal() error {
	backupPath := j.filePath + ".backup"
	if _, err := os.Stat(j.filePath); err == nil {
		if err := os.Rename(j.filePath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	data, err := json.MarshalIndent(j.db, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal database: %w", err)
	}

	if err := os.WriteFile(j.filePath, data, 0o600); err != nil {
		if _, statErr := os.Stat(backupPath); statErr == nil {
			_ = os.Rename(backupPath, j.filePath)
		}
		return fmt.Errorf("failed to write database: %w", err)
	}

	_ = os.Remove(backupPath)
	return nil
}
