package server

import (
	"database/sql"
	"goga/internal/handlers"
	"goga/internal/repository"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router        *gin.Engine
	db            *sql.DB
	uploadDir     string
	configHandler *handlers.ConfigHandler
}

func New(dbPath, uploadDir string) (*Server, error) {
	// Initialize database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Initialize repository
	imageRepo := repository.NewImageRepository(db)
	if err := imageRepo.InitSchema(); err != nil {
		return nil, err
	}

	// Create upload directory
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, err
	}

	// Initialize handlers
	configHandler := handlers.NewConfigHandler()
	configHandler.LoadConfig()
	imageHandler := handlers.NewImageHandler(imageRepo, uploadDir)
	webHandler := handlers.NewWebHandler()

	// Setup router
	router := gin.Default()
	router.LoadHTMLGlob("web/templates/*.html")
	router.Static("/static", "./web/static")

	// Web routes
	router.GET("/", webHandler.Dashboard)
	router.GET("/image/:id", webHandler.ImageDetail)

	// API routes
	api := router.Group("/api")
	{
		api.GET("/images", imageHandler.GetImages)
		api.GET("/images/:id", imageHandler.GetImage)
		api.POST("/images/upload", imageHandler.UploadImage)
		api.POST("/images/:id/convert", imageHandler.ConvertImage)
		api.DELETE("/images/:id", imageHandler.DeleteImage)
		api.GET("/images/:id/file", imageHandler.ServeImage)
		api.GET("/config", configHandler.GetConfig)
		api.POST("/config", configHandler.UpdateConfig)
	}

	return &Server{
		router:        router,
		db:            db,
		uploadDir:     uploadDir,
		configHandler: configHandler,
	}, nil
}

func (s *Server) Start(port string) error {
	log.Printf("Starting server on port %s", port)
	log.Printf("Upload directory: %s", s.uploadDir)
	return s.router.Run(":" + port)
}

func (s *Server) Close() error {
	return s.db.Close()
}