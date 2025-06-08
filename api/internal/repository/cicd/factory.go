package cicd

import (
	"fmt"

	"github.com/hexabase/hexabase-ai/api/internal/cicd/tekton"
	"github.com/hexabase/hexabase-ai/api/internal/domain/cicd"
	tektonclient "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// ProviderFactory implements the provider factory interface
type ProviderFactory struct {
	kubeClient kubernetes.Interface
	k8sConfig  *rest.Config
	namespace  string
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory(kubeClient kubernetes.Interface, k8sConfig *rest.Config, namespace string) cicd.ProviderFactory {
	return &ProviderFactory{
		kubeClient: kubeClient,
		k8sConfig:  k8sConfig,
		namespace:  namespace,
	}
}

// CreateProvider creates a provider instance with the given configuration
func (f *ProviderFactory) CreateProvider(providerType string, config *cicd.ProviderConfig) (cicd.Provider, error) {
	switch providerType {
	case "tekton":
		return f.createTektonProvider(config)
	case "github-actions":
		return nil, fmt.Errorf("github-actions provider not yet implemented")
	case "gitlab-ci":
		return nil, fmt.Errorf("gitlab-ci provider not yet implemented")
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}

// ListProviders returns available provider types
func (f *ProviderFactory) ListProviders() []string {
	return []string{
		"tekton",
		"github-actions",
		"gitlab-ci",
	}
}

// createTektonProvider creates a Tekton provider instance
func (f *ProviderFactory) createTektonProvider(config *cicd.ProviderConfig) (cicd.Provider, error) {
	// Create Tekton client
	tektonClient, err := tektonclient.NewForConfig(f.k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Tekton client: %w", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(f.k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return tekton.NewTektonProvider(tektonClient, f.kubeClient, dynamicClient), nil
}