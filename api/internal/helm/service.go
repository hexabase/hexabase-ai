// Package helm provides a service for programmatic Helm operations.
package helm

import (
	"fmt"
	"log/slog"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/client-go/rest"
)

// Service defines the interface for Helm operations.
type Service interface {
	InstallOrUpgrade(releaseName, chartPath, namespace string, values map[string]interface{}) error
}

// helmService implements the Service interface.
type helmService struct {
	cfg    *rest.Config
	logger *slog.Logger
}

// NewService creates a new Helm service.
func NewService(cfg *rest.Config, logger *slog.Logger) Service {
	return &helmService{
		cfg:    cfg,
		logger: logger,
	}
}

// InstallOrUpgrade performs a Helm install or upgrade for a given release.
func (s *helmService) InstallOrUpgrade(releaseName, chartPath, namespace string, values map[string]interface{}) error {
	actionConfig := new(action.Configuration)
	// We need a debug function for the Helm client.
	debugLog := func(format string, v ...interface{}) {
		s.logger.Debug("helm client", "message", fmt.Sprintf(format, v...))
	}

	// Initialize the action configuration with proper Kubernetes config.
	// For now, we'll use the RESTClientGetter from the environment.
	// In production, you'd want to create a proper RESTClientGetter from s.cfg
	// TODO: Implement proper RESTClientGetter from rest.Config
	if err := actionConfig.Init(nil, namespace, os.Getenv("HELM_DRIVER"), debugLog); err != nil {
		s.logger.Error("failed to initialize Helm action config", "error", err)
		return err
	}
	
	// Check if the release already exists.
	histClient := action.NewHistory(actionConfig)
	histClient.Max = 1
	if _, err := histClient.Run(releaseName); err == nil {
		// Release exists, so we'll upgrade.
		s.logger.Info("release exists, upgrading chart", "release", releaseName)
		upgrade := action.NewUpgrade(actionConfig)
		upgrade.Namespace = namespace
		upgrade.Install = true // If release is not found, install it
		upgrade.MaxHistory = 5
		
		chart, err := loader.Load(chartPath)
		if err != nil {
			s.logger.Error("failed to load chart for upgrade", "path", chartPath, "error", err)
			return err
		}
		
		_, err = upgrade.Run(releaseName, chart, values)
		if err != nil {
			s.logger.Error("helm upgrade failed", "release", releaseName, "error", err)
			return err
		}
		s.logger.Info("helm upgrade successful", "release", releaseName)
	} else {
		// Release does not exist, so we'll install.
		s.logger.Info("release does not exist, installing chart", "release", releaseName)
		install := action.NewInstall(actionConfig)
		install.ReleaseName = releaseName
		install.Namespace = namespace
		install.CreateNamespace = true // Create the namespace if it doesn't exist
		
		chart, err := loader.Load(chartPath)
		if err != nil {
			s.logger.Error("failed to load chart for install", "path", chartPath, "error", err)
			return err
		}

		_, err = install.Run(chart, values)
		if err != nil {
			s.logger.Error("helm install failed", "release", releaseName, "error", err)
			return err
		}
		s.logger.Info("helm install successful", "release", releaseName)
	}

	return nil
} 