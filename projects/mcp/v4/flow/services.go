package flow

type ServiceProvider interface {
	Repository() TaskRepository
	Tasks() TaskServiceProvider
}

type Services struct {
	repository TaskRepository
	tasks      TaskServiceProvider
}

var _ ServiceProvider = (*Services)(nil)

func NewServices(repository TaskRepository, tasks TaskServiceProvider) *Services {
	return &Services{repository: repository, tasks: tasks}
}

func (s *Services) Repository() TaskRepository { return s.repository }

func (s *Services) Tasks() TaskServiceProvider { return s.tasks }
