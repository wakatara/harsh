package main

import (
	"log"

	"github.com/wakatara/harsh/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
