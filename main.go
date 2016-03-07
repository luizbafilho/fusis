package main

import (
	"fmt"
	"log"
	"os"

	"github.com/luizbafilho/fusis/api"
	"github.com/luizbafilho/fusis/engine"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/store/etcd"
)

func main() {
	if err := ipvs.Init(); err != nil {
		log.Fatalf("IPVS initialisation failed: %v\n", err)
	}
	log.Printf("IPVS version %s\n", ipvs.Version())

	nodes := []string{"http://127.0.0.1:2379"}

	env := os.Getenv("FUSIS_ENV")
	if env == "" {
		env = "development"
	}

	s := etcd.New(nodes, fmt.Sprintf("fusis_%v", env))

	apiService := api.NewAPI(s, env)
	engineService := engine.NewEngine(s)

	log.Printf("====> Running enviroment: %v\n", env)
	engineService.Serve()
	apiService.Serve()
}
