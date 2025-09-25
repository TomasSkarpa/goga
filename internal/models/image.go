package models

import (
	"time"
)

type Image struct {
	ID          string    `json:"id" db:"id"`
	Filename    string    `json:"filename" db:"filename"`
	OriginalName string   `json:"original_name" db:"original_name"`
	Path        string    `json:"path" db:"path"`
	Size        int64     `json:"size" db:"size"`
	Width       int       `json:"width" db:"width"`
	Height      int       `json:"height" db:"height"`
	Format      string    `json:"format" db:"format"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type ImageUploadRequest struct {
	File []byte `json:"file"`
	Name string `json:"name"`
}

type ImageConvertRequest struct {
	Format  string `json:"format"`
	Quality int    `json:"quality,omitempty"`
}