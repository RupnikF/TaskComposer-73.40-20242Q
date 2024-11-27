package repository

type Service struct {
	Server string
	Topic  string
}

type ServiceRepository struct {
	serviceMap map[string]*Service
}

func NewServiceRepository() ServiceRepository {
	return ServiceRepository{
		map[string]*Service{
			"native": {
				Server: "aa",
				Topic:  "messageTopic",
			},
		},
	}
}
