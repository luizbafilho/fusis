package main

import (
	"log"
	"net/http"

	"github.com/google/seesaw/ipvs"
)

func main() {
	if err := ipvs.Init(); err != nil {
		log.Fatalf("IPVS initialisation failed: %v\n", err)
	}
	log.Printf("IPVS version %s\n", ipvs.Version())

	router := NewRouter()

	log.Fatal(http.ListenAndServe(":8000", router))
}
