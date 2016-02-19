package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luizbafilho/janus/ipvs"
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
	var newService ipvs.Service
	err := json.NewDecoder(r.Body).Decode(&newService)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if err := ipvs.AddService(newService); err != nil {
		http.Error(w, fmt.Sprintf("ipvs.AddService() failed: %v\n", err), 422)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

//TodoShow expoer
func TodoShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	todoID := vars["todoId"]
	fmt.Fprintln(w, "Todo Fuck:", todoID)
}
