package api

import (
	"github.com/hexabase/kaas-api/internal/config"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handlers holds all HTTP handlers and their dependencies
type Handlers struct {
	db     *gorm.DB
	config *config.Config
	logger *zap.Logger
	
	// Handler groups
	Auth          *AuthHandler
	Organizations *OrganizationHandler
	Workspaces    *WorkspaceHandler
	Projects      *ProjectHandler
	Groups        *GroupHandler
	Webhooks      *WebhookHandler
}

// NewHandlers creates a new handlers instance
func NewHandlers(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *Handlers {
	h := &Handlers{
		db:     db,
		config: cfg,
		logger: logger,
	}

	// Initialize handler groups
	h.Auth = NewAuthHandler(db, cfg, logger)
	h.Organizations = NewOrganizationHandler(db, cfg, logger)
	h.Workspaces = NewWorkspaceHandler(db, cfg, logger)
	h.Projects = NewProjectHandler(db, cfg, logger)
	h.Groups = NewGroupHandler(db, cfg, logger)
	h.Webhooks = NewWebhookHandler(db, cfg, logger)

	return h
}