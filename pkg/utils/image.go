package utils

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

func ProcessImage(inputPath, outputPath string, format string, quality int) error {
	src, err := imaging.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}

	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		err = imaging.Save(src, outputPath, imaging.JPEGQuality(quality))
	case "png":
		err = imaging.Save(src, outputPath, imaging.PNGCompressionLevel(6))
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	return nil
}

func GetImageDimensions(imagePath string) (int, int, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	return img.Width, img.Height, nil
}

func GetImageFormat(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "jpeg"
	case ".png":
		return "png"
	case ".webp":
		return "webp"
	default:
		return "unknown"
	}
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}