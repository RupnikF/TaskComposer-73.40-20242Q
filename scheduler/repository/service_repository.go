package repository

import "os"

type Service struct {
	Server string
	Topic  string
}

type ServiceRepository struct {
	serviceMap map[string]*Service
}

func NewServiceRepository() *ServiceRepository {
	return &ServiceRepository{
		map[string]*Service{
			"echo-service": {
				Server: os.Getenv("NATIVE_HOST") + ":" + os.Getenv("NATIVE_PORT"),
				Topic:  os.Getenv("NATIVE_TOPIC"),
			},
			"native": {
				Server: "",
				Topic:  "",
			},
		},
	}
}

func (r *ServiceRepository) GetService(serviceName string) *Service {
	return r.serviceMap[serviceName]
}
