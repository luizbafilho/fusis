package store

type Store interface {
	// GetService(svc ServiceRequest) (ServiceRequest, error)
	// GetServices([]ServiceRequest) (ServiceRequest, error)
	AddService(svc ServiceRequest) error
	// UpdateService(svc ServiceRequest) error
	// DeleteService(svc ServiceRequest) error
	//
	// AddDestination(dst DestinationRequest) error
	// UpdateDestination(dst DestinationRequest) error
	// DeleteDestination(dst DestinationRequest) error
}
