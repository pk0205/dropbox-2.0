package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func SetupDB(conn *pgx.Conn) error {
	// Create users table
	_, err := conn.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS users (
            id TEXT PRIMARY KEY,
            firstName TEXT,
            lastName TEXT,
            username TEXT NOT NULL UNIQUE,
            email TEXT NOT NULL UNIQUE,
            password TEXT NOT NULL
        );
    `)
	if err != nil {
		return err
	}

	// Create files table with folder support and deduplication
	_, err = conn.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS files (
            id TEXT PRIMARY KEY,
            user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            file_name TEXT NOT NULL,
            original_name TEXT NOT NULL,
            file_path TEXT,
            file_size BIGINT DEFAULT 0,
            mime_type TEXT,
            checksum TEXT,
            parent_id TEXT REFERENCES files(id) ON DELETE CASCADE,
            is_folder BOOLEAN DEFAULT FALSE,
            is_shared BOOLEAN DEFAULT FALSE,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP NOT NULL DEFAULT NOW()
        );

        CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id);
        CREATE INDEX IF NOT EXISTS idx_files_parent_id ON files(parent_id);
        CREATE INDEX IF NOT EXISTS idx_files_checksum ON files(checksum);
    `)
	if err != nil {
		return err
	}

	// Create chunk_uploads table for resumable uploads
	_, err = conn.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS chunk_uploads (
            id TEXT PRIMARY KEY,
            user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            file_name TEXT NOT NULL,
            total_chunks INTEGER NOT NULL,
            chunk_size BIGINT NOT NULL,
            total_size BIGINT NOT NULL,
            uploaded_chunks INTEGER[] DEFAULT ARRAY[]::INTEGER[],
            status TEXT NOT NULL DEFAULT 'pending',
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            expires_at TIMESTAMP NOT NULL
        );

        CREATE INDEX IF NOT EXISTS idx_chunk_uploads_user_id ON chunk_uploads(user_id);
        CREATE INDEX IF NOT EXISTS idx_chunk_uploads_status ON chunk_uploads(status);
    `)
	if err != nil {
		return err
	}

	// Create share_links table for file sharing
	_, err = conn.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS share_links (
            id TEXT PRIMARY KEY,
            file_id TEXT NOT NULL REFERENCES files(id) ON DELETE CASCADE,
            user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            token TEXT NOT NULL UNIQUE,
            expires_at TIMESTAMP,
            password TEXT,
            created_at TIMESTAMP NOT NULL DEFAULT NOW()
        );

        CREATE INDEX IF NOT EXISTS idx_share_links_token ON share_links(token);
        CREATE INDEX IF NOT EXISTS idx_share_links_file_id ON share_links(file_id);
    `)
	if err != nil {
		return err
	}

	// Create file_versions table for version control
	_, err = conn.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS file_versions (
            id TEXT PRIMARY KEY,
            file_id TEXT NOT NULL REFERENCES files(id) ON DELETE CASCADE,
            version_num INTEGER NOT NULL,
            file_path TEXT NOT NULL,
            file_size BIGINT NOT NULL,
            checksum TEXT NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT NOW(),
            UNIQUE(file_id, version_num)
        );

        CREATE INDEX IF NOT EXISTS idx_file_versions_file_id ON file_versions(file_id);
    `)
	if err != nil {
		return err
	}

	return nil
}