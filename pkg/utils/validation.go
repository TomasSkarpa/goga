package utils

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
)

var (
	allowedTypes = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}
	
	maxFileSize = int64(50 << 20) // 50MB
)

func ValidateImageUpload(file multipart.File, header *multipart.FileHeader) error {
	// Check file size
	if header.Size > maxFileSize {
		return errors.New("file too large")
	}
	
	// Read first 512 bytes for MIME detection
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return err
	}
	file.Seek(0, 0) // Reset file pointer
	
	// Detect actual MIME type
	mimeType := http.DetectContentType(buffer)
	if !allowedTypes[mimeType] {
		return errors.New("invalid file type")
	}
	
	// Check for malicious content
	if bytes.Contains(buffer, []byte("<?php")) ||
	   bytes.Contains(buffer, []byte("<script")) ||
	   bytes.Contains(buffer, []byte("#!/bin/")) {
		return errors.New("malicious content detected")
	}
	
	return nil
}