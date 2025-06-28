package repository

import (
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/organization/domain"
)

// dbOrganization represents the database model for organizations
type dbOrganization struct {
	ID                   string     `gorm:"primaryKey"`
	Name                 string     `gorm:"not null"`
	DisplayName          string     `gorm:"column:display_name"`
	Description          string     `gorm:"column:description"`
	Website              string     `gorm:"column:website"`
	Email                string     `gorm:"column:email"`
	Status               string     `gorm:"column:status;default:'active'"`
	OwnerID              string     `gorm:"column:owner_id"`
	StripeCustomerID     *string    `gorm:"unique"`
	StripeSubscriptionID *string    `gorm:"unique"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            *time.Time `gorm:"column:deleted_at"`
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
		ID:                   org.ID,
		Name:                 org.Name,
		DisplayName:          org.DisplayName,
		Description:          org.Description,
		Website:              org.Website,
		Email:                org.Email,
		Status:               org.Status,
		OwnerID:              org.OwnerID,
		CreatedAt:            org.CreatedAt,
		UpdatedAt:            org.UpdatedAt,
		DeletedAt:            org.DeletedAt,
	}
}

func dbToDomainOrganization(dbOrg *dbOrganization) *domain.Organization {
	return &domain.Organization{
		ID:          dbOrg.ID,
		Name:        dbOrg.Name,
		DisplayName: dbOrg.DisplayName,
		Description: dbOrg.Description,
		Website:     dbOrg.Website,
		Email:       dbOrg.Email,
		Status:      dbOrg.Status,
		OwnerID:     dbOrg.OwnerID,
		CreatedAt:   dbOrg.CreatedAt,
		UpdatedAt:   dbOrg.UpdatedAt,
		DeletedAt:   dbOrg.DeletedAt,
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