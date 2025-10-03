package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pk0205/dropbox-2.0/models"
	"golang.org/x/crypto/bcrypt"
)

// CreateShareLink creates a shareable link for a file or folder
func CreateShareLink(conn *pgx.Conn) fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := struct {
			FileID    string  `json:"fileId"`
			ExpiresIn *int    `json:"expiresIn"` // Hours until expiration (optional)
			Password  *string `json:"password"`  // Optional password protection
		}{}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		userID := c.Locals("userID").(string)

		// Verify file exists and belongs to user
		var fileName string
		var isFolder bool
		err := conn.QueryRow(context.Background(),
			`SELECT original_name, is_folder FROM files WHERE id=$1 AND user_id=$2`,
			req.FileID, userID).Scan(&fileName, &isFolder)

		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "File not found"})
		}

		// Generate random token
		tokenBytes := make([]byte, 32)
		if _, err := rand.Read(tokenBytes); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to generate token"})
		}
		token := hex.EncodeToString(tokenBytes)

		shareID := uuid.New().String()
		var expiresAt *time.Time
		if req.ExpiresIn != nil && *req.ExpiresIn > 0 {
			expiry := time.Now().Add(time.Hour * time.Duration(*req.ExpiresIn))
			expiresAt = &expiry
		}

		// Hash password if provided
		var hashedPassword *string
		if req.Password != nil && *req.Password != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
			}
			hashStr := string(hash)
			hashedPassword = &hashStr
		}

		// Create share link
		_, err = conn.Exec(context.Background(),
			`INSERT INTO share_links (id, file_id, user_id, token, expires_at, password, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			shareID, req.FileID, userID, token, expiresAt, hashedPassword, time.Now())

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create share link"})
		}

		// Mark file as shared
		conn.Exec(context.Background(),
			`UPDATE files SET is_shared = true WHERE id=$1`, req.FileID)

		// Build share URL
		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			baseURL = fmt.Sprintf("http://localhost:%s", os.Getenv("PORT"))
		}
		shareURL := fmt.Sprintf("%s/share/%s", baseURL, token)

		return c.Status(201).JSON(fiber.Map{
			"message":        "Share link created successfully",
			"shareId":        shareID,
			"shareUrl":       shareURL,
			"token":          token,
			"fileName":       fileName,
			"isFolder":       isFolder,
			"expiresAt":      expiresAt,
			"passwordProtected": hashedPassword != nil,
		})
	}
}

// GetSharedFile handles public access to shared files
func GetSharedFile(conn *pgx.Conn) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Params("token")
		password := c.Query("password") // Optional password from query

		// Get share link info
		var shareLink models.ShareLink
		var fileID string
		var filePath string
		var fileName string
		var fileSize int64
		var isFolder bool
		var storedPassword *string

		err := conn.QueryRow(context.Background(),
			`SELECT sl.id, sl.file_id, sl.user_id, sl.expires_at, sl.password,
			f.file_path, f.original_name, f.file_size, f.is_folder
			FROM share_links sl
			JOIN files f ON sl.file_id = f.id
			WHERE sl.token=$1`,
			token).Scan(&shareLink.ID, &fileID, &shareLink.UserID, &shareLink.ExpiresAt,
			&storedPassword, &filePath, &fileName, &fileSize, &isFolder)

		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Share link not found"})
		}

		// Check if expired
		if shareLink.ExpiresAt != nil && shareLink.ExpiresAt.Before(time.Now()) {
			return c.Status(410).JSON(fiber.Map{"error": "Share link has expired"})
		}

		// Check password if required
		if storedPassword != nil {
			if password == "" {
				return c.Status(401).JSON(fiber.Map{
					"error":             "Password required",
					"passwordProtected": true,
				})
			}

			if err := bcrypt.CompareHashAndPassword([]byte(*storedPassword), []byte(password)); err != nil {
				return c.Status(401).JSON(fiber.Map{"error": "Invalid password"})
			}
		}

		// If it's a folder, return folder contents
		if isFolder {
			return getSharedFolderContents(conn, c, fileID)
		}

		// For files, stream the download
		file, err := os.Open(filePath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to open file"})
		}
		defer file.Close()

		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
		c.Set("Content-Type", "application/octet-stream")
		c.Set("Content-Length", fmt.Sprintf("%d", fileSize))

		return c.SendStream(file)
	}
}

// getSharedFolderContents returns the contents of a shared folder
func getSharedFolderContents(conn *pgx.Conn, c *fiber.Ctx, folderID string) error {
	rows, err := conn.Query(context.Background(),
		`SELECT id, file_name, original_name, file_size, is_folder, created_at
		FROM files WHERE parent_id=$1 ORDER BY is_folder DESC, original_name ASC`,
		folderID)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get folder contents"})
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var f models.File
		err := rows.Scan(&f.ID, &f.FileName, &f.OriginalName, &f.FileSize, &f.IsFolder, &f.CreatedAt)
		if err != nil {
			continue
		}
		files = append(files, f)
	}

	return c.Status(200).JSON(fiber.Map{
		"type":     "folder",
		"folderID": folderID,
		"files":    files,
	})
}

// GetShareInfo returns information about a share link without downloading
func GetShareInfo(conn *pgx.Conn) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Params("token")

		var fileName string
		var fileSize int64
		var isFolder bool
		var expiresAt *time.Time
		var hasPassword bool
		var createdAt time.Time

		err := conn.QueryRow(context.Background(),
			`SELECT f.original_name, f.file_size, f.is_folder, sl.expires_at, 
			(sl.password IS NOT NULL) as has_password, sl.created_at
			FROM share_links sl
			JOIN files f ON sl.file_id = f.id
			WHERE sl.token=$1`,
			token).Scan(&fileName, &fileSize, &isFolder, &expiresAt, &hasPassword, &createdAt)

		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Share link not found"})
		}

		// Check if expired
		if expiresAt != nil && expiresAt.Before(time.Now()) {
			return c.Status(410).JSON(fiber.Map{"error": "Share link has expired"})
		}

		return c.Status(200).JSON(fiber.Map{
			"fileName":          fileName,
			"fileSize":          fileSize,
			"isFolder":          isFolder,
			"expiresAt":         expiresAt,
			"passwordProtected": hasPassword,
			"createdAt":         createdAt,
		})
	}
}

// ListUserShares lists all share links created by a user
func ListUserShares(conn *pgx.Conn) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(string)

		rows, err := conn.Query(context.Background(),
			`SELECT sl.id, sl.token, sl.expires_at, sl.created_at,
			f.id, f.original_name, f.is_folder, (sl.password IS NOT NULL) as has_password
			FROM share_links sl
			JOIN files f ON sl.file_id = f.id
			WHERE sl.user_id=$1
			ORDER BY sl.created_at DESC`,
			userID)

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get shares"})
		}
		defer rows.Close()

		type ShareInfo struct {
			ID                string     `json:"id"`
			Token             string     `json:"token"`
			FileID            string     `json:"fileId"`
			FileName          string     `json:"fileName"`
			IsFolder          bool       `json:"isFolder"`
			ExpiresAt         *time.Time `json:"expiresAt"`
			PasswordProtected bool       `json:"passwordProtected"`
			CreatedAt         time.Time  `json:"createdAt"`
			ShareURL          string     `json:"shareUrl"`
		}

		var shares []ShareInfo
		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			baseURL = fmt.Sprintf("http://localhost:%s", os.Getenv("PORT"))
		}

		for rows.Next() {
			var s ShareInfo
			err := rows.Scan(&s.ID, &s.Token, &s.ExpiresAt, &s.CreatedAt,
				&s.FileID, &s.FileName, &s.IsFolder, &s.PasswordProtected)
			if err != nil {
				continue
			}
			s.ShareURL = fmt.Sprintf("%s/share/%s", baseURL, s.Token)
			shares = append(shares, s)
		}

		return c.Status(200).JSON(shares)
	}
}

// DeleteShareLink deletes a share link
func DeleteShareLink(conn *pgx.Conn) fiber.Handler {
	return func(c *fiber.Ctx) error {
		shareID := c.Params("shareId")
		userID := c.Locals("userID").(string)

		// Get file_id before deleting
		var fileID string
		err := conn.QueryRow(context.Background(),
			`SELECT file_id FROM share_links WHERE id=$1 AND user_id=$2`,
			shareID, userID).Scan(&fileID)

		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Share link not found"})
		}

		// Delete share link
		_, err = conn.Exec(context.Background(),
			`DELETE FROM share_links WHERE id=$1 AND user_id=$2`,
			shareID, userID)

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to delete share link"})
		}

		// Check if file has any other shares
		var otherShares int
		conn.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM share_links WHERE file_id=$1`, fileID).Scan(&otherShares)

		// If no other shares, mark file as not shared
		if otherShares == 0 {
			conn.Exec(context.Background(),
				`UPDATE files SET is_shared = false WHERE id=$1`, fileID)
		}

		return c.Status(200).JSON(fiber.Map{"message": "Share link deleted successfully"})
	}
}

// UpdateShareLink updates a share link (extend expiration or change password)
func UpdateShareLink(conn *pgx.Conn) fiber.Handler {
	return func(c *fiber.Ctx) error {
		shareID := c.Params("shareId")
		userID := c.Locals("userID").(string)

		req := struct {
			ExpiresIn *int    `json:"expiresIn"` // Hours to extend
			Password  *string `json:"password"`  // New password (empty string to remove)
		}{}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		// Verify share exists and belongs to user
		var exists bool
		err := conn.QueryRow(context.Background(),
			`SELECT EXISTS(SELECT 1 FROM share_links WHERE id=$1 AND user_id=$2)`,
			shareID, userID).Scan(&exists)

		if err != nil || !exists {
			return c.Status(404).JSON(fiber.Map{"error": "Share link not found"})
		}

		// Update expiration if provided
		if req.ExpiresIn != nil {
			var expiresAt *time.Time
			if *req.ExpiresIn > 0 {
				expiry := time.Now().Add(time.Hour * time.Duration(*req.ExpiresIn))
				expiresAt = &expiry
			}

			_, err = conn.Exec(context.Background(),
				`UPDATE share_links SET expires_at=$1 WHERE id=$2`,
				expiresAt, shareID)

			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to update expiration"})
			}
		}

		// Update password if provided
		if req.Password != nil {
			var hashedPassword *string
			if *req.Password != "" {
				hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
				if err != nil {
					return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
				}
				hashStr := string(hash)
				hashedPassword = &hashStr
			}

			_, err = conn.Exec(context.Background(),
				`UPDATE share_links SET password=$1 WHERE id=$2`,
				hashedPassword, shareID)

			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Failed to update password"})
			}
		}

		return c.Status(200).JSON(fiber.Map{"message": "Share link updated successfully"})
	}
}

