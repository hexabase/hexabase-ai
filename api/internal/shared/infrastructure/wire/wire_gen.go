// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package wire

import (
	"context"
	"database/sql"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/wire"
	domain7 "github.com/hexabase/hexabase-ai/api/internal/aiops/domain"
	handler12 "github.com/hexabase/hexabase-ai/api/internal/aiops/handler"
	repository12 "github.com/hexabase/hexabase-ai/api/internal/aiops/repository"
	service12 "github.com/hexabase/hexabase-ai/api/internal/aiops/service"
	domain11 "github.com/hexabase/hexabase-ai/api/internal/application/domain"
	"github.com/hexabase/hexabase-ai/api/internal/application/handler"
	"github.com/hexabase/hexabase-ai/api/internal/application/repository"
	"github.com/hexabase/hexabase-ai/api/internal/application/service"
	"github.com/hexabase/hexabase-ai/api/internal/auth"
	domain8 "github.com/hexabase/hexabase-ai/api/internal/auth/domain"
	handler2 "github.com/hexabase/hexabase-ai/api/internal/auth/handler"
	repository2 "github.com/hexabase/hexabase-ai/api/internal/auth/repository"
	service2 "github.com/hexabase/hexabase-ai/api/internal/auth/service"
	domain6 "github.com/hexabase/hexabase-ai/api/internal/backup/domain"
	handler3 "github.com/hexabase/hexabase-ai/api/internal/backup/handler"
	repository3 "github.com/hexabase/hexabase-ai/api/internal/backup/repository"
	service3 "github.com/hexabase/hexabase-ai/api/internal/backup/service"
	domain2 "github.com/hexabase/hexabase-ai/api/internal/billing/domain"
	handler4 "github.com/hexabase/hexabase-ai/api/internal/billing/handler"
	repository5 "github.com/hexabase/hexabase-ai/api/internal/billing/repository"
	service4 "github.com/hexabase/hexabase-ai/api/internal/billing/service"
	domain3 "github.com/hexabase/hexabase-ai/api/internal/cicd/domain"
	handler5 "github.com/hexabase/hexabase-ai/api/internal/cicd/handler"
	repository6 "github.com/hexabase/hexabase-ai/api/internal/cicd/repository"
	service5 "github.com/hexabase/hexabase-ai/api/internal/cicd/service"
	domain4 "github.com/hexabase/hexabase-ai/api/internal/function/domain"
	handler11 "github.com/hexabase/hexabase-ai/api/internal/function/handler"
	repository11 "github.com/hexabase/hexabase-ai/api/internal/function/repository"
	service11 "github.com/hexabase/hexabase-ai/api/internal/function/service"
	"github.com/hexabase/hexabase-ai/api/internal/helm"
	"github.com/hexabase/hexabase-ai/api/internal/logs/domain"
	repository13 "github.com/hexabase/hexabase-ai/api/internal/logs/repository"
	service13 "github.com/hexabase/hexabase-ai/api/internal/logs/service"
	domain12 "github.com/hexabase/hexabase-ai/api/internal/monitoring/domain"
	handler6 "github.com/hexabase/hexabase-ai/api/internal/monitoring/handler"
	repository7 "github.com/hexabase/hexabase-ai/api/internal/monitoring/repository"
	service6 "github.com/hexabase/hexabase-ai/api/internal/monitoring/service"
	domain5 "github.com/hexabase/hexabase-ai/api/internal/node/domain"
	handler7 "github.com/hexabase/hexabase-ai/api/internal/node/handler"
	repository8 "github.com/hexabase/hexabase-ai/api/internal/node/repository"
	"github.com/hexabase/hexabase-ai/api/internal/node/repository/proxmox"
	service7 "github.com/hexabase/hexabase-ai/api/internal/node/service"
	handler8 "github.com/hexabase/hexabase-ai/api/internal/organization/handler"
	repository9 "github.com/hexabase/hexabase-ai/api/internal/organization/repository"
	service8 "github.com/hexabase/hexabase-ai/api/internal/organization/service"
	domain10 "github.com/hexabase/hexabase-ai/api/internal/project/domain"
	handler9 "github.com/hexabase/hexabase-ai/api/internal/project/handler"
	repository10 "github.com/hexabase/hexabase-ai/api/internal/project/repository"
	service9 "github.com/hexabase/hexabase-ai/api/internal/project/service"
	"github.com/hexabase/hexabase-ai/api/internal/shared/api/handlers"
	"github.com/hexabase/hexabase-ai/api/internal/shared/config"
	kubernetes2 "github.com/hexabase/hexabase-ai/api/internal/shared/kubernetes/repository"
	"github.com/hexabase/hexabase-ai/api/internal/shared/redis"
	domain9 "github.com/hexabase/hexabase-ai/api/internal/workspace/domain"
	handler10 "github.com/hexabase/hexabase-ai/api/internal/workspace/handler"
	repository4 "github.com/hexabase/hexabase-ai/api/internal/workspace/repository"
	service10 "github.com/hexabase/hexabase-ai/api/internal/workspace/service"
	"gorm.io/gorm"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	"log/slog"
	"net/http"
	"time"
)

// Injectors from wire.go:

func InitializeApp(cfg *config.Config, db *gorm.DB, k8sClient kubernetes.Interface, dynamicClient dynamic.Interface, k8sConfig *rest.Config, logger *slog.Logger) (*App, error) {
	domainRepository := repository.NewPostgresRepository(db)
	versionedInterface, err := ProvideMetricsClientset(k8sConfig)
	if err != nil {
		return nil, err
	}
	kubernetesRepository := repository.NewKubernetesRepository(k8sClient, versionedInterface)
	domainService := service.NewService(domainRepository, kubernetesRepository)
	applicationHandler := handler.NewApplicationHandler(domainService)
	postgresRepository := repository2.NewPostgresRepository(db)
	client, err := ProvideRedisClient(cfg, logger)
	if err != nil {
		return nil, err
	}
	redisAuthRepository := repository2.NewRedisAuthRepository(client)
	tokenHashRepository := repository2.NewTokenHashRepository()
	repository14 := repository2.NewCompositeRepository(postgresRepository, redisAuthRepository, tokenHashRepository)
	v := ProvideOAuthProviderConfigs(cfg)
	oAuthRepository := repository2.NewOAuthRepository(v, logger)
	keyRepository, err := repository2.NewKeyRepository()
	if err != nil {
		return nil, err
	}
	tokenManager, err := ProvideTokenManager(keyRepository, cfg)
	if err != nil {
		return nil, err
	}
	tokenDomainService := ProvideTokenDomainService()
	sessionLimiterRepository := repository2.NewSessionLimiterRepository(client)
	sessionManager := service2.NewSessionManager(repository14, sessionLimiterRepository)
	int2 := ProvideDefaultTokenExpiry()
	service14 := service2.NewService(repository14, oAuthRepository, keyRepository, tokenManager, tokenDomainService, sessionManager, logger, int2)
	handlerHandler := handler2.NewHandler(service14, logger)
	repository15 := repository3.NewPostgresRepository(db)
	proxmoxRepository := ProvideBackupProxmoxRepository(cfg)
	repository16 := repository4.NewPostgresRepository(db)
	string2 := ProvideBackupEncryptionKey(cfg)
	service15 := service3.NewService(repository15, proxmoxRepository, domainRepository, repository16, k8sClient, string2)
	handler13 := handler3.NewHandler(service15)
	repository17 := repository5.NewPostgresRepository(db)
	stripeAPIKey := ProvideStripeAPIKey(cfg)
	stripeWebhookSecret := ProvideStripeWebhookSecret(cfg)
	stripeRepository := ProvideStripeRepository(stripeAPIKey, stripeWebhookSecret)
	service16 := service4.NewService(repository17, stripeRepository, logger)
	handler14 := handler4.NewHandler(service16, logger)
	repository18 := repository6.NewPostgresRepository(db)
	cicdNamespace := ProvideCICDNamespace()
	providerFactory := ProvideCICDProviderFactory(k8sClient, k8sConfig, cicdNamespace)
	credentialManager := ProvideCICDCredentialManager(k8sClient, cicdNamespace)
	service17 := service5.NewService(repository18, providerFactory, credentialManager, logger)
	handler15 := handler5.NewHandler(service17, logger)
	repository19 := repository7.NewPostgresRepository(db)
	repository20 := kubernetes2.NewKubernetesRepository(k8sClient)
	service18 := service6.NewService(repository19, repository20, logger)
	handler16 := handler6.NewHandler(service18, logger)
	repositoryPostgresRepository := repository8.NewPostgresRepository(db)
	repository21 := ProvideNodeRepository(repositoryPostgresRepository)
	repositoryProxmoxRepository := ProvideProxmoxRepository(cfg)
	domainProxmoxRepository := ProvideProxmoxRepositoryInterface(repositoryProxmoxRepository)
	serviceService := service7.NewService(repository21, domainProxmoxRepository)
	service19 := ProvideNodeService(serviceService)
	handler17 := handler7.NewHandler(service19, logger)
	repository22 := repository9.NewPostgresRepository(db)
	authRepository := repository9.NewAuthRepositoryAdapter(repository14)
	billingRepository := repository9.NewBillingRepositoryAdapter(stripeRepository)
	service20 := service8.NewService(repository22, authRepository, billingRepository, logger)
	handler18 := handler8.NewHandler(service20, logger)
	repository23 := repository10.NewPostgresRepository(db)
	domainKubernetesRepository := repository10.NewKubernetesRepository(k8sClient, dynamicClient, k8sConfig)
	service21 := service9.NewService(repository23, domainKubernetesRepository, logger)
	handler19 := handler9.NewHandler(service21, logger)
	kubernetesRepository2 := repository4.NewKubernetesRepository(k8sClient, dynamicClient, k8sConfig)
	domainAuthRepository := repository4.NewAuthRepositoryAdapter(repository14)
	helmService := helm.NewService(k8sConfig, logger)
	service22 := service10.NewService(repository16, kubernetesRepository2, domainAuthRepository, helmService, logger)
	handler20 := handler10.NewHandler(service22, logger)
	sqlDB, err := ProvideSQLDB(db)
	if err != nil {
		return nil, err
	}
	postgresRepository2 := repository11.NewPostgresRepository(sqlDB)
	repository24 := ProvideFunctionRepository(postgresRepository2)
	domainProviderFactory := ProvideFunctionProviderFactory(k8sClient, dynamicClient)
	service23 := service11.NewService(repository24, domainProviderFactory, logger)
	service24 := ProvideFunctionService(service23)
	handler21 := handler11.NewHandler(service24, logger)
	llmService := ProvideOllamaService(cfg)
	repository25 := repository12.NewPostgresRepository(db, logger)
	service25 := service12.NewService(llmService, repository25, logger)
	aiOpsService := ProvideAIOpsAdapter(service25)
	handler22 := handler12.NewHandler(aiOpsService, logger)
	ginHandler := handler12.NewGinHandler(service25, logger)
	aiOpsServiceURL, err := ProvideAIOpsServiceURL(cfg)
	if err != nil {
		return nil, err
	}
	aiOpsProxyHandler, err := ProvideAIOpsProxyHandler(service14, logger, aiOpsServiceURL)
	if err != nil {
		return nil, err
	}
	v2, err := ProvideClickHouseConn(cfg)
	if err != nil {
		return nil, err
	}
	repository26 := repository13.NewClickHouseRepository(v2, logger)
	service26 := service13.NewLogService(repository26, logger)
	internalHandler := ProvideInternalHandler(service22, service21, domainService, service19, service26, service18, service25, service17, service15, logger)
	app := NewApp(applicationHandler, handlerHandler, handler13, handler14, handler15, handler16, handler17, handler18, handler19, handler20, handler21, handler22, ginHandler, aiOpsProxyHandler, internalHandler, service26)
	return app, nil
}

// wire.go:

// Updated wire sets for migrated packages
var ApplicationSet = wire.NewSet(repository.NewPostgresRepository, repository.NewKubernetesRepository, service.NewService, handler.NewApplicationHandler)

var AuthSet = wire.NewSet(
	ProvideRedisClient, repository2.NewPostgresRepository, repository2.NewRedisAuthRepository, repository2.NewTokenHashRepository, repository2.NewCompositeRepository, repository2.NewOAuthRepository, repository2.NewKeyRepository, repository2.NewSessionLimiterRepository, ProvideTokenManager,
	ProvideTokenDomainService,
	ProvideDefaultTokenExpiry, service2.NewSessionManager, service2.NewService, handler2.NewHandler,
)

var OrganizationSet = wire.NewSet(repository9.NewPostgresRepository, repository9.NewAuthRepositoryAdapter, repository9.NewBillingRepositoryAdapter, service8.NewService, handler8.NewHandler)

var ProjectSet = wire.NewSet(repository10.NewPostgresRepository, repository10.NewKubernetesRepository, service9.NewService, handler9.NewHandler)

var WorkspaceSet = wire.NewSet(repository4.NewPostgresRepository, repository4.NewKubernetesRepository, repository4.NewAuthRepositoryAdapter, service10.NewService, handler10.NewHandler)

// Legacy wire sets for packages that haven't been migrated yet
var BackupSet = wire.NewSet(repository3.NewPostgresRepository, ProvideBackupProxmoxRepository,
	ProvideBackupEncryptionKey, service3.NewService, handler3.NewHandler,
)

var BillingSet = wire.NewSet(repository5.NewPostgresRepository, ProvideStripeRepository, service4.NewService, handler4.NewHandler)

var MonitoringSet = wire.NewSet(repository7.NewPostgresRepository, kubernetes2.NewKubernetesRepository, service6.NewService, handler6.NewHandler)

var NodeSet = wire.NewSet(repository8.NewPostgresRepository, ProvideNodeRepository,
	ProvideProxmoxRepository,
	ProvideProxmoxRepositoryInterface, service7.NewService, ProvideNodeService, handler7.NewHandler,
)

var CICDSet = wire.NewSet(repository6.NewPostgresRepository, ProvideCICDProviderFactory,
	ProvideCICDCredentialManager, service5.NewService, handler5.NewHandler,
)

var FunctionSet = wire.NewSet(
	ProvideSQLDB, repository11.NewPostgresRepository, ProvideFunctionRepository, repository11.NewProviderFactory, ProvideFunctionProviderFactory, service11.NewService, ProvideFunctionService, handler11.NewHandler,
)

var HelmSet = wire.NewSet(helm.NewService)

var AIOpsSet = wire.NewSet(repository12.NewPostgresRepository, ProvideOllamaService, service12.NewService, ProvideAIOpsAdapter, handler12.NewHandler, handler12.NewGinHandler, ProvideAIOpsServiceURL,
	ProvideAIOpsProxyHandler,
)

var LogSet = wire.NewSet(
	ProvideClickHouseConn, repository13.NewClickHouseRepository, service13.NewLogService,
)

var InternalSet = wire.NewSet(ProvideInternalHandler)

type App struct {
	ApplicationHandler  *handler.ApplicationHandler
	AuthHandler         *handler2.Handler
	BackupHandler       *handler3.Handler
	BillingHandler      *handler4.Handler
	CICDHandler         *handler5.Handler
	MonitoringHandler   *handler6.Handler
	NodeHandler         *handler7.Handler
	OrganizationHandler *handler8.Handler
	ProjectHandler      *handler9.Handler
	WorkspaceHandler    *handler10.Handler
	FunctionHandler     *handler11.Handler
	AIOpsHandler        *handler12.Handler
	AIOpsGinHandler     *handler12.GinHandler
	AIOpsProxyHandler   *handler12.AIOpsProxyHandler
	InternalHandler     *handlers.InternalHandler
	LogSvc              domain.Service
}

func NewApp(
	appH *handler.ApplicationHandler,
	authH *handler2.Handler,
	backupH *handler3.Handler,
	billH *handler4.Handler,
	cicdH *handler5.Handler,
	monH *handler6.Handler,
	nodeH *handler7.Handler,
	orgH *handler8.Handler,
	projH *handler9.Handler,
	workH *handler10.Handler,
	funcH *handler11.Handler,
	aiopsH *handler12.Handler,
	aiopsGinH *handler12.GinHandler,
	aiopsProxyH *handler12.AIOpsProxyHandler,
	internalHandler *handlers.InternalHandler,
	logSvc domain.Service,
) *App {
	return &App{
		ApplicationHandler:  appH,
		AuthHandler:         authH,
		BackupHandler:       backupH,
		BillingHandler:      billH,
		CICDHandler:         cicdH,
		MonitoringHandler:   monH,
		NodeHandler:         nodeH,
		OrganizationHandler: orgH,
		ProjectHandler:      projH,
		WorkspaceHandler:    workH,
		FunctionHandler:     funcH,
		AIOpsHandler:        aiopsH,
		AIOpsGinHandler:     aiopsGinH,
		AIOpsProxyHandler:   aiopsProxyH,
		InternalHandler:     internalHandler,
		LogSvc:              logSvc,
	}
}

type StripeAPIKey string

type StripeWebhookSecret string

type AIOpsServiceURL string

type CICDNamespace string

type BackupEncryptionKey string

func ProvideOAuthProviderConfigs(cfg *config.Config) map[string]*repository2.ProviderConfig {
	providers := make(map[string]*repository2.ProviderConfig)
	if cfg.Auth.ExternalProviders == nil {
		return providers
	}
	for name, p := range cfg.Auth.ExternalProviders {
		providers[name] = &repository2.ProviderConfig{
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

func ProvideStripeWebhookSecret(cfg *config.Config) StripeWebhookSecret {
	return StripeWebhookSecret(cfg.Stripe.WebhookSecret)
}

func ProvideAIOpsServiceURL(cfg *config.Config) (AIOpsServiceURL, error) {
	if cfg.AIOps.URL != "" {
		return AIOpsServiceURL(cfg.AIOps.URL), nil
	}
	return AIOpsServiceURL("http://ai-ops-service.ai-ops.svc.cluster.local:8000"), nil
}

func ProvideStripeRepository(apiKey StripeAPIKey, webhookSecret StripeWebhookSecret) domain2.StripeRepository {
	return repository5.NewStripeRepository(string(apiKey), string(webhookSecret))
}

func ProvideCICDNamespace() CICDNamespace { return CICDNamespace("hexabase-cicd") }

func ProvideCICDProviderFactory(kubeClient kubernetes.Interface, k8sConfig *rest.Config, namespace CICDNamespace) domain3.ProviderFactory {
	return repository6.NewProviderFactory(kubeClient, k8sConfig, string(namespace))
}

func ProvideCICDCredentialManager(kubeClient kubernetes.Interface, namespace CICDNamespace) domain3.CredentialManager {
	return repository6.NewKubernetesCredentialManager(kubeClient, string(namespace))
}

func ProvideFunctionProviderFactory(kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) domain4.ProviderFactory {
	return repository11.NewProviderFactory(kubeClient, dynamicClient)
}

func ProvideFunctionService(service14 *service11.Service) domain4.Service {
	return service14
}

func ProvideSQLDB(gormDB *gorm.DB) (*sql.DB, error) {
	return gormDB.DB()
}

func ProvideFunctionRepository(repo *repository11.PostgresRepository) domain4.Repository {
	return repo
}

func ProvideClickHouseConn(cfg *config.Config) (clickhouse.Conn, error) {

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.ClickHouse.Address},
	})
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func ProvideProxmoxRepository(cfg *config.Config) *repository8.ProxmoxRepository {

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	return repository8.NewProxmoxRepository(httpClient, "https://proxmox.example.com/api2/json", "your-api-token")
}

func ProvideMetricsClientset(k8sConfig *rest.Config) (versioned.Interface, error) {
	return versioned.NewForConfig(k8sConfig)
}

func ProvideNodeService(svc *service7.Service) domain5.Service {
	return svc
}

func ProvideNodeRepository(repo *repository8.PostgresRepository) domain5.Repository {
	return repo
}

func ProvideProxmoxRepositoryInterface(repo *repository8.ProxmoxRepository) domain5.ProxmoxRepository {
	return repo
}

func ProvideBackupProxmoxRepository(cfg *config.Config) domain6.ProxmoxRepository {

	client := proxmox.NewClient("https://proxmox.example.com/api2/json", "root@pam", "tokenID", "tokenSecret")
	return repository3.NewProxmoxRepository(client)
}

func ProvideBackupEncryptionKey(cfg *config.Config) string {

	return "your-backup-encryption-key"
}

func ProvideOllamaService(cfg *config.Config) domain7.LLMService {

	ollamaURL := "http://ollama.ollama.svc.cluster.local:11434"
	timeout := 30 * time.Second
	headers := make(map[string]string)
	return repository12.NewOllamaProvider(ollamaURL, timeout, headers)
}

func ProvideAIOpsProxyHandler(authSvc domain8.Service, logger *slog.Logger, aiopsURL AIOpsServiceURL) (*handler12.AIOpsProxyHandler, error) {
	return handler12.NewAIOpsProxyHandler(authSvc, logger, string(aiopsURL))
}

// ProvideAIOpsAdapter provides an adapter from domain.Service to AIOpsService interface
func ProvideAIOpsAdapter(service14 domain7.Service) handler12.AIOpsService {
	return &aiopsServiceAdapter{service: service14}
}

// aiopsServiceAdapter adapts domain.Service to AIOpsService interface for backward compatibility
type aiopsServiceAdapter struct {
	service domain7.Service
}

func (a *aiopsServiceAdapter) CreateChatSession(workspaceID, userID, model string) (*domain7.ChatSession, error) {
	ctx := context.Background()
	return a.service.CreateChatSession(ctx, workspaceID, userID, "", model)
}

func (a *aiopsServiceAdapter) GetChatSession(sessionID string) (*domain7.ChatSession, error) {
	ctx := context.Background()
	return a.service.GetChatSession(ctx, sessionID)
}

func (a *aiopsServiceAdapter) ListChatSessions(workspaceID string, limit, offset int) ([]*domain7.ChatSession, error) {
	ctx := context.Background()
	return a.service.ListChatSessions(ctx, workspaceID, limit, offset)
}

func (a *aiopsServiceAdapter) DeleteChatSession(sessionID string) error {
	ctx := context.Background()
	return a.service.DeleteChatSession(ctx, sessionID)
}

func (a *aiopsServiceAdapter) Chat(sessionID string, message string, contextInts []int) (*domain7.ChatResponse, error) {
	ctx := context.Background()
	chatMessage := domain7.ChatMessage{
		Role:    "user",
		Content: message,
	}
	return a.service.SendMessage(ctx, sessionID, chatMessage)
}

func (a *aiopsServiceAdapter) StreamChat(sessionID string, message string, contextInts []int) (<-chan *domain7.ChatStreamResponse, error) {
	ctx := context.Background()
	chatMessage := domain7.ChatMessage{
		Role:    "user",
		Content: message,
	}
	return a.service.StreamMessage(ctx, sessionID, chatMessage)
}

func (a *aiopsServiceAdapter) GetAvailableModels() ([]*domain7.ModelInfo, error) {
	ctx := context.Background()
	return a.service.ListAvailableModels(ctx)
}

func (a *aiopsServiceAdapter) GetTokenUsage(workspaceID, model string, limit, offset int) ([]*domain7.ModelUsage, error) {

	return []*domain7.ModelUsage{}, nil
}

func ProvideInternalHandler(
	workspaceSvc domain9.Service,
	projectSvc domain10.Service,
	applicationSvc domain11.Service,
	nodeSvc domain5.Service,
	logSvc domain.Service,
	monitoringSvc domain12.Service,
	aiopsSvc domain7.Service,
	cicdSvc domain3.Service,
	backupSvc domain6.Service,
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

// ProvideTokenManager provides a TokenManager instance
func ProvideTokenManager(keyRepo domain8.KeyRepository, cfg *config.Config) (*auth.TokenManager, error) {
	privateKeyPEM, err := keyRepo.GetPrivateKey()
	if err != nil {
		return nil, err
	}

	publicKeyPEM, err := keyRepo.GetPublicKey()
	if err != nil {
		return nil, err
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return nil, err
	}

	return auth.NewTokenManager(privateKey, publicKey, "https://api.hexabase-kaas.io", time.Hour), nil
}

// ProvideTokenDomainService provides a TokenDomainService instance
func ProvideTokenDomainService() domain8.TokenDomainService {
	return domain8.NewTokenDomainService()
}

// ProvideDefaultTokenExpiry provides the default token expiry
func ProvideDefaultTokenExpiry() int {
	return 3600
}

// ProvideRedisClient provides a Redis client instance
func ProvideRedisClient(cfg *config.Config, logger *slog.Logger) (*redis.Client, error) {
	return redis.NewClient(&cfg.Redis, logger)
}
