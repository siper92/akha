package v3

var _ FlowServices = (*FlowSP)(nil)

type FlowSP struct {
	debug Debug
	repo  TaskRepository
	tasks TaskServices
}

func NewFlowSP(debug Debug, repo TaskRepository, tasks TaskServices) *FlowSP {
	return &FlowSP{debug: debug, repo: repo, tasks: tasks}
}

func (s *FlowSP) DebugService() Debug        { return s.debug }
func (s *FlowSP) Tasks() TaskRepository      { return s.repo }
func (s *FlowSP) TaskServices() TaskServices { return s.tasks }
