# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a face recognition CLI tool written in Go. The system can enroll users with face images and later identify them from photos. It uses Pigo for face detection (pure Go, no CGO required) and stores data in a JSON database with images on the filesystem.

## Common Commands

### Building
```bash
go build -o face
```

### Running
```bash
# Enroll a user
./face enroll --name "John Doe" --email "john@example.com" --images "photo1.jpg,photo2.jpg"

# Identify from image
./face identify --image "unknown.jpg"

# Verify against specific user
./face verify --user-id "uuid" --image "photo.jpg"

# List users
./face list

# Update user
./face update --id "uuid" --email "new@example.com"

# Delete user
./face delete --id "uuid"
```

### Testing
```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/face -v

# Run with coverage
go test -cover ./...
```

### Development
```bash
# Format code
go fmt ./...

# Install dependencies
go mod tidy
```

## Architecture

### Project Structure

```
face/
├── main.go              # CLI entry point with Cobra
├── cmd/                 # CLI command implementations
│   ├── enroll.go       # User enrollment with face images
│   ├── identify.go     # Person identification from image
│   ├── verify.go       # 1:1 verification against specific user
│   ├── list.go         # List all users
│   ├── delete.go       # Delete user
│   └── update.go       # Update user info or faces
├── internal/
│   ├── database/       # Database layer
│   │   ├── models.go   # Core data structures (User, Face, Database)
│   │   └── json.go     # Thread-safe JSON file operations
│   ├── face/           # Face processing (pure Go)
│   │   ├── detector.go # Pigo-based face detection
│   │   ├── embeddings.go # LBP+HOG+Color feature extraction
│   │   ├── extractor.go # Extractor interface
│   │   └── matcher.go  # Cosine similarity matching
│   └── storage/        # File storage
│       └── filesystem.go # Image save/load operations
└── config/
    └── config.go       # Configuration management
```

### Key Components

1. **Database Layer** (`internal/database/`)
   - Uses JSON file (`db.json`) for persistence
   - Thread-safe with mutex for concurrent access
   - Automatic backup before writes
   - Supports CRUD operations for users and faces

2. **Face Detection** (`internal/face/detector.go`)
   - Uses Pigo library (pure Go, no CGO)
   - Auto-downloads cascade file on first run
   - Detects faces and calculates quality scores
   - Crops faces with padding for embedding extraction

3. **Embedding Extraction** (`internal/face/embeddings.go`)
   - Pure Go implementation using classical computer vision features
   - LBP (Local Binary Patterns) for texture
   - HOG (Histogram of Oriented Gradients) for edges
   - Color histograms for color information
   - Grid-based intensity features for spatial information
   - Generates 128-dimensional L2-normalized vectors

4. **Face Matching** (`internal/face/matcher.go`)
   - Uses cosine similarity for comparison
   - Configurable threshold (default: 0.75)
   - Returns best match or no match

5. **Storage** (`internal/storage/filesystem.go`)
   - Saves images as `user_{uuid}_face_{uuid}.jpg`
   - Supports JPEG and PNG formats
   - Images stored in `faces/` directory

6. **CLI** (`main.go`, `cmd/`)
   - Built with Cobra framework
   - Each command in separate file
   - Global flags: --db, --faces-dir, --verbose

### Data Flow

**Enrollment:**
```
Image → Detector → Face Detection → Crop → Extractor → Embedding → Database + Storage
```

**Identification:**
```
Image → Detector → Face Detection → Crop → Extractor → Embedding → Matcher → Result
```

### Database Schema

```go
type User struct {
    ID        string                 // UUID
    Name      string                 // Required
    Email     string                 // Optional
    Phone     string                 // Optional
    Metadata  map[string]interface{} // Custom fields
    Faces     []Face                 // Multiple faces per user
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Face struct {
    ID           string    // UUID
    Filename     string    // user_uuid_face_uuid.jpg
    Embedding    []float32 // 128-d vector
    EnrolledAt   time.Time
    QualityScore float64   // 0.0-1.0
}
```

## Important Implementation Notes

### Face Detection
- Pigo works best with frontal faces
- Quality score based on face size, brightness, and sharpness
- Minimum quality threshold: 0.3 for enrollment

### Embedding Extraction
- Pure Go implementation (no external models required)
- Uses LBP, HOG, color histograms, and grid features
- Embeddings are L2-normalized before storage
- Suitable for small-scale deployments

### Face Matching
- Cosine similarity threshold: 0.75 (adjustable)
- Compares query embedding against all stored faces
- Returns user with highest similarity above threshold
- Multiple faces per user improve accuracy

### Error Handling
- Custom errors defined in `internal/database/models.go`
- Commands should return errors, not call `os.Exit()`
- User-friendly error messages in CLI layer

### Thread Safety
- Database operations use mutex (`internal/database/json.go`)
- Backup created before each write
- `.backup` file used for rollback on error

## Development Guidelines

### Adding New Commands
1. Create file in `cmd/` directory
2. Implement `NewXxxCmd(cfg *config.Config) *cobra.Command`
3. Register in `main.go` init function
4. Follow existing command patterns

### Modifying Database Schema
1. Update structs in `internal/database/models.go`
2. Update JSON marshal/unmarshal if needed
3. Consider backward compatibility
4. Update validation functions

### Adding Tests
- Use table-driven tests
- Mock database for unit tests
- Use sample images for integration tests
- Test error cases thoroughly

## Dependencies

Key external packages (all pure Go, no CGO):
- `github.com/esimov/pigo/core` - Face detection
- `github.com/spf13/cobra` - CLI framework
- `github.com/google/uuid` - UUID generation
- `golang.org/x/image/draw` - Image processing

## Configuration

```go
DatabasePath:     "db.json"
FacesDir:         "faces"
ModelsDir:        "models"
DefaultThreshold: 0.75
```

Environment variables:
- FACE_CLI_DB_PATH
- FACE_CLI_FACES_DIR
- FACE_CLI_MODEL_DIR

## Troubleshooting

### Build Errors
- Ensure Go 1.21+ is installed
- Run `go mod tidy` to sync dependencies

### Runtime Errors
- "No face detected": Image quality too low or no frontal face
- "User not found": Check UUID is correct with `face list`
- "Database corrupted": Restore from `.backup` file

### Accuracy Notes
- Best results with clear frontal face photos
- Good lighting improves accuracy
- Multiple enrollment photos per user recommended