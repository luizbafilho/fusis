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
	Route{
		"PUT",
		"/services",
		ServiceUpdate,
	},
	Route{
		"DELETE",
		"/services",
		ServiceDelete,
	},
	Route{
		"POST",
		"/services/{service_id}/destinations",
		DestinationCreate,
	},
	Route{
		"PUT",
		"/services/{service_id}/destinations",
		DestinationUpdate,
	},
	Route{
		"DELETE",
		"/services/{service_id}/destinations",
		DestinationDelete,
	},
}
