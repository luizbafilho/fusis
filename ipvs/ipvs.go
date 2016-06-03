package ipvs

import (
	"sync"

	log "github.com/Sirupsen/logrus"
	ip_vs "github.com/google/seesaw/ipvs"
)

type Ipvs struct {
	sync.Mutex
}

func New() *Ipvs {
	log.Infof("Initialising IPVS Module...")
	if err := ip_vs.Init(); err != nil {
		log.Fatalf("IPVS initialisation failed: %v", err)
	}

	ipvs := &Ipvs{}
	if err := ipvs.Flush(); err != nil {
		log.Fatalf("IPVS flushing table failed: %v", err)
	}

	return ipvs
}

// Flush flushes all services and destinations from the IPVS table.
func (ipvs *Ipvs) Flush() error {
	return ip_vs.Flush()
}

// GetServices gets the current services
func (ipvs *Ipvs) GetServices() ([]*ip_vs.Service, error) {
	ipvs.Lock()
	defer ipvs.Unlock()
	return ip_vs.GetServices()
}

// GetService gets given service
func (ipvs *Ipvs) GetService(svc *ip_vs.Service) (*ip_vs.Service, error) {
	ipvs.Lock()
	defer ipvs.Unlock()
	return ip_vs.GetService(svc)
}

// AddService adds given service to IPVS table.
func (ipvs *Ipvs) AddService(svc *ip_vs.Service) error {
	ipvs.Lock()
	defer ipvs.Unlock()
	return ip_vs.AddService(*svc)
}

// DeleteService deletes given service from IPVS table.
func (ipvs *Ipvs) DeleteService(svc *ip_vs.Service) error {
	ipvs.Lock()
	defer ipvs.Unlock()
	return ip_vs.DeleteService(*svc)
}

// AddDestination adds given destination to the IPVS table.
func (ipvs *Ipvs) AddDestination(svc ip_vs.Service, dst ip_vs.Destination) error {
	ipvs.Lock()
	defer ipvs.Unlock()
	return ip_vs.AddDestination(svc, dst)
}

// GetDestinations gets all destination from a service
func (ipvs *Ipvs) GetDestinations(svc *ip_vs.Service) ([]*ip_vs.Destination, error) {
	ipvs.Lock()
	defer ipvs.Unlock()
	svc, err := ip_vs.GetService(svc)
	if err != nil {
		return nil, err
	}

	return svc.Destinations, nil
}

// DeleteDestination deletes given destination from the IPVS table.
func (ipvs *Ipvs) DeleteDestination(svc ip_vs.Service, dst ip_vs.Destination) error {
	ipvs.Lock()
	defer ipvs.Unlock()
	return ip_vs.DeleteDestination(svc, dst)
}
