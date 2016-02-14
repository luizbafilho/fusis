package main

import "net/http"

type Route struct {
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"GET",
		"/",
		Index,
	},
	Route{
		"GET",
		"/todos/{todoId}",
		TodoShow,
	},
	Route{
		"POST",
		"/services",
		ServiceCreate,
	},
}
