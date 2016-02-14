package main

import (
	"reflect"
	"runtime"

	"github.com/gorilla/mux"
)

//NewRouter returns a mux router
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		routeName := runtime.FuncForPC(reflect.ValueOf(route.HandlerFunc).Pointer()).Name()

		handler := Logger(route.HandlerFunc, routeName)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Handler(handler)

	}
	return router
}
