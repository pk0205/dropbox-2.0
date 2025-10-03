package models

import "time"

type File struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	FileName    string    `json:"fileName"`
	OriginalName string   `json:"originalName"`
	FilePath    string    `json:"filePath"`
	FileSize    int64     `json:"fileSize"`
	MimeType    string    `json:"mimeType"`
	Checksum    string    `json:"checksum"` // SHA-256 hash for deduplication
	ParentID    *string   `json:"parentId"` // For folder structure
	IsFolder    bool      `json:"isFolder"`
	IsShared    bool      `json:"isShared"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ChunkUpload struct {
	ID           string    `json:"id"`
	UserID       string    `json:"userId"`
	FileName     string    `json:"fileName"`
	TotalChunks  int       `json:"totalChunks"`
	ChunkSize    int64     `json:"chunkSize"`
	TotalSize    int64     `json:"totalSize"`
	UploadedChunks []int   `json:"uploadedChunks"`
	Status       string    `json:"status"` // pending, uploading, completed, failed
	CreatedAt    time.Time `json:"createdAt"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type ShareLink struct {
	ID        string    `json:"id"`
	FileID    string    `json:"fileId"`
	UserID    string    `json:"userId"`
	Token     string    `json:"token"`
	ExpiresAt *time.Time `json:"expiresAt"`
	Password  *string   `json:"password"`
	CreatedAt time.Time `json:"createdAt"`
}

type FileVersion struct {
	ID         string    `json:"id"`
	FileID     string    `json:"fileId"`
	VersionNum int       `json:"versionNum"`
	FilePath   string    `json:"filePath"`
	FileSize   int64     `json:"fileSize"`
	Checksum   string    `json:"checksum"`
	CreatedAt  time.Time `json:"createdAt"`
}

