package main

import (
	"log"

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
	s := etcd.New(nodes, "fusis")

	apiService := api.NewAPI(s)
	engineService := engine.NewEngine(s)

	engineService.Serve()
	apiService.Serve()
}
