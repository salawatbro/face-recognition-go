# Face Recognition CLI

A lightweight, pure Go command-line tool for face detection, recognition, and verification. No external dependencies, no cloud services required - everything runs locally.

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)
![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)

## Features

| Feature | Description |
|---------|-------------|
| **Pure Go** | No CGO, no external libraries, single binary |
| **Cross-Platform** | Works on macOS, Linux, Windows (amd64/arm64) |
| **Face Detection** | Pigo-based detection with quality scoring |
| **1:N Identification** | Find a person among all enrolled users |
| **1:1 Verification** | Verify if a photo matches a specific user |
| **Multi-Face Enrollment** | Multiple photos per user for better accuracy |
| **Local Storage** | JSON database + filesystem, no cloud needed |
| **Privacy First** | All processing happens locally |

## Quick Start

```bash
# Clone and build
git clone https://github.com/salawatbro/face-recognition-go.git
cd face-recognition-go
go build -o face

# Enroll a user
./face enroll --name "John Doe" --images "john1.jpg,john2.jpg"

# Identify someone
./face identify --image "unknown.jpg"

# Verify against specific user
./face verify --user-id "abc123" --image "photo.jpg"
```

## Installation

### From Source

```bash
# Requires Go 1.21+
go build -o face
```

### Pre-built Binaries

Download from [Releases](https://github.com/salawatbro/face-recognition-go/releases) page.

## Commands

### `enroll` - Register a New User

```bash
./face enroll --name "John Doe" \
              --email "john@example.com" \
              --phone "+1234567890" \
              --images "photo1.jpg,photo2.jpg,photo3.jpg" \
              --metadata '{"department":"Engineering"}'
```

| Flag | Required | Description |
|------|----------|-------------|
| `--name`, `-n` | Yes | User's full name |
| `--images`, `-i` | Yes | Comma-separated image paths |
| `--email`, `-e` | No | Email address |
| `--phone`, `-p` | No | Phone number |
| `--metadata`, `-m` | No | Custom JSON metadata |

**Output:**
```
User enrolled successfully!
ID:     a1b2c3d4-e5f6-7890-abcd-ef1234567890
Name:   John Doe
Faces:  3 enrolled
```

### `identify` - Find a Person (1:N)

Search all enrolled users to identify someone:

```bash
./face identify --image "unknown.jpg" --threshold 0.5
```

| Flag | Default | Description |
|------|---------|-------------|
| `--image`, `-i` | - | Image to identify (required) |
| `--threshold`, `-t` | 0.4 | Minimum similarity score |

**Output:**
```
Match found!
Name:       John Doe
ID:         a1b2c3d4-e5f6-7890-abcd-ef1234567890
Confidence: 87.32%

Top matches:
  1. John Doe (87.32%)
  2. Jane Smith (42.15%)
  3. Bob Wilson (38.90%)
```

### `verify` - Verify Identity (1:1)

Check if a photo matches a specific user:

```bash
./face verify --user-id "a1b2c3d4" --image "photo.jpg"
```

| Flag | Description |
|------|-------------|
| `--user-id`, `-u` | User ID to verify against (required) |
| `--image`, `-i` | Image to verify (required) |
| `--threshold`, `-t` | Minimum similarity score |

**Output:**
```
VERIFIED - Face matches the user!
Name:       John Doe
Confidence: 89.45%
```

### `list` - Show All Users

```bash
# Table format
./face list

# JSON format
./face list --json
```

**Output:**
```
ID                                    Name          Email              Faces  Created
------------------------------------  ------------  -----------------  -----  -------------------
a1b2c3d4-e5f6-7890-abcd-ef1234567890  John Doe      john@example.com   3      2025-01-07 10:30:00
b2c3d4e5-f6a7-8901-bcde-f12345678901  Jane Smith    jane@example.com   2      2025-01-07 11:45:00
```

### `update` - Modify User

```bash
# Update info
./face update --id "a1b2c3d4" --name "John Smith" --email "new@email.com"

# Add new face
./face update --id "a1b2c3d4" --add-face "newphoto.jpg"

# Remove face
./face update --id "a1b2c3d4" --remove-face "face-uuid"
```

### `delete` - Remove User

```bash
./face delete --id "a1b2c3d4"

# Skip confirmation
./face delete --id "a1b2c3d4" --confirm
```

## Global Flags

| Flag | Environment Variable | Default | Description |
|------|---------------------|---------|-------------|
| `--db` | `FACE_CLI_DB_PATH` | `db.json` | Database file path |
| `--faces-dir` | `FACE_CLI_FACES_DIR` | `faces/` | Face images directory |
| `--verbose`, `-v` | - | false | Enable verbose output |

## How It Works

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI (Cobra)                             │
├─────────────────────────────────────────────────────────────────┤
│  enroll  │  identify  │  verify  │  list  │  update  │  delete  │
├──────────┴────────────┴──────────┴────────┴──────────┴──────────┤
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │  Detector   │  │  Extractor  │  │   Matcher   │             │
│  │   (Pigo)    │→ │ (LBP+HOG)   │→ │  (Cosine)   │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────┐  ┌─────────────────────┐              │
│  │   JSON Database     │  │   Filesystem        │              │
│  │   (db.json)         │  │   (faces/)          │              │
│  └─────────────────────┘  └─────────────────────┘              │
└─────────────────────────────────────────────────────────────────┘
```

### Processing Pipeline

**Enrollment:**
```
Image → Face Detection → Quality Check → Crop → Feature Extraction → Save to DB
```

**Identification:**
```
Image → Face Detection → Crop → Feature Extraction → Compare All → Best Match
```

### Feature Extraction

The system uses classical computer vision techniques:

| Feature | Weight | Description |
|---------|--------|-------------|
| HOG | 40% | Histogram of Oriented Gradients - captures edges |
| LBP | 20% | Local Binary Patterns - captures texture |
| Region | 40% | Eye, nose, mouth specific features |

Output: 128-dimensional L2-normalized embedding vector.

### Matching

Uses cosine similarity:
```
similarity = (A · B) / (||A|| × ||B||)
```

Match threshold: 0.4 (default), adjustable per command.

## Project Structure

```
face/
├── main.go                 # Entry point
├── cmd/                    # CLI commands
│   ├── enroll.go
│   ├── identify.go
│   ├── verify.go
│   ├── list.go
│   ├── update.go
│   └── delete.go
├── internal/
│   ├── database/          # JSON database layer
│   │   ├── models.go      # User, Face structs
│   │   └── json.go        # CRUD operations
│   ├── face/              # Face processing
│   │   ├── detector.go    # Pigo face detection
│   │   ├── embeddings.go  # Feature extraction
│   │   ├── extractor.go   # Interface
│   │   └── matcher.go     # Similarity matching
│   └── storage/           # File storage
│       └── filesystem.go
├── config/
│   └── config.go          # Configuration
├── db.json                # Database (auto-created)
├── faces/                 # Images (auto-created)
└── models/                # Cascade file (auto-downloaded)
```

## Performance

| Metric | Value |
|--------|-------|
| Detection Speed | 30-50ms per image |
| Extraction Speed | 20-30ms per face |
| Matching Speed | <1ms per comparison |
| Memory Usage | ~50MB |
| Database Limit | ~10,000 users |

## Accuracy

| Condition | Accuracy |
|-----------|----------|
| Controlled (good light, frontal) | 80-85% |
| Real-world (varied conditions) | 70-75% |
| Challenging (poor light, angles) | 50-60% |

### Tips for Better Accuracy

1. **Enroll multiple photos** - 3-5 photos per user significantly improves accuracy
2. **Use frontal faces** - Direct front-facing photos work best
3. **Good lighting** - Even, diffused lighting is ideal
4. **High resolution** - At least 200x200 pixels for the face
5. **Adjust threshold** - Lower threshold = more matches (more false positives)

## Use Cases

**Good for:**
- Small-scale access control (<100 users)
- Prototyping and development
- Educational purposes
- Privacy-sensitive applications
- Offline/air-gapped systems

**Not recommended for:**
- High-security authentication
- Large-scale deployments (>10,000 users)
- Real-time video processing
- Applications requiring 99%+ accuracy

## Dependencies

All pure Go, no CGO required:

| Package | Purpose |
|---------|---------|
| `github.com/esimov/pigo` | Face detection |
| `github.com/spf13/cobra` | CLI framework |
| `github.com/google/uuid` | UUID generation |
| `golang.org/x/image` | Image processing |

## Development

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Format code
go fmt ./...

# Lint
golangci-lint run
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing`)
5. Open a Pull Request

## License

Apache License 2.0 - See [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Pigo](https://github.com/esimov/pigo) - Pure Go face detection library
- [Cobra](https://github.com/spf13/cobra) - CLI framework