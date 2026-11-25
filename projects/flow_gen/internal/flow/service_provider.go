package flow

import (
	"github.com/siper92/akha/projects/flow_gen/internal/service/ai"
	"github.com/siper92/akha/projects/flow_gen/internal/service/vfs"
)

var _ ServiceProvider = (*FlowService)(nil)

type FlowService struct {
	ai ai.Client
	fs vfs.FSService
}

func (f FlowService) Service(name string) (any, bool) {
	switch name {
	case "ai":
		return f.ai, true
	case "fs":
		return f.fs, true
	default:
		return nil, false
	}
}
