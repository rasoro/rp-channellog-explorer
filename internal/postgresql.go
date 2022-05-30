package postgresql

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewPostgreSQL(dbdsn string) (*sqlx.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dsn := "postgres://temba:temba@localhost:5432/temba?sslmode=disable"
	return sqlx.ConnectContext(ctx, "postgres", dsn)
}
