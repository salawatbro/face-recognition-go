package database

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"face/internal/database/models"

	"github.com/google/uuid"
)

// jsonData represents the internal JSON file structure
type jsonData struct {
	Version  string           `json:"version"`
	Users    []models.User    `json:"users"`
	Settings models.Settings  `json:"settings"`
}

// newJSONData creates a new JSON data structure with defaults
func newJSONData() *jsonData {
	return &jsonData{
		Version: "1.0",
		Users:   []models.User{},
		Settings: models.Settings{
			ID:                 1,
			MatchThreshold:     0.6,
			MaxFacesPerUser:    10,
			EmbeddingDimension: 128,
		},
	}
}

// JSONDatabase implements a thread-safe JSON file-based database
type JSONDatabase struct {
	filePath string
	data     *jsonData
	mutex    sync.RWMutex
}

// NewJSONDatabase creates a new JSON database instance
func NewJSONDatabase(filePath string) (*JSONDatabase, error) {
	jdb := &JSONDatabase{
		filePath: filePath,
		data:     newJSONData(),
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

	jd := newJSONData()
	if err := json.Unmarshal(data, jd); err != nil {
		return models.ErrDatabaseCorrupt
	}

	j.data = jd
	return nil
}

// Save writes the database to disk with backup
func (j *JSONDatabase) Save() error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	return j.saveInternal()
}

// CreateUser adds a new user to the database
func (j *JSONDatabase) CreateUser(user *models.User) error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	if err := user.Validate(); err != nil {
		return err
	}

	for i := range j.data.Users {
		if j.data.Users[i].ID == user.ID {
			return models.ErrUserAlreadyExists
		}
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	if user.Faces == nil {
		user.Faces = []models.Face{}
	}

	if user.Metadata == nil {
		user.Metadata = make(models.Metadata)
	}

	j.data.Users = append(j.data.Users, *user)
	return j.saveInternal()
}

// GetUser retrieves a user by ID
func (j *JSONDatabase) GetUser(id string) (*models.User, error) {
	j.mutex.RLock()
	defer j.mutex.RUnlock()

	for i := range j.data.Users {
		if j.data.Users[i].ID == id {
			user := j.data.Users[i]
			return &user, nil
		}
	}

	return nil, models.ErrUserNotFound
}

// GetUserByName retrieves a user by name (case-sensitive)
func (j *JSONDatabase) GetUserByName(name string) (*models.User, error) {
	j.mutex.RLock()
	defer j.mutex.RUnlock()

	for i := range j.data.Users {
		if j.data.Users[i].Name == name {
			user := j.data.Users[i]
			return &user, nil
		}
	}

	return nil, models.ErrUserNotFound
}

// UpdateUser updates an existing user
func (j *JSONDatabase) UpdateUser(user *models.User) error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	if err := user.Validate(); err != nil {
		return err
	}

	for i := range j.data.Users {
		if j.data.Users[i].ID == user.ID {
			user.UpdatedAt = time.Now()
			user.CreatedAt = j.data.Users[i].CreatedAt
			j.data.Users[i] = *user
			return j.saveInternal()
		}
	}

	return models.ErrUserNotFound
}

// DeleteUser removes a user from the database
func (j *JSONDatabase) DeleteUser(id string) error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	for i := range j.data.Users {
		if j.data.Users[i].ID == id {
			j.data.Users = append(j.data.Users[:i], j.data.Users[i+1:]...)
			return j.saveInternal()
		}
	}

	return models.ErrUserNotFound
}

// ListUsers returns all users in the database
func (j *JSONDatabase) ListUsers() ([]models.User, error) {
	j.mutex.RLock()
	defer j.mutex.RUnlock()

	users := make([]models.User, len(j.data.Users))
	copy(users, j.data.Users)
	return users, nil
}

// AddFace adds a face to a user
func (j *JSONDatabase) AddFace(userID string, face *models.Face) error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	if err := face.Validate(); err != nil {
		return err
	}

	for i := range j.data.Users {
		if j.data.Users[i].ID != userID {
			continue
		}
		if len(j.data.Users[i].Faces) >= j.data.Settings.MaxFacesPerUser {
			return models.ErrMaxFacesReached
		}

		if face.ID == "" {
			face.ID = uuid.New().String()
		}

		face.EnrolledAt = time.Now()
		j.data.Users[i].Faces = append(j.data.Users[i].Faces, *face)
		j.data.Users[i].UpdatedAt = time.Now()
		return j.saveInternal()
	}

	return models.ErrUserNotFound
}

// RemoveFace removes a face from a user
func (j *JSONDatabase) RemoveFace(userID, faceID string) error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	for i := range j.data.Users {
		if j.data.Users[i].ID == userID {
			for k := range j.data.Users[i].Faces {
				if j.data.Users[i].Faces[k].ID == faceID {
					j.data.Users[i].Faces = append(
						j.data.Users[i].Faces[:k],
						j.data.Users[i].Faces[k+1:]...,
					)
					j.data.Users[i].UpdatedAt = time.Now()
					return j.saveInternal()
				}
			}
			return fmt.Errorf("face with ID %s not found", faceID)
		}
	}

	return models.ErrUserNotFound
}

// GetAllEmbeddings returns a map of userID to faces for matching
func (j *JSONDatabase) GetAllEmbeddings() (map[string][]models.Face, error) {
	j.mutex.RLock()
	defer j.mutex.RUnlock()

	embeddings := make(map[string][]models.Face)
	for i := range j.data.Users {
		if len(j.data.Users[i].Faces) > 0 {
			embeddings[j.data.Users[i].ID] = j.data.Users[i].Faces
		}
	}

	return embeddings, nil
}

// GetSettings returns the current settings
func (j *JSONDatabase) GetSettings() (*models.Settings, error) {
	j.mutex.RLock()
	defer j.mutex.RUnlock()

	settings := j.data.Settings
	return &settings, nil
}

// UpdateSettings updates the database settings
func (j *JSONDatabase) UpdateSettings(settings *models.Settings) error {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	j.data.Settings = *settings
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

	data, err := json.MarshalIndent(j.data, "", "  ")
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

// Close implements the Database interface (no-op for JSON)
func (j *JSONDatabase) Close() error {
	return j.Save()
}
