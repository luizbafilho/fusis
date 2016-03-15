package engine

func GetServices() (*[]Service, error) {
	ipvsSvcs, err := IPVSGetServices()
	if err != nil {
		return nil, err
	}

	services := []Service{}
	for _, svc := range ipvsSvcs {
		services = append(services, NewService(svc))
	}

	return &services, nil
}
