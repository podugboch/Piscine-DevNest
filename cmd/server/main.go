package main

import (
	"log"
	"piscine-devnest/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
