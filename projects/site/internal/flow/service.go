package flow

import (
	"errors"

	"github.com/siper92/akha/internal/cache"
	flow "github.com/siper92/akha/internal/flow_v3"
	"github.com/siper92/akha/projects/site/internal/db"
)

var (
	ErrNoCacheDir  = errors.New("no cache dir")
	ErrInvalidName = errors.New("invalid file name")
)

type SiteService interface {
	flow.DataServiceProvide[SiteData]
	AI() asker
	Cache() cache.Cache
	Tasks() db.TaskRepository
}

type Providers struct {
	AI    asker
	Cache cache.Cache
	Tasks db.TaskRepository
}

var _ SiteService = (*siteService)(nil)

type siteService struct {
	*flow.StepData[SiteData]
	p Providers
}

func newSiteService(d SiteData, p Providers) *siteService {
	return &siteService{StepData: flow.NewStepData(d), p: p}
}

func (s *siteService) ServiceProvide() flow.ServiceProvide {
	return s
}

func (s *siteService) AI() asker {
	return s.p.AI
}

func (s *siteService) Cache() cache.Cache {
	return s.p.Cache
}

func (s *siteService) Tasks() db.TaskRepository {
	return s.p.Tasks
}

func siteStep(fn flow.ServiceExec[SiteData, SiteService]) flow.StepExec[SiteData] {
	return flow.WithService(fn)
}
