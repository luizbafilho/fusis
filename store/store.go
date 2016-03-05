package store

const (
	CreateEvent = "create"
	UpdateEvent = "update"
	SetEvent    = "set"
	DeleteEvent = "delete"
)

type ServiceEvent struct {
	Action  string
	Service Service
}

type DestinationEvent struct {
	Action      string
	Service     Service
	Destination Destination
}

type Store interface {
	// GetService(svc ServiceRequest) (ServiceRequest, error)
	GetServices() (*[]Service, error)
	UpsertService(svc Service) error
	DeleteService(svc Service) error

	UpsertDestination(svc Service, dst Destination) error
	DeleteDestination(svc Service, dst Destination) error

	Subscribe(changes chan interface{}) error
}
