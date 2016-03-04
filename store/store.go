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
	Service     ServiceRequest
	Destination DestinationRequest
}

type Store interface {
	// GetService(svc ServiceRequest) (ServiceRequest, error)
	GetServices() (*[]ServiceRequest, error)
	UpsertService(svc ServiceRequest) error
	DeleteService(svc ServiceRequest) error

	UpsertDestination(svc ServiceRequest, dst DestinationRequest) error
	DeleteDestination(svc ServiceRequest, dst DestinationRequest) error

	Subscribe(changes chan interface{}) error
}
