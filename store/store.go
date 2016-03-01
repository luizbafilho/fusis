package store

const (
	CreateEvent = "create"
	UpdateEvent = "update"
	SetEvent    = "set"
	DeleteEvent = "delete"
)

type ServiceEvent struct {
	Action  string
	Service ServiceRequest
}

type DestinationEvent struct {
	Action      string
	Destination DestinationRequest
}

type Store interface {
	// GetService(svc ServiceRequest) (ServiceRequest, error)
	// GetServices([]ServiceRequest) (ServiceRequest, error)
	AddService(svc ServiceRequest) error
	UpdateService(svc ServiceRequest) error
	DeleteService(svc ServiceRequest) error
	//
	// AddDestination(dst DestinationRequest) error
	// UpdateDestination(dst DestinationRequest) error
	// DeleteDestination(dst DestinationRequest) error
	Subscribe(changes chan interface{}) error
}
