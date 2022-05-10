package main

import (
	"fmt"
	"log"

	_ "github.com/lib/pq"

	postgresql "github.com/rasoro/rp-channellog-explorer/internal"
	"github.com/rasoro/rp-channellog-explorer/internal/db"
	"github.com/rasoro/rp-channellog-explorer/ui"
)

func main() {
	if err := execute(); err != nil {
		log.Fatal(err)
	}
}

func execute() error {
	dbConn, err := postgresql.NewPostgreSQL()
	if err != nil {
		log.Fatal(err)
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
