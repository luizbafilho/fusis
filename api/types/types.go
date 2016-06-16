package types

var (
	ErrServiceNotFound     error = ErrNotFound("service not found")
	ErrDestinationNotFound error = ErrNotFound("destination not found")
)

type ErrNotFound string

func (e ErrNotFound) Error() string {
	return string(e)
}

type Service struct {
	Name         string `valid:"required"`
	Host         string
	Port         uint16 `valid:"required"`
	Protocol     string `valid:"required"`
	Scheduler    string `valid:"required"`
	Destinations []Destination
}

type Destination struct {
	Name      string `valid:"required"`
	Host      string `valid:"required"`
	Port      uint16 `valid:"required"`
	Weight    int32
	Mode      string `valid:"required"`
	ServiceId string `valid:"required"`
}

func (svc Service) GetId() string {
	return svc.Name
}

func (dst Destination) GetId() string {
	return dst.Name
}
