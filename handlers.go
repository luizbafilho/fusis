package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"

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
	var service Service
	err := json.NewDecoder(r.Body).Decode(&service)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	_, err = exec.Command("sudo", "bash", "-c", "ipvsadm -A -t 172.18.0.200:80").Output()

	if err != nil {
		println(err.Error())
		return
	}

}

//TodoShow expoer
func TodoShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	todoID := vars["todoId"]
	fmt.Fprintln(w, "Todo Fuck:", todoID)
}
