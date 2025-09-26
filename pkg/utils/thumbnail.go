package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func ClearThumbnailCache(uploadDir, imageID string) {
	thumbDir := filepath.Join(uploadDir, "thumbs")
	if _, err := os.Stat(thumbDir); err != nil {
		return
	}
	
	filepath.Walk(thumbDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		filename := filepath.Base(path)
		if len(filename) >= len(imageID) && filename[:len(imageID)] == imageID {
			os.Remove(path)
			fmt.Printf("Cleared thumbnail: %s\n", path)
		}
		return nil
	})
}