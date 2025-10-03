package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/pk0205/dropbox-2.0/db"
	"github.com/pk0205/dropbox-2.0/handlers"
	"github.com/pk0205/dropbox-2.0/middleware"
)

func main() {
	app := fiber.New(fiber.Config{
		BodyLimit: 100 * 1024 * 1024, // 100MB max body size
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	PORT := os.Getenv("PORT")

	conn, err := db.Connect()
    if err != nil {
        log.Fatal("Error connecting DB:", err)
    }
    defer conn.Close(context.Background())

    if err := db.PingDB(conn); err != nil {
        log.Fatal("Unable to ping database:", err)
    }

    if err := db.SetupDB(conn); err != nil {
        log.Fatal("Unable to setup database:", err)
    }


	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{"msg": "Dropbox 2.0 API Server"})
	})

	// Public routes
	app.Post("/api/user/signup", handlers.SignUp(conn))
	app.Post("/api/user/login", handlers.Login(conn))
	app.Post("/api/user/logout", handlers.Logout())

	// Public share routes (no authentication required)
	app.Get("/share/:token", handlers.GetSharedFile(conn))
	app.Get("/api/share/:token/info", handlers.GetShareInfo(conn))

	// Protected routes - require authentication
	api := app.Group("/api", middleware.RequireAuth(conn))

	// User routes
	api.Get("/users", handlers.GetUsers(conn))

	// File management routes
	api.Get("/files", handlers.ListFiles(conn))
	api.Delete("/files/:fileId", handlers.DeleteFile(conn))
	
	// Basic upload/download (for small files)
	api.Post("/files/upload", handlers.UploadFile())
	api.Get("/files/download/:fileName", handlers.DownloadFile())

	// Advanced file operations
	api.Post("/files/parallel-upload", handlers.ParallelUpload(conn))
	api.Get("/files/stream-download/:fileId", handlers.StreamDownload(conn))

	// Chunked upload for large files
	api.Post("/files/chunk-upload/init", handlers.ChunkedUploadInit(conn))
	api.Post("/files/chunk-upload/:uploadId", handlers.ChunkedUploadChunk(conn))
	api.Post("/files/chunk-upload/:uploadId/complete", handlers.ChunkedUploadComplete(conn))

	// Folder operations
	api.Post("/folders", handlers.CreateFolder(conn))

	// Share management (authenticated)
	api.Post("/shares", handlers.CreateShareLink(conn))
	api.Get("/shares", handlers.ListUserShares(conn))
	api.Delete("/shares/:shareId", handlers.DeleteShareLink(conn))
	api.Put("/shares/:shareId", handlers.UpdateShareLink(conn))


	log.Fatal(app.Listen(":" + PORT))

}