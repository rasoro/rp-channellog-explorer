package main

import (
	"fmt"
	"log"

	"github.com/rasoro/rp-channellog-explorer/ui"
)

func main() {
	if err := execute(); err != nil {
		log.Fatal(err)
	}
}

func execute() error {
	p := ui.NewProgram()
	if err := p.Start(); err != nil {
		return err
	}
	fmt.Println("\nFlw!!!")
	return nil
}
