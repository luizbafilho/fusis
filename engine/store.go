package engine

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
