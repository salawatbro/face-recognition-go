package cmd

import (
	"fmt"
	"image"

	"face/config"
	"face/internal/database"
	"face/internal/face"
	"face/internal/storage"
)

type FaceSystem struct {
	DB        *database.JSONDatabase
	Storage   *storage.FileSystemStorage
	Detector  *face.Detector
	Extractor face.Extractor
}

func NewFaceSystem(cfg *config.Config) (*FaceSystem, error) {
	db, err := database.NewJSONDatabase(cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	stor, err := storage.NewFileSystemStorage(cfg.FacesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	detector, err := face.NewDetector(cfg.ModelsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize detector: %w", err)
	}

	extractor, err := face.NewExtractor(cfg.ModelsDir)
	if err != nil {
		detector.Close()
		return nil, fmt.Errorf("failed to initialize extractor: %w", err)
	}

	return &FaceSystem{
		DB:        db,
		Storage:   stor,
		Detector:  detector,
		Extractor: extractor,
	}, nil
}

func (fs *FaceSystem) Close() {
	if fs.Detector != nil {
		fs.Detector.Close()
	}
	if fs.Extractor != nil {
		fs.Extractor.Close()
	}
}

type FaceResult struct {
	Image        image.Image
	CroppedFace  image.Image
	Embedding    []float32
	QualityScore float64
}

func (fs *FaceSystem) ProcessImage(imagePath string) (*FaceResult, error) {
	img, err := fs.Storage.LoadImageFromPath(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	faceRect, err := fs.Detector.DetectLargestFace(img)
	if err != nil {
		return nil, fmt.Errorf("no face detected in image")
	}

	croppedFace := fs.Detector.CropFace(img, faceRect)
	qualityScore := fs.Detector.CalculateQuality(img, faceRect)

	embedding, err := fs.Extractor.Extract(croppedFace)
	if err != nil {
		return nil, fmt.Errorf("failed to extract embedding: %w", err)
	}

	return &FaceResult{
		Image:        img,
		CroppedFace:  croppedFace,
		Embedding:    embedding,
		QualityScore: qualityScore,
	}, nil
}
