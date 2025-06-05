//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/hexabase/hexabase-kaas/api/internal/api/handlers"
	"github.com/hexabase/hexabase-kaas/api/internal/config"
	
	// Domain imports
	"github.com/hexabase/hexabase-kaas/api/internal/domain/auth"
	"github.com/hexabase/hexabase-kaas/api/internal/domain/billing"
	"github.com/hexabase/hexabase-kaas/api/internal/domain/monitoring"
	"github.com/hexabase/hexabase-kaas/api/internal/domain/organization"
	"github.com/hexabase/hexabase-kaas/api/internal/domain/project"
	"github.com/hexabase/hexabase-kaas/api/internal/domain/workspace"
	
	// Repository imports
	authRepo "github.com/hexabase/hexabase-kaas/api/internal/repository/auth"
	billingRepo "github.com/hexabase/hexabase-kaas/api/internal/repository/billing"
	monitoringRepo "github.com/hexabase/hexabase-kaas/api/internal/repository/monitoring"
	orgRepo "github.com/hexabase/hexabase-kaas/api/internal/repository/organization"
	projectRepo "github.com/hexabase/hexabase-kaas/api/internal/repository/project"
	workspaceRepo "github.com/hexabase/hexabase-kaas/api/internal/repository/workspace"
	
	// Service imports
	authSvc "github.com/hexabase/hexabase-kaas/api/internal/service/auth"
	billingSvc "github.com/hexabase/hexabase-kaas/api/internal/service/billing"
	monitoringSvc "github.com/hexabase/hexabase-kaas/api/internal/service/monitoring"
	orgSvc "github.com/hexabase/hexabase-kaas/api/internal/service/organization"
	projectSvc "github.com/hexabase/hexabase-kaas/api/internal/service/project"
	workspaceSvc "github.com/hexabase/hexabase-kaas/api/internal/service/workspace"
	
	"go.uber.org/zap"
	"gorm.io/gorm"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// AuthSet is a Wire provider set for authentication
var AuthSet = wire.NewSet(
	authRepo.NewPostgresRepository,
	wire.Bind(new(auth.Repository), new(*authRepo.PostgresRepository)),
	
	authRepo.NewOAuthRepository,
	wire.Bind(new(auth.OAuthRepository), new(*authRepo.OAuthRepository)),
	
	authRepo.NewKeyRepository,
	wire.Bind(new(auth.KeyRepository), new(*authRepo.KeyRepository)),
	
	authSvc.NewService,
	wire.Bind(new(auth.Service), new(*authSvc.Service)),
	
	handlers.NewAuthHandler,
)

// BillingSet is a Wire provider set for billing
var BillingSet = wire.NewSet(
	billingRepo.NewPostgresRepository,
	wire.Bind(new(billing.Repository), new(*billingRepo.PostgresRepository)),
	
	billingRepo.NewStripeRepository,
	wire.Bind(new(billing.StripeRepository), new(*billingRepo.StripeRepository)),
	
	billingSvc.NewService,
	wire.Bind(new(billing.Service), new(*billingSvc.Service)),
	
	handlers.NewBillingHandler,
)

// MonitoringSet is a Wire provider set for monitoring
var MonitoringSet = wire.NewSet(
	monitoringRepo.NewPostgresRepository,
	wire.Bind(new(monitoring.Repository), new(*monitoringRepo.PostgresRepository)),
	
	monitoringRepo.NewKubernetesRepository,
	wire.Bind(new(monitoring.KubernetesRepository), new(*monitoringRepo.KubernetesRepository)),
	
	monitoringSvc.NewService,
	wire.Bind(new(monitoring.Service), new(*monitoringSvc.Service)),
	
	handlers.NewMonitoringHandler,
)

// OrganizationSet is a Wire provider set for organizations
var OrganizationSet = wire.NewSet(
	orgRepo.NewPostgresRepository,
	wire.Bind(new(organization.Repository), new(*orgRepo.PostgresRepository)),
	
	orgSvc.NewService,
	wire.Bind(new(organization.Service), new(*orgSvc.Service)),
	
	handlers.NewOrganizationHandler,
)

// ProjectSet is a Wire provider set for projects
var ProjectSet = wire.NewSet(
	projectRepo.NewPostgresRepository,
	wire.Bind(new(project.Repository), new(*projectRepo.PostgresRepository)),
	
	projectRepo.NewKubernetesRepository,
	wire.Bind(new(project.KubernetesRepository), new(*projectRepo.KubernetesRepository)),
	
	projectSvc.NewService,
	wire.Bind(new(project.Service), new(*projectSvc.Service)),
	
	handlers.NewProjectHandler,
)

// WorkspaceSet is a Wire provider set for workspaces
var WorkspaceSet = wire.NewSet(
	workspaceRepo.NewPostgresRepository,
	wire.Bind(new(workspace.Repository), new(*workspaceRepo.PostgresRepository)),
	
	workspaceRepo.NewKubernetesRepository,
	wire.Bind(new(workspace.KubernetesRepository), new(*workspaceRepo.KubernetesRepository)),
	
	workspaceSvc.NewService,
	wire.Bind(new(workspace.Service), new(*workspaceSvc.Service)),
	
	handlers.NewWorkspaceHandler,
)

// ProvideOAuthConfig provides OAuth configuration
func ProvideOAuthConfig(cfg *config.Config) map[string]*authRepo.ProviderConfig {
	// Convert config to OAuth provider configs
	providers := make(map[string]*authRepo.ProviderConfig)
	// Add provider configurations from config
	return providers
}

// ProvideStripeConfig provides Stripe configuration
func ProvideStripeConfig(cfg *config.Config) (string, string) {
	// Return API key and webhook secret from config
	return "", "" // Replace with actual config values
}

// SharedDependencies provides shared dependencies
var SharedDependencies = wire.NewSet(
	wire.Bind(new(organization.AuthRepository), new(*authRepo.PostgresRepository)),
	wire.Bind(new(organization.BillingRepository), new(*billingRepo.StripeRepository)),
	wire.Bind(new(workspace.AuthRepository), new(*authRepo.PostgresRepository)),
)

// App represents the application with all handlers
type App struct {
	AuthHandler         *handlers.AuthHandler
	BillingHandler      *handlers.BillingHandler
	MonitoringHandler   *handlers.MonitoringHandler
	OrganizationHandler *handlers.OrganizationHandler
	ProjectHandler      *handlers.ProjectHandler
	WorkspaceHandler    *handlers.WorkspaceHandler
}

// NewApp creates a new App instance
func NewApp(
	authHandler *handlers.AuthHandler,
	billingHandler *handlers.BillingHandler,
	monitoringHandler *handlers.MonitoringHandler,
	organizationHandler *handlers.OrganizationHandler,
	projectHandler *handlers.ProjectHandler,
	workspaceHandler *handlers.WorkspaceHandler,
) *App {
	return &App{
		AuthHandler:         authHandler,
		BillingHandler:      billingHandler,
		MonitoringHandler:   monitoringHandler,
		OrganizationHandler: organizationHandler,
		ProjectHandler:      projectHandler,
		WorkspaceHandler:    workspaceHandler,
	}
}

// InitializeApp creates the entire application with all dependencies
func InitializeApp(
	cfg *config.Config,
	db *gorm.DB,
	k8sClient kubernetes.Interface,
	dynamicClient dynamic.Interface,
	k8sConfig *rest.Config,
	logger *zap.Logger,
) (*App, error) {
	wire.Build(
		AuthSet,
		BillingSet,
		MonitoringSet,
		OrganizationSet,
		ProjectSet,
		WorkspaceSet,
		SharedDependencies,
		ProvideOAuthConfig,
		ProvideStripeConfig,
		NewApp,
	)
	return nil, nil
}