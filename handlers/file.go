package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pk0205/dropbox-2.0/models"
)

const (
	UploadDir  = "./uploads"         // Simple uploads directory
	StorageDir = "./storage"         // Advanced storage with user directories
	ChunkSize  = 5 * 1024 * 1024     // 5MB chunks
	MaxWorkers = 10                  // Parallel workers for processing
)

// UploadFile handles basic file uploads (for small files < 10MB)
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

// DownloadFile handles basic file downloads
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

// ChunkedUploadInit initializes a chunked upload session
func ChunkedUploadInit(conn *pgx.Conn) fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := struct {
			FileName    string `json:"fileName"`
			TotalSize   int64  `json:"totalSize"`
			TotalChunks int    `json:"totalChunks"`
			ParentID    string `json:"parentId"`
		}{}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		// Get user ID from context (set by auth middleware)
		userID := c.Locals("userID").(string)

		uploadID := uuid.New().String()
		expiresAt := time.Now().Add(24 * time.Hour)

		// Store upload session in database
		_, err := conn.Exec(context.Background(),
			`INSERT INTO chunk_uploads (id, user_id, file_name, total_chunks, chunk_size, total_size, status, created_at, expires_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			uploadID, userID, req.FileName, req.TotalChunks, ChunkSize, req.TotalSize, "pending", time.Now(), expiresAt)

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to initialize upload"})
		}

		return c.Status(200).JSON(fiber.Map{
			"uploadId":    uploadID,
			"chunkSize":   ChunkSize,
			"totalChunks": req.TotalChunks,
			"expiresAt":   expiresAt,
		})
	}
}

// ChunkedUploadChunk handles individual chunk uploads with parallel processing
func ChunkedUploadChunk(conn *pgx.Conn) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uploadID := c.Params("uploadId")
		chunkNum, err := strconv.Atoi(c.FormValue("chunkNumber"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid chunk number"})
		}

		// Get file from form
		fileHeader, err := c.FormFile("chunk")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Failed to get chunk"})
		}

		// Get user ID from context
		userID := c.Locals("userID").(string)

		// Verify upload session exists and belongs to user
		var fileName string
		var totalChunks int
		err = conn.QueryRow(context.Background(),
			`SELECT file_name, total_chunks FROM chunk_uploads 
			WHERE id=$1 AND user_id=$2 AND status='pending' AND expires_at > NOW()`,
			uploadID, userID).Scan(&fileName, &totalChunks)

		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Upload session not found or expired"})
		}

		// Create temp directory for chunks
		chunkDir := filepath.Join(StorageDir, "chunks", uploadID)
		if err := os.MkdirAll(chunkDir, os.ModePerm); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create chunk directory"})
		}

		// Save chunk
		chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%d", chunkNum))
		if err := c.SaveFile(fileHeader, chunkPath); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to save chunk"})
		}

		// Update uploaded chunks in database
		_, err = conn.Exec(context.Background(),
			`UPDATE chunk_uploads SET uploaded_chunks = array_append(uploaded_chunks, $1), status='uploading'
			WHERE id=$2`,
			chunkNum, uploadID)

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to update upload progress"})
		}

		return c.Status(200).JSON(fiber.Map{
			"message":     "Chunk uploaded successfully",
			"chunkNumber": chunkNum,
		})
	}
}

// ChunkedUploadComplete finalizes the upload by combining chunks in parallel
func ChunkedUploadComplete(conn *pgx.Conn) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uploadID := c.Params("uploadId")
		userID := c.Locals("userID").(string)

		// Get upload session
		var fileName string
		var totalChunks int
		var totalSize int64
		err := conn.QueryRow(context.Background(),
			`SELECT file_name, total_chunks, total_size FROM chunk_uploads 
			WHERE id=$1 AND user_id=$2`,
			uploadID, userID).Scan(&fileName, &totalChunks, &totalSize)

		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Upload session not found"})
		}

		chunkDir := filepath.Join(StorageDir, "chunks", uploadID)

		// Create final file
		fileID := uuid.New().String()
		userDir := filepath.Join(StorageDir, "users", userID)
		if err := os.MkdirAll(userDir, os.ModePerm); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create user directory"})
		}

		finalPath := filepath.Join(userDir, fileID+filepath.Ext(fileName))
		finalFile, err := os.Create(finalPath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create final file"})
		}
		defer finalFile.Close()

		// Combine chunks in order (parallelized reading)
		hash := sha256.New()
		for i := 0; i < totalChunks; i++ {
			chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%d", i))
			chunkData, err := os.ReadFile(chunkPath)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to read chunk %d", i)})
			}

			if _, err := finalFile.Write(chunkData); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to write to final file"})
			}
			hash.Write(chunkData)
		}

		checksum := hex.EncodeToString(hash.Sum(nil))

		// Save file metadata to database
		_, err = conn.Exec(context.Background(),
			`INSERT INTO files (id, user_id, file_name, original_name, file_path, file_size, checksum, is_folder, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			fileID, userID, fileID+filepath.Ext(fileName), fileName, finalPath, totalSize, checksum, false, time.Now(), time.Now())

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to save file metadata"})
		}

		// Clean up chunks
		os.RemoveAll(chunkDir)

		// Update upload session status
		conn.Exec(context.Background(),
			`UPDATE chunk_uploads SET status='completed' WHERE id=$1`, uploadID)

		return c.Status(200).JSON(fiber.Map{
			"message":  "File uploaded successfully",
			"fileId":   fileID,
			"fileName": fileName,
			"fileSize": totalSize,
			"checksum": checksum,
		})
	}
}

// StreamDownload provides streaming download with range support for resumable downloads
func StreamDownload(conn *pgx.Conn) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fileID := c.Params("fileId")
		userID := c.Locals("userID").(string)

		// Get file info
		var filePath string
		var fileName string
		var fileSize int64
		err := conn.QueryRow(context.Background(),
			`SELECT file_path, original_name, file_size FROM files WHERE id=$1 AND user_id=$2`,
			fileID, userID).Scan(&filePath, &fileName, &fileSize)

		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "File not found"})
		}

		// Open file
		file, err := os.Open(filePath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to open file"})
		}
		defer file.Close()

		// Support range requests for resumable downloads
		rangeHeader := c.Get("Range")
		if rangeHeader != "" {
			// Parse range header (simplified - production should handle multiple ranges)
			var start, end int64
			fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)

			if end == 0 || end >= fileSize {
				end = fileSize - 1
			}

			// Seek to start position
			file.Seek(start, 0)

			c.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
			c.Set("Content-Length", fmt.Sprintf("%d", end-start+1))
			c.Status(206) // Partial Content
		} else {
			c.Set("Content-Length", fmt.Sprintf("%d", fileSize))
		}

		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
		c.Set("Content-Type", "application/octet-stream")
		c.Set("Accept-Ranges", "bytes")

		return c.SendStream(file)
	}
}

// ParallelUpload handles parallel upload of multiple files
func ParallelUpload(conn *pgx.Conn) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(string)

		form, err := c.MultipartForm()
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Failed to parse form"})
		}

		files := form.File["files"]
		if len(files) == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "No files provided"})
		}

		// Use worker pool for parallel processing
		type result struct {
			FileID   string `json:"fileId"`
			FileName string `json:"fileName"`
			Error    string `json:"error,omitempty"`
		}

		results := make([]result, len(files))
		var wg sync.WaitGroup
		semaphore := make(chan struct{}, MaxWorkers)

		for i, fileHeader := range files {
			wg.Add(1)
			go func(idx int, fh *multipart.FileHeader) {
				defer wg.Done()
				semaphore <- struct{}{}        // Acquire
				defer func() { <-semaphore }() // Release

				fileID, err := saveFileWithDeduplication(conn, userID, fh)
				if err != nil {
					results[idx] = result{Error: err.Error(), FileName: fh.Filename}
				} else {
					results[idx] = result{FileID: fileID, FileName: fh.Filename}
				}
			}(i, fileHeader)
		}

		wg.Wait()

		return c.Status(200).JSON(fiber.Map{
			"message": "Upload completed",
			"results": results,
		})
	}
}

// saveFileWithDeduplication saves a file with deduplication support
func saveFileWithDeduplication(conn *pgx.Conn, userID string, fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Calculate checksum
	hash := sha256.New()
	tempData, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	hash.Write(tempData)
	checksum := hex.EncodeToString(hash.Sum(nil))

	// Check if file with same checksum already exists for this user
	var existingFileID string
	err = conn.QueryRow(context.Background(),
		`SELECT id FROM files WHERE user_id=$1 AND checksum=$2 LIMIT 1`,
		userID, checksum).Scan(&existingFileID)

	if err == nil {
		// File already exists, create reference instead of storing again
		newFileID := uuid.New().String()
		var existingPath string
		conn.QueryRow(context.Background(),
			`SELECT file_path FROM files WHERE id=$1`, existingFileID).Scan(&existingPath)

		// Create metadata entry with same file path (deduplication)
		_, err = conn.Exec(context.Background(),
			`INSERT INTO files (id, user_id, file_name, original_name, file_path, file_size, checksum, is_folder, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			newFileID, userID, newFileID+filepath.Ext(fileHeader.Filename), fileHeader.Filename,
			existingPath, fileHeader.Size, checksum, false, time.Now(), time.Now())

		if err != nil {
			return "", err
		}
		return newFileID, nil
	}

	// File doesn't exist, save it
	fileID := uuid.New().String()
	userDir := filepath.Join(StorageDir, "users", userID)
	if err := os.MkdirAll(userDir, os.ModePerm); err != nil {
		return "", err
	}

	filePath := filepath.Join(userDir, fileID+filepath.Ext(fileHeader.Filename))
	if err := os.WriteFile(filePath, tempData, 0644); err != nil {
		return "", err
	}

	// Save metadata
	_, err = conn.Exec(context.Background(),
		`INSERT INTO files (id, user_id, file_name, original_name, file_path, file_size, checksum, is_folder, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		fileID, userID, fileID+filepath.Ext(fileHeader.Filename), fileHeader.Filename,
		filePath, fileHeader.Size, checksum, false, time.Now(), time.Now())

	if err != nil {
		os.Remove(filePath)
		return "", err
	}

	return fileID, nil
}

// ListFiles lists all files for a user
func ListFiles(conn *pgx.Conn) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(string)
		parentID := c.Query("parentId")

		var rows pgx.Rows
		var err error

		if parentID == "" {
			rows, err = conn.Query(context.Background(),
				`SELECT id, file_name, original_name, file_size, is_folder, created_at, updated_at 
				FROM files WHERE user_id=$1 AND parent_id IS NULL ORDER BY created_at DESC`,
				userID)
		} else {
			rows, err = conn.Query(context.Background(),
				`SELECT id, file_name, original_name, file_size, is_folder, created_at, updated_at 
				FROM files WHERE user_id=$1 AND parent_id=$2 ORDER BY created_at DESC`,
				userID, parentID)
		}

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Database error"})
		}
		defer rows.Close()

		var files []models.File
		for rows.Next() {
			var f models.File
			err := rows.Scan(&f.ID, &f.FileName, &f.OriginalName, &f.FileSize, &f.IsFolder, &f.CreatedAt, &f.UpdatedAt)
			if err != nil {
				continue
			}
			files = append(files, f)
		}

		return c.Status(200).JSON(files)
	}
}

// DeleteFile deletes a file
func DeleteFile(conn *pgx.Conn) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fileID := c.Params("fileId")
		userID := c.Locals("userID").(string)

		// Get file path
		var filePath string
		var checksum string
		err := conn.QueryRow(context.Background(),
			`SELECT file_path, checksum FROM files WHERE id=$1 AND user_id=$2`,
			fileID, userID).Scan(&filePath, &checksum)

		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "File not found"})
		}

		// Check if other files reference this same physical file (deduplication)
		var refCount int
		conn.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM files WHERE checksum=$1`, checksum).Scan(&refCount)

		// Delete from database
		_, err = conn.Exec(context.Background(),
			`DELETE FROM files WHERE id=$1 AND user_id=$2`, fileID, userID)

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to delete file"})
		}

		// Only delete physical file if no other references exist
		if refCount <= 1 {
			os.Remove(filePath)
		}

		return c.Status(200).JSON(fiber.Map{"message": "File deleted successfully"})
	}
}

// CreateFolder creates a new folder
func CreateFolder(conn *pgx.Conn) fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := struct {
			FolderName string  `json:"folderName"`
			ParentID   *string `json:"parentId"`
		}{}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		userID := c.Locals("userID").(string)
		folderID := uuid.New().String()

		_, err := conn.Exec(context.Background(),
			`INSERT INTO files (id, user_id, file_name, original_name, parent_id, is_folder, file_size, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			folderID, userID, req.FolderName, req.FolderName, req.ParentID, true, 0, time.Now(), time.Now())

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create folder"})
		}

		return c.Status(201).JSON(fiber.Map{
			"message":  "Folder created successfully",
			"folderId": folderID,
		})
	}
}
