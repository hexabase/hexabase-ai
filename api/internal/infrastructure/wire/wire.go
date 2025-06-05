//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/hexabase/kaas-api/internal/api/handlers"
	"github.com/hexabase/kaas-api/internal/config"
	"github.com/hexabase/kaas-api/internal/domain/kubernetes"
	"github.com/hexabase/kaas-api/internal/domain/monitoring"
	k8sRepo "github.com/hexabase/kaas-api/internal/repository/kubernetes"
	monitoringRepo "github.com/hexabase/kaas-api/internal/repository/monitoring"
	monitoringService "github.com/hexabase/kaas-api/internal/service/monitoring"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"k8s.io/client-go/kubernetes"
)

// ProvideMonitoringRepository provides monitoring repository
func ProvideMonitoringRepository(db *gorm.DB) monitoring.Repository {
	return monitoringRepo.NewPostgresRepository(db)
}

// ProvideKubernetesRepository provides Kubernetes repository
func ProvideKubernetesRepository(client kubernetes.Interface) kubernetes.Repository {
	// This would be implemented when we create the k8s repository
	return k8sRepo.NewKubernetesRepository(client)
}

// ProvideMonitoringService provides monitoring service
func ProvideMonitoringService(
	repo monitoring.Repository,
	k8sRepo kubernetes.Repository,
	logger *zap.Logger,
) monitoring.Service {
	return monitoringService.NewService(repo, k8sRepo, logger)
}

// ProvideMonitoringHandler provides monitoring handler
func ProvideMonitoringHandler(
	service monitoring.Service,
	logger *zap.Logger,
) *handlers.MonitoringHandler {
	return handlers.NewMonitoringHandler(service, logger)
}

// MonitoringSet is a Wire provider set for monitoring
var MonitoringSet = wire.NewSet(
	ProvideMonitoringRepository,
	ProvideKubernetesRepository,
	ProvideMonitoringService,
	ProvideMonitoringHandler,
)

// InitializeMonitoringHandler creates a monitoring handler with all dependencies
func InitializeMonitoringHandler(
	db *gorm.DB,
	k8sClient kubernetes.Interface,
	logger *zap.Logger,
) (*handlers.MonitoringHandler, error) {
	wire.Build(MonitoringSet)
	return nil, nil
}

// InitializeApp creates the entire application with all dependencies
func InitializeApp(cfg *config.Config, db *gorm.DB, k8sClient kubernetes.Interface, logger *zap.Logger) (*App, error) {
	wire.Build(
		MonitoringSet,
		// Add other domain sets here as we refactor them
		// WorkspaceSet,
		// BillingSet,
		// etc.
		NewApp,
	)
	return nil, nil
}

// App represents the application with all handlers
type App struct {
	MonitoringHandler *handlers.MonitoringHandler
	// Add other handlers here
}

// NewApp creates a new App instance
func NewApp(
	monitoringHandler *handlers.MonitoringHandler,
	// Add other handlers as parameters
) *App {
	return &App{
		MonitoringHandler: monitoringHandler,
	}
}