package flow

import "aiflow/ai"

type TaskServiceProvider interface {
	AI() ai.Client
	Tools() []ai.Tool
}

type TaskServices struct {
	client ai.Client
	tools  []ai.Tool
}

var _ TaskServiceProvider = (*TaskServices)(nil)

func NewTaskServices(client ai.Client, tools []ai.Tool) *TaskServices {
	return &TaskServices{client: client, tools: tools}
}

func (s *TaskServices) AI() ai.Client { return s.client }

func (s *TaskServices) Tools() []ai.Tool { return s.tools }
