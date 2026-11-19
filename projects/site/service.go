package site

import (
	"errors"
	"fmt"
	"time"

	"github.com/siper92/akha/internal/cache"
	flow "github.com/siper92/akha/internal/flow_v3"
	"github.com/siper92/akha/projects/site/sdk/db"
)

var (
	ErrNoCacheDir    = errors.New("no cache dir")
	ErrInvalidName   = errors.New("invalid file name")
	ErrNoSiteService = errors.New("service is not a site service")
)

type SiteService interface {
	flow.DataServiceProvide[SiteData]
	Cache() cache.Cache
	Tasks() db.TaskRepository
}

var _ SiteService = (*siteService)(nil)

type siteService struct {
	*flow.StepData[SiteData]
	cache cache.Cache
	tasks db.TaskRepository
}

func newSiteService(d SiteData, cache cache.Cache, tasks db.TaskRepository) *siteService {
	return &siteService{
		StepData: flow.NewStepData(d),
		cache:    cache,
		tasks:    tasks,
	}
}

func (s *siteService) Tasks() db.TaskRepository {
	return s.tasks
}

func (s *siteService) ServiceProvide() flow.ServiceProvide {
	return s
}

func (s *siteService) Cache() cache.Cache {
	return s.cache
}

func asSiteService(s flow.DataServiceProvide[SiteData]) (SiteService, error) {
	svc, ok := s.ServiceProvide().(SiteService)
	if !ok {
		return nil, ErrNoSiteService
	}

	return svc, nil
}

func stepResponseName(name flow.StepName, at time.Time) string {
	return fmt.Sprintf("%s_response_%d.md", name, at.UnixMicro())
}
