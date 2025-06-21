package repository

import (
	"time"
	"github.com/hexabase/hexabase-ai/api/internal/organization/domain"
)

// dbOrganization represents the database model for organizations
type dbOrganization struct {
	ID                   string    `gorm:"primaryKey"`
	Name                 string    `gorm:"not null"`
	StripeCustomerID     *string   `gorm:"unique"`
	StripeSubscriptionID *string   `gorm:"unique"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (dbOrganization) TableName() string {
	return "organizations"
}

// dbOrganizationUser represents the database model for organization users
type dbOrganizationUser struct {
	OrganizationID string    `gorm:"primaryKey"`
	UserID         string    `gorm:"primaryKey"`
	Role           string    `gorm:"not null;default:'member';check:role IN ('admin','member')"`
	JoinedAt       time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func (dbOrganizationUser) TableName() string {
	return "organization_users"
}

// Conversion functions

func domainToDBOrganization(org *domain.Organization) *dbOrganization {
	return &dbOrganization{
		ID:        org.ID,
		Name:      org.Name,
		CreatedAt: org.CreatedAt,
		UpdatedAt: org.UpdatedAt,
	}
}

func dbToDomainOrganization(dbOrg *dbOrganization) *domain.Organization {
	return &domain.Organization{
		ID:          dbOrg.ID,
		Name:        dbOrg.Name,
		DisplayName: dbOrg.Name, // Use name as display name
		Status:      "active",   // Default status
		CreatedAt:   dbOrg.CreatedAt,
		UpdatedAt:   dbOrg.UpdatedAt,
	}
}

func domainToDBOrganizationUser(member *domain.OrganizationUser) *dbOrganizationUser {
	return &dbOrganizationUser{
		OrganizationID: member.OrganizationID,
		UserID:         member.UserID,
		Role:           member.Role,
		JoinedAt:       member.JoinedAt,
	}
}

func dbToDomainOrganizationUser(dbMember *dbOrganizationUser) *domain.OrganizationUser {
	return &domain.OrganizationUser{
		OrganizationID: dbMember.OrganizationID,
		UserID:         dbMember.UserID,
		Role:           dbMember.Role,
		JoinedAt:       dbMember.JoinedAt,
		Status:         "active", // Default status
		InvitedAt:      dbMember.JoinedAt, // Use joined_at as invited_at
	}
}