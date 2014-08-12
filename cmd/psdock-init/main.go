package main

import (
	"github.com/applidget/psdock"
	"log"
)

func main() {
	if err := psdock.SetNetwork(); err != nil {
		log.Fatal(err)
	}
	psdock.Runner()
}
