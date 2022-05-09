package postgresql

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func NewPostgreSQL() (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dsn := "postgres://temba:temba@localhost/temba?sslmode=disable"
	return pgxpool.Connect(ctx, dsn)
}
