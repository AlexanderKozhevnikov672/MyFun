package main

import (
	"fun/internal"
	"log"
)

func main() {
	if err := internal.Run(); err != nil {
		log.Fatalf("From main: %s\n", err.Error())
	}
}
