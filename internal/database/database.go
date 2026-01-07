package database

import "fmt"

// Database defines the interface for all database implementations
type Database interface {
	// User operations
	CreateUser(user *User) error
	GetUser(id string) (*User, error)
	GetUserByName(name string) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id string) error
	ListUsers() ([]User, error)

	// Face operations
	AddFace(userID string, face *Face) error
	RemoveFace(userID, faceID string) error
	GetAllEmbeddings() (map[string][]Face, error)

	// Settings operations
	GetSettings() (*Settings, error)
	UpdateSettings(settings *Settings) error

	// Connection management
	Close() error
}

// DatabaseType represents the type of database backend
type DatabaseType string

const (
	DatabaseTypeSQLite   DatabaseType = "sqlite"
	DatabaseTypePostgres DatabaseType = "postgres"
	DatabaseTypeJSON     DatabaseType = "json"
)

// NewDatabaseConnection creates a new database instance based on the type
func NewDatabaseConnection(dbType DatabaseType, connectionString string) (Database, error) {
	switch dbType {
	case DatabaseTypeSQLite:
		return NewSQLiteDatabase(connectionString)
	case DatabaseTypePostgres:
		return NewPostgresDatabase(connectionString)
	case DatabaseTypeJSON:
		return NewJSONDatabase(connectionString)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

// ParseDatabaseType converts a string to DatabaseType
func ParseDatabaseType(s string) DatabaseType {
	switch s {
	case "sqlite", "sqlite3":
		return DatabaseTypeSQLite
	case "postgres", "postgresql", "pg":
		return DatabaseTypePostgres
	case "json":
		return DatabaseTypeJSON
	default:
		return DatabaseTypeSQLite
	}
}
