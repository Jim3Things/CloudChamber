package main

import (
	"log"

	"github.com/Jim3Things/CloudChamber/internal/services/frontend"
)

func main() {

	if err := frontend.StartService(); err != nil {
		log.Fatalf("Error running service: %v", err)
	}
}
