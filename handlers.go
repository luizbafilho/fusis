package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"syscall"

	"github.com/google/seesaw/ipvs"
	"github.com/gorilla/mux"
)

//Index Handles index
func Index(w http.ResponseWriter, r *http.Request) {
	// todos := Todos{
	// 	Todo{Name: "Write presentation"},
	// 	Todo{Name: "Host meetup"},
	// }
	// w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	//
	// w.WriteHeader(http.StatusOK)
	// err := json.NewEncoder(w).Encode(todos)
	//
	// if err != nil {
	// 	panic(err)
	// }
}

//ServiceCreate ...
func ServiceCreate(w http.ResponseWriter, r *http.Request) {
	var newService Service
	err := json.NewDecoder(r.Body).Decode(&newService)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	ipvsService := ipvs.Service{
		Address:   net.ParseIP(newService.Host),
		Protocol:  syscall.IPPROTO_TCP,
		Port:      newService.Port,
		Scheduler: newService.Scheduler,
	}

	if err := ipvs.AddService(ipvsService); err != nil {
		log.Fatalf("ipvs.AddService() failed: %v\n", err)
	}

	w.WriteHeader(http.StatusCreated)
}

//TodoShow expoer
func TodoShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	todoID := vars["todoId"]
	fmt.Fprintln(w, "Todo Fuck:", todoID)
}
