package v3

import "github.com/siper92/akha/internal/ai"

var _ TaskServices = (*TaskSP)(nil)

type TaskSP struct {
	debug Debug
	asker Asker
	tools ai.ToolRunner
}

func NewTaskSP(debug Debug, asker Asker, tools ai.ToolRunner) *TaskSP {
	return &TaskSP{debug: debug, asker: asker, tools: tools}
}

func (s *TaskSP) DebugService() Debug  { return s.debug }
func (s *TaskSP) AI() Asker            { return s.asker }
func (s *TaskSP) Tools() ai.ToolRunner { return s.tools }
