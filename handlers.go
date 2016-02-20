package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/luizbafilho/janus/ipvs"
)

//Index Handles index
func ServiceIndex(w http.ResponseWriter, r *http.Request) {
	ipvsServices, err := ipvs.GetServices()

	if err != nil {
		http.Error(w, fmt.Sprintf("ipvs.GetServices() failed: %v\n", err), 422)
		return
	}

	services := []ServiceRequest{}

	for _, s := range ipvsServices {
		services = append(services, newServiceRequest(s))
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(services)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
}

//ServiceCreate ...
func ServiceCreate(w http.ResponseWriter, r *http.Request) {
	var newService ServiceRequest
	err := json.NewDecoder(r.Body).Decode(&newService)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	err = ipvs.AddService(newService.toIpvsService())

	if err != nil {
		http.Error(w, fmt.Sprintf("ipvs.AddService() failed: %v\n", err), 422)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func ServiceUpdate(w http.ResponseWriter, r *http.Request) {
	var service ServiceRequest
	err := json.NewDecoder(r.Body).Decode(&service)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	err = ipvs.UpdateService(service.toIpvsService())

	if err != nil {
		http.Error(w, fmt.Sprintf("ipvs.UpdateService() failed: %v\n", err), 422)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func ServiceDelete(w http.ResponseWriter, r *http.Request) {
	var service ServiceRequest
	err := json.NewDecoder(r.Body).Decode(&service)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	err = ipvs.DeleteService(service.toIpvsService())

	if err != nil {
		http.Error(w, fmt.Sprintf("ipvs.DeleteService() failed: %v\n", err), 422)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DestinationCreate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	service_id := vars["service_id"]
	service_attrs := strings.Split(service_id, "-")

	port, err := strconv.ParseUint(service_attrs[1], 10, 16)
	if err != nil {
		http.Error(w, fmt.Sprintf("Port invalid: %v\n", err), 422)
		return
	}

	var proto IPProto
	proto.UnmarshalJSON([]byte(service_attrs[2]))

	service := ipvs.Service{
		Address:  net.ParseIP(service_attrs[0]),
		Port:     uint16(port),
		Protocol: ipvs.IPProto(proto),
	}

	var destination DestinationRequest
	err = json.NewDecoder(r.Body).Decode(&destination)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	err = ipvs.AddDestination(service, *destination.toIpvsDestination())

	if err != nil {
		http.Error(w, fmt.Sprintf("ipvs.AddDestination() failed: %v\n", err), 422)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DestinationUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	service_id := vars["service_id"]
	service_attrs := strings.Split(service_id, "-")

	port, err := strconv.ParseUint(service_attrs[1], 10, 16)
	if err != nil {
		http.Error(w, fmt.Sprintf("Port invalid: %v\n", err), 422)
		return
	}

	var proto IPProto
	proto.UnmarshalJSON([]byte(service_attrs[2]))

	service := ipvs.Service{
		Address:  net.ParseIP(service_attrs[0]),
		Port:     uint16(port),
		Protocol: ipvs.IPProto(proto),
	}

	var destination DestinationRequest
	err = json.NewDecoder(r.Body).Decode(&destination)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	err = ipvs.UpdateDestination(service, *destination.toIpvsDestination())

	if err != nil {
		http.Error(w, fmt.Sprintf("ipvs.UpdateDestination() failed: %v\n", err), 422)
		return
	}
}

func DestinationDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	service_id := vars["service_id"]
	service_attrs := strings.Split(service_id, "-")

	port, err := strconv.ParseUint(service_attrs[1], 10, 16)
	if err != nil {
		http.Error(w, fmt.Sprintf("Port invalid: %v\n", err), 422)
		return
	}

	var proto IPProto
	proto.UnmarshalJSON([]byte(service_attrs[2]))

	service := ipvs.Service{
		Address:  net.ParseIP(service_attrs[0]),
		Port:     uint16(port),
		Protocol: ipvs.IPProto(proto),
	}

	var destination DestinationRequest
	err = json.NewDecoder(r.Body).Decode(&destination)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	err = ipvs.DeleteDestination(service, *destination.toIpvsDestination())

	if err != nil {
		http.Error(w, fmt.Sprintf("ipvs.DeleteDestination() failed: %v\n", err), 422)
		return
	}
}
