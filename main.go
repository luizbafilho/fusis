package main

import (
	"log"
	"os"

	"github.com/google/seesaw/ipvs"
	"github.com/luizbafilho/fusis/api"
	"github.com/luizbafilho/fusis/cluster"
)

func main() {
	if err := ipvs.Init(); err != nil {
		log.Fatalf("IPVS initialisation failed: %v\n", err)
	}
	log.Printf("IPVS version %s\n", ipvs.Version())

	env := os.Getenv("FUSIS_ENV")
	if env == "" {
		env = "development"
	}

	cluster.NewCluster()
	apiService := api.NewAPI(env)

	log.Printf("====> Running enviroment: %v\n", env)
	apiService.Serve()
}
