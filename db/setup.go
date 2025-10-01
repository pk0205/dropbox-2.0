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

	// Insert initial user if not exists
	_, err = conn.Exec(context.Background(), `
        INSERT INTO users (id, firstName, lastName, username, email, password) VALUES
        ('1', 'Alice', 'Noone', 'alice', 'alice@example.com', 'password123')
        ON CONFLICT (id) DO NOTHING;
    `)
	if err != nil {
		return err
	}

	return err
}