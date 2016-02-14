package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

//Index Handles index
func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK!")
}

//TodoShow expoer
func TodoShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	todoID := vars["todoId"]
	fmt.Fprintln(w, "Todo Fuck:", todoID)
}
