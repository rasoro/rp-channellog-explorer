package db

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/stretchr/testify/assert"
)

func TestGetChannel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	conn, err := sqlx.Open("postgres", "postgres://temba:temba@localhost:5432/temba?sslmode=disable")
	assert.NoError(t, err)
	testdb := New(conn)

	ch, err := testdb.GetChannel(ctx, "cac4a1fe-0559-423e-97d6-f4a24f8d98cf")
	assert.NoError(t, err)
	assert.Equal(t, ch.Uuid, "cac4a1fe-0559-423e-97d6-f4a24f8d98cf")
	log.Println(ch)
}
