package domain

import "errors"

// Sentinel errors for organization operations
var (
	// ErrOrganizationNotFound is returned when an organization is not found
	ErrOrganizationNotFound = errors.New("organization not found")
)
