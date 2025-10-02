package handlers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Directory to store uploaded files
const UploadDir = "./uploads"

// UploadFile handles file uploads
func UploadFile() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Parse the uploaded file
        file, err := c.FormFile("file")
        if err != nil {
            return c.Status(400).JSON(fiber.Map{"error": "Failed to parse file: " + err.Error()})
        }

        // Validate file size (e.g., max 10MB)
        if file.Size > 10*1024*1024 {
            return c.Status(400).JSON(fiber.Map{"error": "File size exceeds 10MB limit"})
        }

        // Generate a unique file name
        uniqueFileName := uuid.New().String() + filepath.Ext(file.Filename)

        // Ensure the upload directory exists
        if err := os.MkdirAll(UploadDir, os.ModePerm); err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Failed to create upload directory"})
        }

        // Save the file to the server
        filePath := filepath.Join(UploadDir, uniqueFileName)
        if err := c.SaveFile(file, filePath); err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Failed to save file: " + err.Error()})
        }

        // Return success response
        return c.Status(201).JSON(fiber.Map{
            "message": "File uploaded successfully",
            "file":    uniqueFileName,
        })
    }
}

// DownloadFile handles file downloads
func DownloadFile() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Get the file name from the URL parameter
        fileName := c.Params("fileName")

        // Construct the file path
        filePath := filepath.Join(UploadDir, fileName)

        // Check if the file exists
        if _, err := os.Stat(filePath); os.IsNotExist(err) {
            return c.Status(404).JSON(fiber.Map{"error": "File not found"})
        }

        // Set the appropriate headers for file download
        c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
        c.Set("Content-Type", "application/octet-stream")

        // Send the file
        return c.SendFile(filePath)
    }
}