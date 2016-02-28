package main

import (
	"log"

	"github.com/luizbafilho/janus/api"
	"github.com/luizbafilho/janus/engine"
	"github.com/luizbafilho/janus/ipvs"
	"github.com/luizbafilho/janus/store/etcd"
)

func main() {
	if err := ipvs.Init(); err != nil {
		log.Fatalf("IPVS initialisation failed: %v\n", err)
	}
	log.Printf("IPVS version %s\n", ipvs.Version())

	nodes := []string{"http://127.0.0.1:2379"}
	s := etcd.New(nodes)

	apiService := api.NewAPI(s)
	engineService := engine.NewEngine(s)

	engineService.Serve()
	apiService.Serve()
}
