//go:build wireinject
// +build wireinject

package wire

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/wire"
	"github.com/hexabase/hexabase-ai/api/internal/api/handlers"

	// Application domain - using new package-by-feature structure
	"github.com/hexabase/hexabase-ai/api/internal/application/domain"
	applicationHandler "github.com/hexabase/hexabase-ai/api/internal/application/handler"
	applicationRepo "github.com/hexabase/hexabase-ai/api/internal/application/repository"
	applicationSvc "github.com/hexabase/hexabase-ai/api/internal/application/service"

	// Auth domain
	authDomain "github.com/hexabase/hexabase-ai/api/internal/auth/domain"
	authHandler "github.com/hexabase/hexabase-ai/api/internal/auth/handler"
	authRepo "github.com/hexabase/hexabase-ai/api/internal/auth/repository"
	authSvc "github.com/hexabase/hexabase-ai/api/internal/auth/service"

	// Organization domain

	orgHandler "github.com/hexabase/hexabase-ai/api/internal/organization/handler"
	orgRepo "github.com/hexabase/hexabase-ai/api/internal/organization/repository"
	orgSvc "github.com/hexabase/hexabase-ai/api/internal/organization/service"

	// Project domain
	projectDomain "github.com/hexabase/hexabase-ai/api/internal/project/domain"
	projectHandler "github.com/hexabase/hexabase-ai/api/internal/project/handler"
	projectRepo "github.com/hexabase/hexabase-ai/api/internal/project/repository"
	projectSvc "github.com/hexabase/hexabase-ai/api/internal/project/service"

	// Workspace domain
	workspaceDomain "github.com/hexabase/hexabase-ai/api/internal/workspace/domain"
	workspaceHandler "github.com/hexabase/hexabase-ai/api/internal/workspace/handler"
	workspaceRepo "github.com/hexabase/hexabase-ai/api/internal/workspace/repository"
	workspaceSvc "github.com/hexabase/hexabase-ai/api/internal/workspace/service"

	// Legacy domains that haven't been migrated yet
	"github.com/hexabase/hexabase-ai/api/internal/domain/aiops"
	"github.com/hexabase/hexabase-ai/api/internal/domain/backup"
	"github.com/hexabase/hexabase-ai/api/internal/domain/billing"
	"github.com/hexabase/hexabase-ai/api/internal/domain/cicd"
	"github.com/hexabase/hexabase-ai/api/internal/domain/function"
	"github.com/hexabase/hexabase-ai/api/internal/domain/logs"
	"github.com/hexabase/hexabase-ai/api/internal/domain/monitoring"
	"github.com/hexabase/hexabase-ai/api/internal/domain/node"

	// Legacy repositories that haven't been migrated yet
	aiopsRepo "github.com/hexabase/hexabase-ai/api/internal/repository/aiops"
	backupRepo "github.com/hexabase/hexabase-ai/api/internal/repository/backup"
	billingRepo "github.com/hexabase/hexabase-ai/api/internal/repository/billing"
	cicdRepo "github.com/hexabase/hexabase-ai/api/internal/repository/cicd"
	functionRepo "github.com/hexabase/hexabase-ai/api/internal/repository/function"
	k8sRepo "github.com/hexabase/hexabase-ai/api/internal/repository/kubernetes"
	logRepo "github.com/hexabase/hexabase-ai/api/internal/repository/logs"
	monitoringRepo "github.com/hexabase/hexabase-ai/api/internal/repository/monitoring"
	nodeRepo "github.com/hexabase/hexabase-ai/api/internal/repository/node"
	"github.com/hexabase/hexabase-ai/api/internal/repository/proxmox"

	// Legacy services that haven't been migrated yet
	aiopsSvc "github.com/hexabase/hexabase-ai/api/internal/service/aiops"
	backupSvc "github.com/hexabase/hexabase-ai/api/internal/service/backup"
	billingSvc "github.com/hexabase/hexabase-ai/api/internal/service/billing"
	cicdSvc "github.com/hexabase/hexabase-ai/api/internal/service/cicd"
	functionSvc "github.com/hexabase/hexabase-ai/api/internal/service/function"
	logSvc "github.com/hexabase/hexabase-ai/api/internal/service/logs"
	monitoringSvc "github.com/hexabase/hexabase-ai/api/internal/service/monitoring"
	nodeSvc "github.com/hexabase/hexabase-ai/api/internal/service/node"

	"github.com/hexabase/hexabase-ai/api/internal/helm"
	"github.com/hexabase/hexabase-ai/api/internal/shared/config"
	"gorm.io/gorm"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

// Updated wire sets for migrated packages
var ApplicationSet = wire.NewSet(
	applicationRepo.NewPostgresRepository,
	applicationRepo.NewKubernetesRepository,
	applicationSvc.NewService,
	applicationHandler.NewApplicationHandler,
)

var AuthSet = wire.NewSet(
	authRepo.NewPostgresRepository,
	authRepo.NewOAuthRepository,
	authRepo.NewKeyRepository,
	authSvc.NewService,
	authHandler.NewHandler,
)

var OrganizationSet = wire.NewSet(
	orgRepo.NewPostgresRepository,
	orgRepo.NewAuthRepositoryAdapter,
	orgRepo.NewBillingRepositoryAdapter,
	orgSvc.NewService,
	orgHandler.NewHandler,
)

var ProjectSet = wire.NewSet(
	projectRepo.NewPostgresRepository,
	projectRepo.NewKubernetesRepository,
	projectSvc.NewService,
	projectHandler.NewHandler,
)

var WorkspaceSet = wire.NewSet(
	workspaceRepo.NewPostgresRepository,
	workspaceRepo.NewKubernetesRepository,
	workspaceRepo.NewAuthRepositoryAdapter,
	workspaceSvc.NewService,
	workspaceHandler.NewHandler,
)

// Legacy wire sets for packages that haven't been migrated yet
var BackupSet = wire.NewSet(
	backupRepo.NewPostgresRepository, 
	ProvideBackupProxmoxRepository, 
	ProvideBackupService,
	handlers.NewBackupHandler,
)

var BillingSet = wire.NewSet(
	billingRepo.NewPostgresRepository,
	ProvideStripeRepository,
	billingSvc.NewService,
	handlers.NewBillingHandler,
)

var MonitoringSet = wire.NewSet(
	monitoringRepo.NewPostgresRepository,
	k8sRepo.NewKubernetesRepository,
	monitoringSvc.NewService,
	handlers.NewMonitoringHandler,
)

var NodeSet = wire.NewSet(
	nodeRepo.NewPostgresRepository,
	ProvideNodeRepository,
	ProvideProxmoxRepository,
	ProvideProxmoxRepositoryInterface,
	nodeSvc.NewService,
	ProvideNodeService,
	handlers.NewNodeHandler,
)

var CICDSet = wire.NewSet(
	cicdRepo.NewPostgresRepository,
	ProvideCICDProviderFactory,
	ProvideCICDCredentialManager,
	cicdSvc.NewService,
	handlers.NewCICDHandler,
)

var FunctionSet = wire.NewSet(
	ProvideSQLDB,
	functionRepo.NewPostgresRepository,
	ProvideFunctionRepository,
	ProvideFunctionProviderFactory,
	functionSvc.NewService,
	ProvideFunctionService,
	handlers.NewFunctionHandler,
)

var HelmSet = wire.NewSet(helm.NewService)

var AIOpsProxySet = wire.NewSet(
	ProvideAIOpsProxyHandler,
)

var AIOpsSet = wire.NewSet(
	aiopsRepo.NewPostgresRepository,
	ProvideOllamaService,
	aiopsSvc.NewService,
)

var LogSet = wire.NewSet(
	ProvideClickHouseConnection,
	logRepo.NewClickHouseRepository,
	logSvc.NewLogService,
)

var InternalSet = wire.NewSet(ProvideInternalHandler)

type App struct {
	ApplicationHandler  *applicationHandler.ApplicationHandler
	AuthHandler        *authHandler.Handler
	BackupHandler      *handlers.BackupHandler
	BillingHandler     *handlers.BillingHandler
	MonitoringHandler  *handlers.MonitoringHandler
	NodeHandler        *handlers.NodeHandler
	OrganizationHandler *orgHandler.Handler
	ProjectHandler     *projectHandler.Handler
	WorkspaceHandler   *workspaceHandler.Handler
	CICDHandler        *handlers.CICDHandler
	FunctionHandler    *handlers.FunctionHandler
	AIOpsProxyHandler  *handlers.AIOpsProxyHandler
	InternalHandler    *handlers.InternalHandler
}

func NewApp(
	appH *applicationHandler.ApplicationHandler,
	authH *authHandler.Handler,
	backupH *handlers.BackupHandler,
	billH *handlers.BillingHandler,
	monH *handlers.MonitoringHandler,
	nodeH *handlers.NodeHandler,
	orgH *orgHandler.Handler,
	projH *projectHandler.Handler,
	workH *workspaceHandler.Handler,
	cicdH *handlers.CICDHandler,
	funcH *handlers.FunctionHandler,
	aiopsH *handlers.AIOpsProxyHandler,
	internalHandler *handlers.InternalHandler,
) *App {
	return &App{
		ApplicationHandler:  appH,
		AuthHandler:        authH,
		BackupHandler:      backupH,
		BillingHandler:     billH,
		MonitoringHandler:  monH,
		NodeHandler:        nodeH,
		OrganizationHandler: orgH,
		ProjectHandler:     projH,
		WorkspaceHandler:   workH,
		CICDHandler:        cicdH,
		FunctionHandler:    funcH,
		AIOpsProxyHandler:  aiopsH,
		InternalHandler:    internalHandler,
	}
}

type StripeAPIKey string
type StripeWebhookSecret string
type AIOpsServiceURL string
type CICDNamespace string
type BackupEncryptionKey string

func ProvideOAuthProviderConfigs(cfg *config.Config) map[string]*authRepo.ProviderConfig {
	providers := make(map[string]*authRepo.ProviderConfig)
	if cfg.Auth.ExternalProviders == nil {
		return providers
	}
	for name, p := range cfg.Auth.ExternalProviders {
		providers[name] = &authRepo.ProviderConfig{
			ClientID:     p.ClientID,
			ClientSecret: p.ClientSecret,
			RedirectURL:  p.RedirectURL,
			Scopes:       p.Scopes,
			AuthURL:      p.AuthURL,
			TokenURL:     p.TokenURL,
		}
	}
	return providers
}

func ProvideStripeAPIKey(cfg *config.Config) StripeAPIKey { return StripeAPIKey(cfg.Stripe.APIKey) }
func ProvideStripeWebhookSecret(cfg *config.Config) StripeWebhookSecret { return StripeWebhookSecret(cfg.Stripe.WebhookSecret) }
func ProvideAIOpsServiceURL(cfg *config.Config) (AIOpsServiceURL, error) { 
	if cfg.AIOps.URL != "" { 
		return AIOpsServiceURL(cfg.AIOps.URL), nil 
	}
	return AIOpsServiceURL("http://ai-ops-service.ai-ops.svc.cluster.local:8000"), nil 
}
func ProvideStripeRepository(apiKey StripeAPIKey, webhookSecret StripeWebhookSecret) billing.StripeRepository { return billingRepo.NewStripeRepository(string(apiKey), string(webhookSecret)) }
func ProvideCICDNamespace() CICDNamespace { return CICDNamespace("hexabase-cicd") }
func ProvideCICDProviderFactory(kubeClient kubernetes.Interface, k8sConfig *rest.Config, namespace CICDNamespace) cicd.ProviderFactory { return cicdRepo.NewProviderFactory(kubeClient, k8sConfig, string(namespace)) }
func ProvideCICDCredentialManager(kubeClient kubernetes.Interface, namespace CICDNamespace) cicd.CredentialManager { return cicdRepo.NewKubernetesCredentialManager(kubeClient, string(namespace)) }
func ProvideFunctionProviderFactory(kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) function.ProviderFactory {
	return functionRepo.NewProviderFactory(kubeClient, dynamicClient)
}
func ProvideFunctionService(service *functionSvc.Service) function.Service {
	return service
}
func ProvideSQLDB(gormDB *gorm.DB) (*sql.DB, error) {
	return gormDB.DB()
}
func ProvideFunctionRepository(repo *functionRepo.PostgresRepository) function.Repository {
	return repo
}
func ProvideClickHouseConnection(cfg *config.Config) (clickhouse.Conn, error) {
	// This should be expanded with full config options (user, pass, etc.)
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.ClickHouse.Address},
	})
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func ProvideProxmoxRepository(cfg *config.Config) *nodeRepo.ProxmoxRepository {
	// TODO: Get Proxmox configuration from config
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	return nodeRepo.NewProxmoxRepository(httpClient, "https://proxmox.example.com/api2/json", "your-api-token")
}

func ProvideMetricsClientset(k8sConfig *rest.Config) (versioned.Interface, error) {
	return versioned.NewForConfig(k8sConfig)
}

func ProvideNodeService(svc *nodeSvc.Service) node.Service {
	return svc
}

func ProvideNodeRepository(repo *nodeRepo.PostgresRepository) node.Repository {
	return repo
}

func ProvideProxmoxRepositoryInterface(repo *nodeRepo.ProxmoxRepository) node.ProxmoxRepository {
	return repo
}

func ProvideBackupProxmoxRepository(cfg *config.Config) backup.ProxmoxRepository {
	// Reuse the same Proxmox connection settings
	// TODO: Get from config
	client := proxmox.NewClient("https://proxmox.example.com/api2/json", "root@pam", "tokenID", "tokenSecret")
	return backupRepo.NewProxmoxRepository(client)
}

func ProvideBackupService(
	repo backup.Repository,
	proxmoxRepo backup.ProxmoxRepository,
	appRepo domain.Repository,
	workspaceRepo workspaceDomain.Repository,
	k8sClient kubernetes.Interface,
	cfg *config.Config,
) backup.Service {
	// TODO: Get encryption key from config
	encryptionKey := "your-backup-encryption-key"
	return backupSvc.NewService(repo, proxmoxRepo, appRepo, workspaceRepo, k8sClient, encryptionKey)
}

func ProvideAIOpsProxyHandler(authSvc authDomain.Service, logger *slog.Logger, cfg *config.Config) (*handlers.AIOpsProxyHandler, error) {
	var aiopsURL string
	if cfg.AIOps.URL != "" {
		aiopsURL = cfg.AIOps.URL
	} else {
		aiopsURL = "http://ai-ops-service.ai-ops.svc.cluster.local:8000"
	}
	return handlers.NewAIOpsProxyHandler(authSvc, logger, aiopsURL)
}

func ProvideOllamaService(cfg *config.Config) aiops.LLMService {
	// TODO: Get Ollama configuration from config
	ollamaURL := "http://ollama.ollama.svc.cluster.local:11434"
	timeout := 30 * time.Second
	headers := make(map[string]string)
	return aiopsRepo.NewOllamaProvider(ollamaURL, timeout, headers)
}

func ProvideInternalHandler(
	workspaceSvc workspaceDomain.Service,
	projectSvc projectDomain.Service,
	applicationSvc domain.Service,
	nodeSvc node.Service,
	logSvc logs.Service,
	monitoringSvc monitoring.Service,
	aiopsSvc aiops.Service,
	cicdSvc cicd.Service,
	backupSvc backup.Service,
	logger *slog.Logger,
) *handlers.InternalHandler {
	return handlers.NewInternalHandler(
		workspaceSvc,
		projectSvc,
		applicationSvc,
		nodeSvc,
		logSvc,
		monitoringSvc,
		aiopsSvc,
		cicdSvc,
		backupSvc,
		logger,
	)
}

func InitializeApp(cfg *config.Config, db *gorm.DB, k8sClient kubernetes.Interface, dynamicClient dynamic.Interface, k8sConfig *rest.Config, logger *slog.Logger) (*App, error) {
	wire.Build(
		ApplicationSet,
		AuthSet,
		BackupSet,
		BillingSet,
		MonitoringSet,
		NodeSet,
		OrganizationSet,
		ProjectSet,
		WorkspaceSet,
		CICDSet,
		FunctionSet,
		HelmSet,
		AIOpsProxySet,
		AIOpsSet,
		LogSet,
		InternalSet,
		ProvideOAuthProviderConfigs,
		ProvideStripeAPIKey,
		ProvideStripeWebhookSecret,
		ProvideCICDNamespace,
		ProvideMetricsClientset,
		NewApp,
	)
	return nil, nil
}