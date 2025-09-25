package repository

import (
	"database/sql"
	"goga/internal/models"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ImageRepository struct {
	db *sql.DB
}

func NewImageRepository(db *sql.DB) *ImageRepository {
	return &ImageRepository{db: db}
}

func (r *ImageRepository) Create(image *models.Image) error {
	query := `
		INSERT INTO images (id, filename, original_name, path, size, width, height, format, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(query, image.ID, image.Filename, image.OriginalName, image.Path,
		image.Size, image.Width, image.Height, image.Format, image.CreatedAt, image.UpdatedAt)
	return err
}

func (r *ImageRepository) GetAll() ([]models.Image, error) {
	query := `SELECT id, filename, original_name, path, size, width, height, format, created_at, updated_at FROM images ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		var img models.Image
		err := rows.Scan(&img.ID, &img.Filename, &img.OriginalName, &img.Path,
			&img.Size, &img.Width, &img.Height, &img.Format, &img.CreatedAt, &img.UpdatedAt)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, nil
}

func (r *ImageRepository) GetByID(id string) (*models.Image, error) {
	query := `SELECT id, filename, original_name, path, size, width, height, format, created_at, updated_at FROM images WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var img models.Image
	err := row.Scan(&img.ID, &img.Filename, &img.OriginalName, &img.Path,
		&img.Size, &img.Width, &img.Height, &img.Format, &img.CreatedAt, &img.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &img, nil
}

func (r *ImageRepository) Delete(id string) error {
	query := `DELETE FROM images WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *ImageRepository) InitSchema() error {
	query := `
		CREATE TABLE IF NOT EXISTS images (
			id TEXT PRIMARY KEY,
			filename TEXT NOT NULL,
			original_name TEXT NOT NULL,
			path TEXT NOT NULL,
			size INTEGER NOT NULL,
			width INTEGER NOT NULL,
			height INTEGER NOT NULL,
			format TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`
	_, err := r.db.Exec(query)
	return err
}