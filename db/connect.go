package db

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

func Connect() (*pgx.Conn, error) {
    connString := os.Getenv("DATABASE_URL")
    return pgx.Connect(context.Background(), connString)
}

func PingDB(conn *pgx.Conn) error {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    return conn.Ping(ctx)
}