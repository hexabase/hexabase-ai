package backup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hexabase/hexabase-ai/api/internal/db"
	"github.com/hexabase/hexabase-ai/api/internal/domain/backup"
	"gorm.io/gorm"
)

// PostgresRepository implements the backup repository using PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository creates a new PostgreSQL backup repository
func NewPostgresRepository(database *gorm.DB) backup.Repository {
	return &PostgresRepository{
		db: database,
	}
}

// BackupStorage operations

func (r *PostgresRepository) CreateBackupStorage(ctx context.Context, storage *backup.BackupStorage) error {
	dbStorage := r.domainToDBStorage(storage)
	if err := r.db.WithContext(ctx).Create(dbStorage).Error; err != nil {
		return fmt.Errorf("failed to create backup storage: %w", err)
	}
	*storage = *r.dbToDomainStorage(dbStorage)
	return nil
}

func (r *PostgresRepository) GetBackupStorage(ctx context.Context, storageID string) (*backup.BackupStorage, error) {
	var dbStorage db.BackupStorage
	if err := r.db.WithContext(ctx).Where("id = ?", storageID).First(&dbStorage).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("backup storage not found")
		}
		return nil, fmt.Errorf("failed to get backup storage: %w", err)
	}
	return r.dbToDomainStorage(&dbStorage), nil
}

func (r *PostgresRepository) GetBackupStorageByName(ctx context.Context, workspaceID, name string) (*backup.BackupStorage, error) {
	var dbStorage db.BackupStorage
	if err := r.db.WithContext(ctx).Where("workspace_id = ? AND name = ?", workspaceID, name).First(&dbStorage).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get backup storage by name: %w", err)
	}
	return r.dbToDomainStorage(&dbStorage), nil
}

func (r *PostgresRepository) ListBackupStorages(ctx context.Context, workspaceID string) ([]backup.BackupStorage, error) {
	var dbStorages []db.BackupStorage
	if err := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).Find(&dbStorages).Error; err != nil {
		return nil, fmt.Errorf("failed to list backup storages: %w", err)
	}
	
	storages := make([]backup.BackupStorage, len(dbStorages))
	for i, s := range dbStorages {
		storages[i] = *r.dbToDomainStorage(&s)
	}
	return storages, nil
}

func (r *PostgresRepository) UpdateBackupStorage(ctx context.Context, storage *backup.BackupStorage) error {
	dbStorage := r.domainToDBStorage(storage)
	if err := r.db.WithContext(ctx).Save(dbStorage).Error; err != nil {
		return fmt.Errorf("failed to update backup storage: %w", err)
	}
	return nil
}

func (r *PostgresRepository) DeleteBackupStorage(ctx context.Context, storageID string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", storageID).Delete(&db.BackupStorage{}).Error; err != nil {
		return fmt.Errorf("failed to delete backup storage: %w", err)
	}
	return nil
}

func (r *PostgresRepository) UpdateStorageUsage(ctx context.Context, storageID string, usedGB int) error {
	if err := r.db.WithContext(ctx).Model(&db.BackupStorage{}).
		Where("id = ?", storageID).
		Update("used_gb", usedGB).Error; err != nil {
		return fmt.Errorf("failed to update storage usage: %w", err)
	}
	return nil
}

// BackupPolicy operations

func (r *PostgresRepository) CreateBackupPolicy(ctx context.Context, policy *backup.BackupPolicy) error {
	dbPolicy := r.domainToDBPolicy(policy)
	if err := r.db.WithContext(ctx).Create(dbPolicy).Error; err != nil {
		return fmt.Errorf("failed to create backup policy: %w", err)
	}
	*policy = *r.dbToDomainPolicy(dbPolicy)
	return nil
}

func (r *PostgresRepository) GetBackupPolicy(ctx context.Context, policyID string) (*backup.BackupPolicy, error) {
	var dbPolicy db.BackupPolicy
	if err := r.db.WithContext(ctx).Where("id = ?", policyID).First(&dbPolicy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("backup policy not found")
		}
		return nil, fmt.Errorf("failed to get backup policy: %w", err)
	}
	return r.dbToDomainPolicy(&dbPolicy), nil
}

func (r *PostgresRepository) GetBackupPolicyByApplication(ctx context.Context, applicationID string) (*backup.BackupPolicy, error) {
	var dbPolicy db.BackupPolicy
	if err := r.db.WithContext(ctx).Where("application_id = ?", applicationID).First(&dbPolicy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get backup policy by application: %w", err)
	}
	return r.dbToDomainPolicy(&dbPolicy), nil
}

func (r *PostgresRepository) ListBackupPolicies(ctx context.Context, workspaceID string) ([]backup.BackupPolicy, error) {
	var dbPolicies []db.BackupPolicy
	if err := r.db.WithContext(ctx).
		Joins("JOIN applications ON backup_policies.application_id = applications.id").
		Where("applications.workspace_id = ?", workspaceID).
		Find(&dbPolicies).Error; err != nil {
		return nil, fmt.Errorf("failed to list backup policies: %w", err)
	}
	
	policies := make([]backup.BackupPolicy, len(dbPolicies))
	for i, p := range dbPolicies {
		policies[i] = *r.dbToDomainPolicy(&p)
	}
	return policies, nil
}

func (r *PostgresRepository) UpdateBackupPolicy(ctx context.Context, policy *backup.BackupPolicy) error {
	dbPolicy := r.domainToDBPolicy(policy)
	if err := r.db.WithContext(ctx).Save(dbPolicy).Error; err != nil {
		return fmt.Errorf("failed to update backup policy: %w", err)
	}
	return nil
}

func (r *PostgresRepository) DeleteBackupPolicy(ctx context.Context, policyID string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", policyID).Delete(&db.BackupPolicy{}).Error; err != nil {
		return fmt.Errorf("failed to delete backup policy: %w", err)
	}
	return nil
}

func (r *PostgresRepository) ListEnabledPolicies(ctx context.Context) ([]backup.BackupPolicy, error) {
	var dbPolicies []db.BackupPolicy
	if err := r.db.WithContext(ctx).Where("enabled = ?", true).Find(&dbPolicies).Error; err != nil {
		return nil, fmt.Errorf("failed to list enabled policies: %w", err)
	}
	
	policies := make([]backup.BackupPolicy, len(dbPolicies))
	for i, p := range dbPolicies {
		policies[i] = *r.dbToDomainPolicy(&p)
	}
	return policies, nil
}

// BackupExecution operations

func (r *PostgresRepository) CreateBackupExecution(ctx context.Context, execution *backup.BackupExecution) error {
	dbExecution := r.domainToDBExecution(execution)
	if err := r.db.WithContext(ctx).Create(dbExecution).Error; err != nil {
		return fmt.Errorf("failed to create backup execution: %w", err)
	}
	*execution = *r.dbToDomainExecution(dbExecution)
	return nil
}

func (r *PostgresRepository) GetBackupExecution(ctx context.Context, executionID string) (*backup.BackupExecution, error) {
	var dbExecution db.BackupExecution
	if err := r.db.WithContext(ctx).Where("id = ?", executionID).First(&dbExecution).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("backup execution not found")
		}
		return nil, fmt.Errorf("failed to get backup execution: %w", err)
	}
	return r.dbToDomainExecution(&dbExecution), nil
}

func (r *PostgresRepository) ListBackupExecutions(ctx context.Context, policyID string, limit, offset int) ([]backup.BackupExecution, int, error) {
	var total int64
	var dbExecutions []db.BackupExecution
	
	query := r.db.WithContext(ctx).Where("policy_id = ?", policyID)
	
	// Count total
	if err := query.Model(&db.BackupExecution{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count backup executions: %w", err)
	}
	
	// Get paginated results
	if err := query.Order("started_at DESC").Limit(limit).Offset(offset).Find(&dbExecutions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list backup executions: %w", err)
	}
	
	executions := make([]backup.BackupExecution, len(dbExecutions))
	for i, e := range dbExecutions {
		executions[i] = *r.dbToDomainExecution(&e)
	}
	
	return executions, int(total), nil
}

func (r *PostgresRepository) UpdateBackupExecution(ctx context.Context, execution *backup.BackupExecution) error {
	dbExecution := r.domainToDBExecution(execution)
	if err := r.db.WithContext(ctx).Save(dbExecution).Error; err != nil {
		return fmt.Errorf("failed to update backup execution: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetLatestBackupExecution(ctx context.Context, policyID string) (*backup.BackupExecution, error) {
	var dbExecution db.BackupExecution
	if err := r.db.WithContext(ctx).
		Where("policy_id = ? AND status = ?", policyID, "succeeded").
		Order("started_at DESC").
		First(&dbExecution).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest backup execution: %w", err)
	}
	return r.dbToDomainExecution(&dbExecution), nil
}

func (r *PostgresRepository) GetBackupExecutionsByApplication(ctx context.Context, applicationID string, limit, offset int) ([]backup.BackupExecution, int, error) {
	var total int64
	var dbExecutions []db.BackupExecution
	
	query := r.db.WithContext(ctx).
		Joins("JOIN backup_policies ON backup_executions.policy_id = backup_policies.id").
		Where("backup_policies.application_id = ?", applicationID)
	
	// Count total
	if err := query.Model(&db.BackupExecution{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count backup executions: %w", err)
	}
	
	// Get paginated results
	if err := query.Order("backup_executions.started_at DESC").
		Limit(limit).Offset(offset).Find(&dbExecutions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list backup executions by application: %w", err)
	}
	
	executions := make([]backup.BackupExecution, len(dbExecutions))
	for i, e := range dbExecutions {
		executions[i] = *r.dbToDomainExecution(&e)
	}
	
	return executions, int(total), nil
}

func (r *PostgresRepository) CleanupOldBackups(ctx context.Context, policyID string, retentionDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	
	// Delete old backup executions
	if err := r.db.WithContext(ctx).
		Where("policy_id = ? AND started_at < ? AND status = ?", policyID, cutoffDate, "succeeded").
		Delete(&db.BackupExecution{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup old backups: %w", err)
	}
	
	return nil
}

// BackupRestore operations

func (r *PostgresRepository) CreateBackupRestore(ctx context.Context, restore *backup.BackupRestore) error {
	dbRestore := r.domainToDBRestore(restore)
	if err := r.db.WithContext(ctx).Create(dbRestore).Error; err != nil {
		return fmt.Errorf("failed to create backup restore: %w", err)
	}
	*restore = *r.dbToDomainRestore(dbRestore)
	return nil
}

func (r *PostgresRepository) GetBackupRestore(ctx context.Context, restoreID string) (*backup.BackupRestore, error) {
	var dbRestore db.BackupRestore
	if err := r.db.WithContext(ctx).Where("id = ?", restoreID).First(&dbRestore).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("backup restore not found")
		}
		return nil, fmt.Errorf("failed to get backup restore: %w", err)
	}
	return r.dbToDomainRestore(&dbRestore), nil
}

func (r *PostgresRepository) ListBackupRestores(ctx context.Context, applicationID string, limit, offset int) ([]backup.BackupRestore, int, error) {
	var total int64
	var dbRestores []db.BackupRestore
	
	query := r.db.WithContext(ctx).Where("application_id = ?", applicationID)
	
	// Count total
	if err := query.Model(&db.BackupRestore{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count backup restores: %w", err)
	}
	
	// Get paginated results
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&dbRestores).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list backup restores: %w", err)
	}
	
	restores := make([]backup.BackupRestore, len(dbRestores))
	for i := range dbRestores {
		dbRestore := dbRestores[i]
		restores[i] = *r.dbToDomainRestore(&dbRestore)
	}
	
	return restores, int(total), nil
}

func (r *PostgresRepository) UpdateBackupRestore(ctx context.Context, restore *backup.BackupRestore) error {
	dbRestore := r.domainToDBRestore(restore)
	if err := r.db.WithContext(ctx).Save(dbRestore).Error; err != nil {
		return fmt.Errorf("failed to update backup restore: %w", err)
	}
	return nil
}

// Storage usage operations

func (r *PostgresRepository) GetStorageUsage(ctx context.Context, storageID string) (*backup.BackupStorageUsage, error) {
	var storage db.BackupStorage
	if err := r.db.WithContext(ctx).Where("id = ?", storageID).First(&storage).Error; err != nil {
		return nil, fmt.Errorf("failed to get storage: %w", err)
	}
	
	// Get backup count
	var backupCount int64
	r.db.WithContext(ctx).Model(&db.BackupExecution{}).
		Joins("JOIN backup_policies ON backup_executions.policy_id = backup_policies.id").
		Where("backup_policies.storage_id = ? AND backup_executions.status = ?", storageID, "succeeded").
		Count(&backupCount)
	
	// Get oldest and latest backup
	var oldestBackup, latestBackup time.Time
	r.db.WithContext(ctx).Model(&db.BackupExecution{}).
		Joins("JOIN backup_policies ON backup_executions.policy_id = backup_policies.id").
		Where("backup_policies.storage_id = ? AND backup_executions.status = ?", storageID, "succeeded").
		Order("started_at ASC").
		Limit(1).
		Pluck("started_at", &oldestBackup)
	
	r.db.WithContext(ctx).Model(&db.BackupExecution{}).
		Joins("JOIN backup_policies ON backup_executions.policy_id = backup_policies.id").
		Where("backup_policies.storage_id = ? AND backup_executions.status = ?", storageID, "succeeded").
		Order("started_at DESC").
		Limit(1).
		Pluck("started_at", &latestBackup)
	
	usage := &backup.BackupStorageUsage{
		StorageID:    storage.ID,
		TotalGB:      storage.CapacityGB,
		UsedGB:       storage.UsedGB,
		AvailableGB:  storage.CapacityGB - storage.UsedGB,
		UsagePercent: float64(storage.UsedGB) / float64(storage.CapacityGB) * 100,
		BackupCount:  int(backupCount),
	}
	
	if !oldestBackup.IsZero() {
		usage.OldestBackup = &oldestBackup
	}
	if !latestBackup.IsZero() {
		usage.LatestBackup = &latestBackup
	}
	
	return usage, nil
}

func (r *PostgresRepository) GetWorkspaceStorageUsage(ctx context.Context, workspaceID string) ([]backup.BackupStorageUsage, error) {
	var storages []db.BackupStorage
	if err := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).Find(&storages).Error; err != nil {
		return nil, fmt.Errorf("failed to get workspace storages: %w", err)
	}
	
	usages := make([]backup.BackupStorageUsage, len(storages))
	for i, storage := range storages {
		usage, err := r.GetStorageUsage(ctx, storage.ID)
		if err != nil {
			return nil, err
		}
		usages[i] = *usage
	}
	
	return usages, nil
}

// Helper functions for converting between domain and database models

func (r *PostgresRepository) domainToDBStorage(storage *backup.BackupStorage) *db.BackupStorage {
	return &db.BackupStorage{
		ID:               storage.ID,
		WorkspaceID:      storage.WorkspaceID,
		Name:             storage.Name,
		Type:             string(storage.Type),
		ProxmoxStorageID: storage.ProxmoxStorageID,
		ProxmoxNodeID:    storage.ProxmoxNodeID,
		CapacityGB:       storage.CapacityGB,
		UsedGB:           storage.UsedGB,
		Status:           string(storage.Status),
		ConnectionConfig: mapToJSON(storage.ConnectionConfig),
		ErrorMessage:     storage.ErrorMessage,
		CreatedAt:        storage.CreatedAt,
		UpdatedAt:        storage.UpdatedAt,
	}
}

func (r *PostgresRepository) dbToDomainStorage(dbStorage *db.BackupStorage) *backup.BackupStorage {
	return &backup.BackupStorage{
		ID:               dbStorage.ID,
		WorkspaceID:      dbStorage.WorkspaceID,
		Name:             dbStorage.Name,
		Type:             backup.StorageType(dbStorage.Type),
		ProxmoxStorageID: dbStorage.ProxmoxStorageID,
		ProxmoxNodeID:    dbStorage.ProxmoxNodeID,
		CapacityGB:       dbStorage.CapacityGB,
		UsedGB:           dbStorage.UsedGB,
		Status:           backup.StorageStatus(dbStorage.Status),
		ConnectionConfig: jsonToMap(dbStorage.ConnectionConfig),
		ErrorMessage:     dbStorage.ErrorMessage,
		CreatedAt:        dbStorage.CreatedAt,
		UpdatedAt:        dbStorage.UpdatedAt,
	}
}

func (r *PostgresRepository) domainToDBPolicy(policy *backup.BackupPolicy) *db.BackupPolicy {
	return &db.BackupPolicy{
		ID:                 policy.ID,
		ApplicationID:      policy.ApplicationID,
		StorageID:          policy.StorageID,
		Enabled:            policy.Enabled,
		Schedule:           policy.Schedule,
		RetentionDays:      policy.RetentionDays,
		BackupType:         string(policy.BackupType),
		IncludeVolumes:     policy.IncludeVolumes,
		IncludeDatabase:    policy.IncludeDatabase,
		IncludeConfig:      policy.IncludeConfig,
		CompressionEnabled: policy.CompressionEnabled,
		EncryptionEnabled:  policy.EncryptionEnabled,
		EncryptionKeyRef:   policy.EncryptionKeyRef,
		PreBackupHook:      policy.PreBackupHook,
		PostBackupHook:     policy.PostBackupHook,
		CreatedAt:          policy.CreatedAt,
		UpdatedAt:          policy.UpdatedAt,
	}
}

func (r *PostgresRepository) dbToDomainPolicy(dbPolicy *db.BackupPolicy) *backup.BackupPolicy {
	return &backup.BackupPolicy{
		ID:                 dbPolicy.ID,
		ApplicationID:      dbPolicy.ApplicationID,
		StorageID:          dbPolicy.StorageID,
		Enabled:            dbPolicy.Enabled,
		Schedule:           dbPolicy.Schedule,
		RetentionDays:      dbPolicy.RetentionDays,
		BackupType:         backup.BackupType(dbPolicy.BackupType),
		IncludeVolumes:     dbPolicy.IncludeVolumes,
		IncludeDatabase:    dbPolicy.IncludeDatabase,
		IncludeConfig:      dbPolicy.IncludeConfig,
		CompressionEnabled: dbPolicy.CompressionEnabled,
		EncryptionEnabled:  dbPolicy.EncryptionEnabled,
		EncryptionKeyRef:   dbPolicy.EncryptionKeyRef,
		PreBackupHook:      dbPolicy.PreBackupHook,
		PostBackupHook:     dbPolicy.PostBackupHook,
		CreatedAt:          dbPolicy.CreatedAt,
		UpdatedAt:          dbPolicy.UpdatedAt,
	}
}

func (r *PostgresRepository) domainToDBExecution(execution *backup.BackupExecution) *db.BackupExecution {
	var cronJobExecutionID *string
	if execution.CronJobExecutionID != "" {
		cronJobExecutionID = &execution.CronJobExecutionID
	}
	
	return &db.BackupExecution{
		ID:                  execution.ID,
		PolicyID:            execution.PolicyID,
		CronJobExecutionID:  cronJobExecutionID,
		Status:              string(execution.Status),
		SizeBytes:           execution.SizeBytes,
		CompressedSizeBytes: execution.CompressedSizeBytes,
		BackupPath:          execution.BackupPath,
		BackupManifest:      mapToJSON(execution.BackupManifest),
		StartedAt:           execution.StartedAt,
		CompletedAt:         execution.CompletedAt,
		ErrorMessage:        execution.ErrorMessage,
		Metadata:            mapToJSON(execution.Metadata),
		CreatedAt:           execution.CreatedAt,
	}
}

func (r *PostgresRepository) dbToDomainExecution(dbExecution *db.BackupExecution) *backup.BackupExecution {
	cronJobExecutionID := ""
	if dbExecution.CronJobExecutionID != nil {
		cronJobExecutionID = *dbExecution.CronJobExecutionID
	}
	
	return &backup.BackupExecution{
		ID:                  dbExecution.ID,
		PolicyID:            dbExecution.PolicyID,
		CronJobExecutionID:  cronJobExecutionID,
		Status:              backup.BackupExecutionStatus(dbExecution.Status),
		SizeBytes:           dbExecution.SizeBytes,
		CompressedSizeBytes: dbExecution.CompressedSizeBytes,
		BackupPath:          dbExecution.BackupPath,
		BackupManifest:      jsonToMap(dbExecution.BackupManifest),
		StartedAt:           dbExecution.StartedAt,
		CompletedAt:         dbExecution.CompletedAt,
		ErrorMessage:        dbExecution.ErrorMessage,
		Metadata:            jsonToMap(dbExecution.Metadata),
		CreatedAt:           dbExecution.CreatedAt,
	}
}

func (r *PostgresRepository) domainToDBRestore(restore *backup.BackupRestore) *db.BackupRestore {
	var newApplicationID *string
	if restore.NewApplicationID != "" {
		newApplicationID = &restore.NewApplicationID
	}
	
	return &db.BackupRestore{
		ID:                restore.ID,
		BackupExecutionID: restore.BackupExecutionID,
		ApplicationID:     restore.ApplicationID,
		Status:            string(restore.Status),
		RestoreType:       string(restore.RestoreType),
		RestoreOptions:    mapToJSON(restore.RestoreOptions),
		NewApplicationID:  newApplicationID,
		StartedAt:         restore.StartedAt,
		CompletedAt:       restore.CompletedAt,
		ErrorMessage:      restore.ErrorMessage,
		ValidationResults: mapToJSON(restore.ValidationResults),
		CreatedAt:         restore.CreatedAt,
	}
}

func (r *PostgresRepository) dbToDomainRestore(dbRestore *db.BackupRestore) *backup.BackupRestore {
	newApplicationID := ""
	if dbRestore.NewApplicationID != nil {
		newApplicationID = *dbRestore.NewApplicationID
	}
	
	return &backup.BackupRestore{
		ID:                dbRestore.ID,
		BackupExecutionID: dbRestore.BackupExecutionID,
		ApplicationID:     dbRestore.ApplicationID,
		Status:            backup.RestoreStatus(dbRestore.Status),
		RestoreType:       backup.RestoreType(dbRestore.RestoreType),
		RestoreOptions:    jsonToMap(dbRestore.RestoreOptions),
		NewApplicationID:  newApplicationID,
		StartedAt:         dbRestore.StartedAt,
		CompletedAt:       dbRestore.CompletedAt,
		ErrorMessage:      dbRestore.ErrorMessage,
		ValidationResults: jsonToMap(dbRestore.ValidationResults),
		CreatedAt:         dbRestore.CreatedAt,
	}
}

// Helper functions for JSON conversion
func mapToJSON(m map[string]interface{}) db.JSON {
	if m == nil {
		return db.JSON("null")
	}
	b, err := json.Marshal(m)
	if err != nil {
		return db.JSON("{}")
	}
	return db.JSON(b)
}

func jsonToMap(j db.JSON) map[string]interface{} {
	if len(j) == 0 || string(j) == "null" {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(j), &m); err != nil {
		return nil
	}
	return m
}