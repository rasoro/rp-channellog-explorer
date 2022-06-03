package main

import (
	"flag"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	postgresql "github.com/rasoro/rp-channellog-explorer/internal"
	"github.com/rasoro/rp-channellog-explorer/internal/db"
	"github.com/rasoro/rp-channellog-explorer/ui"
)

var defaultdbdsn = "postgres://temba:temba@localhost:5432/temba?sslmode=disable"
var dbdsnHelp = "The dsn to establish connection with the postgres database."
var dbdsn *string

func main() {
	dbdsn = flag.String("db", defaultdbdsn, dbdsnHelp)

	flag.Parse()

	if err := execute(); err != nil {
		log.Fatal(err)
	}
}

func execute() error {
	dbConn, err := postgresql.NewPostgreSQL(*dbdsn)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to connect to database: %s", err))
		return err
	}

	dbase := db.New(dbConn)
	p := ui.NewProgram(dbase)
	if err := p.Start(); err != nil {
		return err
	}
	fmt.Println("\nFlw!!!")
	return nil
}
