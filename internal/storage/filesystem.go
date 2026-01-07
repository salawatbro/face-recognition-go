package storage

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

// FileSystemStorage handles file-based image storage
type FileSystemStorage struct {
	baseDir string
}

// NewFileSystemStorage creates a new filesystem storage
func NewFileSystemStorage(baseDir string) (*FileSystemStorage, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &FileSystemStorage{
		baseDir: baseDir,
	}, nil
}

// SaveImage saves an image with a specific filename
func (fs *FileSystemStorage) SaveImage(userID, faceID string, img image.Image) (string, error) {
	filename := fmt.Sprintf("user_%s_face_%s.jpg", userID, faceID)
	fullPath := filepath.Join(fs.baseDir, filename)

	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create image file: %w", err)
	}
	defer file.Close()

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 95}); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	return filename, nil
}

// LoadImage loads an image from a filename
func (fs *FileSystemStorage) LoadImage(filename string) (image.Image, error) {
	fullPath := filepath.Join(fs.baseDir, filename)

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

// LoadImageFromPath loads an image from an absolute or relative path
func (fs *FileSystemStorage) LoadImageFromPath(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(path))
	var img image.Image

	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	default:
		img, _, err = image.Decode(file)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

// DeleteImage removes an image file
func (fs *FileSystemStorage) DeleteImage(filename string) error {
	fullPath := filepath.Join(fs.baseDir, filename)

	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	return nil
}

// ListImages lists all images for a specific user
func (fs *FileSystemStorage) ListImages(userID string) ([]string, error) {
	pattern := filepath.Join(fs.baseDir, fmt.Sprintf("user_%s_face_*.jpg", userID))

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	filenames := make([]string, len(matches))
	for i, match := range matches {
		filenames[i] = filepath.Base(match)
	}

	return filenames, nil
}

// DeleteAllUserImages removes all images for a user
func (fs *FileSystemStorage) DeleteAllUserImages(userID string) error {
	images, err := fs.ListImages(userID)
	if err != nil {
		return err
	}

	for _, filename := range images {
		if err := fs.DeleteImage(filename); err != nil {
			return err
		}
	}

	return nil
}

// Exists checks if an image file exists
func (fs *FileSystemStorage) Exists(filename string) bool {
	fullPath := filepath.Join(fs.baseDir, filename)
	_, err := os.Stat(fullPath)
	return err == nil
}
