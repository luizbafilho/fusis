package engine

import (
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/provider"
	"github.com/luizbafilho/fusis/steps"
)

// Adding service to store
type addServiceStore struct {
	Service *ipvs.Service
}

func (as addServiceStore) Do(prev steps.Result) (steps.Result, error) {
	return nil, ipvs.Store.AddService(as.Service)
}
func (as addServiceStore) Undo() error {
	return ipvs.Store.DeleteService(as.Service)
}

// Adding service to Ipvs module
type addServiceIpvs struct {
	Service *ipvs.Service
}

func (as addServiceIpvs) Do(prev steps.Result) (steps.Result, error) {
	return nil, ipvs.Kernel.AddService(as.Service.ToIpvsService())
}

func (as addServiceIpvs) Undo() error {
	return ipvs.Kernel.DeleteService(as.Service.ToIpvsService())
}

//Allocating a VIP and setting it to the network interface
type setVip struct {
	Service *ipvs.Service
}

func (sv setVip) Do(prev steps.Result) (steps.Result, error) {
	prov, err := provider.GetProvider()
	if err != nil {
		return nil, err
	}

	return nil, prov.AssignVIP(*sv.Service)
}
func (sv setVip) Undo() error {
	prov, err := provider.GetProvider()
	if err != nil {
		return err
	}

	return prov.UnassignVIP(*sv.Service)
}

// Deleting service from store
type deleteServiceStore struct {
	Service *ipvs.Service
}

func (ds deleteServiceStore) Do(prev steps.Result) (steps.Result, error) {
	return nil, ipvs.Store.DeleteService(ds.Service)
}
func (ds deleteServiceStore) Undo() error {
	return ipvs.Store.AddService(ds.Service)
}

// Deleting service from ipvs module
type deleteServiceIpvs struct {
	Service *ipvs.Service
}

func (ds deleteServiceIpvs) Do(prev steps.Result) (steps.Result, error) {
	return nil, ipvs.Kernel.DeleteService(ds.Service.ToIpvsService())
}
func (ds deleteServiceIpvs) Undo() error {
	return ipvs.Kernel.AddService(ds.Service.ToIpvsService())
}

type addDestinationStore struct {
	Service     *ipvs.Service
	Destination *ipvs.Destination
}

func (ad addDestinationStore) Do(prev steps.Result) (steps.Result, error) {
	ad.Destination.ServiceId = ad.Service.GetId()
	return nil, ipvs.Store.AddDestination(ad.Destination)
}

func (ad addDestinationStore) Undo() error {
	return ipvs.Store.DeleteDestination(ad.Destination)
}

type addDestinationIpvs struct {
	Service     *ipvs.Service
	Destination *ipvs.Destination
}

func (ad addDestinationIpvs) Do(prev steps.Result) (steps.Result, error) {
	return nil, ipvs.Kernel.AddDestination(*ad.Service.ToIpvsService(), *ad.Destination.ToIpvsDestination())
}

func (ad addDestinationIpvs) Undo() error {
	return ipvs.Kernel.DeleteDestination(*ad.Service.ToIpvsService(), *ad.Destination.ToIpvsDestination())
}

// Deleting Destination from store
type deleteDestinationStore struct {
	Service     *ipvs.Service
	Destination *ipvs.Destination
}

func (dd deleteDestinationStore) Do(prev steps.Result) (steps.Result, error) {
	err := ipvs.Store.DeleteDestination(dd.Destination)
	return nil, err
}
func (dd deleteDestinationStore) Undo() error {
	return ipvs.Store.AddDestination(dd.Destination)
}

// Deleting Destination from ipvs module
type deleteDestinationIpvs struct {
	Service     *ipvs.Service
	Destination *ipvs.Destination
}

func (dd deleteDestinationIpvs) Do(prev steps.Result) (steps.Result, error) {
	err := ipvs.Kernel.DeleteDestination(*dd.Service.ToIpvsService(), *dd.Destination.ToIpvsDestination())
	return nil, err
}
func (dd deleteDestinationIpvs) Undo() error {
	return ipvs.Kernel.AddDestination(*dd.Service.ToIpvsService(), *dd.Destination.ToIpvsDestination())
}
