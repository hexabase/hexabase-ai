package application

import (
	"github.com/hexabase/hexabase-ai/api/internal/domain/backup"
	"github.com/hexabase/hexabase-ai/api/internal/domain/monitoring"
	"github.com/hexabase/hexabase-ai/api/internal/domain/project"
)

// ExtendedService extends the base Service with additional dependencies for backup integration
type ExtendedService struct {
	*Service
	projectRepo       project.Repository
	backupService     backup.Service
	monitoringService monitoring.Service
}

// NewExtendedService creates a new extended application service with backup support
func NewExtendedService(
	service *Service,
	projectRepo project.Repository,
	backupService backup.Service,
	monitoringService monitoring.Service,
) *ExtendedService {
	return &ExtendedService{
		Service:           service,
		projectRepo:       projectRepo,
		backupService:     backupService,
		monitoringService: monitoringService,
	}
}