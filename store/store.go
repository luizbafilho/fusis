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
	GetService(serviceId string) (*Service, error)
	GetServices() (*[]Service, error)

	UpsertService(svc Service) error
	DeleteService(svc Service) error

	GetDestinations(svc Service) (*[]Destination, error)
	UpsertDestination(svc Service, dst Destination) error
	DeleteDestination(svc Service, dst Destination) error

	Subscribe(changes chan interface{}) error
	Flush() error
}
