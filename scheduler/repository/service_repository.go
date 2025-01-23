package repository

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Service struct {
	Server      string `json:"server"`
	Name        string `json:"name"`
	InputTopic  string `json:"inputTopic"`
	OutputTopic string `json:"outputTopic"`
}

type ServiceRepository struct {
	Services map[string]Service `json:"services"`
}

func NewServiceRepository() *ServiceRepository {
	filePath := os.Getenv("SERVICES_FILE_PATH")

	bytes, err := os.ReadFile(filePath)

	if err != nil {
		log.Printf("Failed to read services file: %s\n", err)
	}
	serviceRepository := ServiceRepository{}
	err = json.Unmarshal(bytes, &serviceRepository)
	if err != nil {
		log.Printf("Failed to unmarshal services file: %s\n", err)
	}
	return &serviceRepository
}

func (sr *ServiceRepository) GetService(name string) (Service, error) {
	service, ok := sr.Services[name]
	fmt.Println("EXISTING SERVICES", sr.Services)
	if !ok {
		return Service{}, fmt.Errorf("service not found: %s", name)
	}
	return service, nil
}
func (sr *ServiceRepository) GetServices() []Service {
	services := make([]Service, 0)
	for _, s := range sr.Services {
		services = append(services, s)
	}
	return services
}
