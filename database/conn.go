package database

import (
	"context"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func Connect(ctx context.Context) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", "user=cifer dbname=leetcodestats sslmode=disable")
	if err != nil {
		return nil, err
	}
	return db, nil
}
