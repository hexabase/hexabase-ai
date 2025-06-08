//go:build wireinject
// +build wireinject

package wire

import (
	"log/slog"

	"github.com/google/wire"
	"github.com/hexabase/hexabase-ai/api/internal/api/handlers"
	"github.com/hexabase/hexabase-ai/api/internal/config"
	"github.com/hexabase/hexabase-ai/api/internal/domain/billing"
	"github.com/hexabase/hexabase-ai/api/internal/domain/cicd"
	"github.com/hexabase/hexabase-ai/api/internal/helm"
	authRepo "github.com/hexabase/hexabase-ai/api/internal/repository/auth"
	billingRepo "github.com/hexabase/hexabase-ai/api/internal/repository/billing"
	cicdRepo "github.com/hexabase/hexabase-ai/api/internal/repository/cicd"
	k8sRepo "github.com/hexabase/hexabase-ai/api/internal/repository/kubernetes"
	monitoringRepo "github.com/hexabase/hexabase-ai/api/internal/repository/monitoring"
	orgRepo "github.com/hexabase/hexabase-ai/api/internal/repository/organization"
	projectRepo "github.com/hexabase/hexabase-ai/api/internal/repository/project"
	workspaceRepo "github.com/hexabase/hexabase-ai/api/internal/repository/workspace"
	authSvc "github.com/hexabase/hexabase-ai/api/internal/service/auth"
	billingSvc "github.com/hexabase/hexabase-ai/api/internal/service/billing"
	cicdSvc "github.com/hexabase/hexabase-ai/api/internal/service/cicd"
	monitoringSvc "github.com/hexabase/hexabase-ai/api/internal/service/monitoring"
	orgSvc "github.com/hexabase/hexabase-ai/api/internal/service/organization"
	projectSvc "github.com/hexabase/hexabase-ai/api/internal/service/project"
	workspaceSvc "github.com/hexabase/hexabase-ai/api/internal/service/workspace"
	"gorm.io/gorm"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var AuthSet = wire.NewSet(authRepo.NewPostgresRepository, authRepo.NewOAuthRepository, authRepo.NewKeyRepository, authSvc.NewService, handlers.NewAuthHandler)
var BillingSet = wire.NewSet(billingRepo.NewPostgresRepository, ProvideStripeRepository, billingSvc.NewService, handlers.NewBillingHandler)
var MonitoringSet = wire.NewSet(monitoringRepo.NewPostgresRepository, k8sRepo.NewKubernetesRepository, monitoringSvc.NewService, handlers.NewMonitoringHandler)
var OrganizationSet = wire.NewSet(orgRepo.NewPostgresRepository, orgRepo.NewAuthRepositoryAdapter, orgRepo.NewBillingRepositoryAdapter, orgSvc.NewService, handlers.NewOrganizationHandler)
var ProjectSet = wire.NewSet(projectRepo.NewPostgresRepository, projectRepo.NewKubernetesRepository, projectSvc.NewService, handlers.NewProjectHandler)
var WorkspaceSet = wire.NewSet(workspaceRepo.NewPostgresRepository, workspaceRepo.NewKubernetesRepository, workspaceRepo.NewAuthRepositoryAdapter, workspaceSvc.NewService, handlers.NewWorkspaceHandler)
var CICDSet = wire.NewSet(cicdRepo.NewPostgresRepository, ProvideCICDProviderFactory, ProvideCICDCredentialManager, cicdSvc.NewService, handlers.NewCICDHandler)
var HelmSet = wire.NewSet(helm.NewService)
var AIOpsProxySet = wire.NewSet(ProvideAIOpsServiceURL, handlers.NewAIOpsProxyHandler)

type App struct {
	AuthHandler *handlers.AuthHandler; BillingHandler *handlers.BillingHandler; MonitoringHandler *handlers.MonitoringHandler; OrganizationHandler *handlers.OrganizationHandler; ProjectHandler *handlers.ProjectHandler; WorkspaceHandler *handlers.WorkspaceHandler; CICDHandler *handlers.CICDHandler; AIOpsProxyHandler *handlers.AIOpsProxyHandler
}

func NewApp(authH *handlers.AuthHandler, billH *handlers.BillingHandler, monH *handlers.MonitoringHandler, orgH *handlers.OrganizationHandler, projH *handlers.ProjectHandler, workH *handlers.WorkspaceHandler, cicdH *handlers.CICDHandler, aiopsH *handlers.AIOpsProxyHandler) *App {
	return &App{AuthHandler: authH, BillingHandler: billH, MonitoringHandler: monH, OrganizationHandler: orgH, ProjectHandler: projH, WorkspaceHandler: workH, CICDHandler: cicdH, AIOpsProxyHandler: aiopsH}
}

type StripeAPIKey string
type StripeWebhookSecret string
type AIOpsServiceURL string
type CICDNamespace string

func ProvideOAuthProviderConfigs(cfg *config.Config) map[string]*authRepo.ProviderConfig {
	// TODO: Load from config
	return make(map[string]*authRepo.ProviderConfig)
}

func ProvideStripeAPIKey(cfg *config.Config) StripeAPIKey { return StripeAPIKey(cfg.Stripe.APIKey) }
func ProvideStripeWebhookSecret(cfg *config.Config) StripeWebhookSecret { return StripeWebhookSecret(cfg.Stripe.WebhookSecret) }
func ProvideAIOpsServiceURL(cfg *config.Config) (string, error) { if cfg.AIOps.URL != "" { return cfg.AIOps.URL, nil }; return "http://ai-ops-service.ai-ops.svc.cluster.local:8000", nil }
func ProvideStripeRepository(apiKey StripeAPIKey, webhookSecret StripeWebhookSecret) billing.StripeRepository { return billingRepo.NewStripeRepository(string(apiKey), string(webhookSecret)) }
func ProvideCICDNamespace() CICDNamespace { return CICDNamespace("hexabase-cicd") }
func ProvideCICDProviderFactory(kubeClient kubernetes.Interface, k8sConfig *rest.Config, namespace CICDNamespace) cicd.ProviderFactory { return cicdRepo.NewProviderFactory(kubeClient, k8sConfig, string(namespace)) }
func ProvideCICDCredentialManager(kubeClient kubernetes.Interface, namespace CICDNamespace) cicd.CredentialManager { return cicdRepo.NewKubernetesCredentialManager(kubeClient, string(namespace)) }

func InitializeApp(cfg *config.Config, db *gorm.DB, k8sClient kubernetes.Interface, dynamicClient dynamic.Interface, k8sConfig *rest.Config, logger *slog.Logger) (*App, error) {
	wire.Build(
		AuthSet,
		BillingSet,
		MonitoringSet,
		OrganizationSet,
		ProjectSet,
		WorkspaceSet,
		CICDSet,
		HelmSet,
		AIOpsProxySet,
		ProvideOAuthProviderConfigs,
		ProvideStripeAPIKey,
		ProvideStripeWebhookSecret,
		ProvideCICDNamespace,
		NewApp,
	)
	return nil, nil
}