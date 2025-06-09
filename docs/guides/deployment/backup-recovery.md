# Backup and Recovery Guide

Comprehensive guide for implementing backup and disaster recovery for Hexabase AI platform.

## Overview

The Hexabase AI backup strategy covers:
- **Application Data**: PostgreSQL databases, Redis cache
- **Kubernetes Resources**: Workspaces, configurations, secrets
- **Persistent Volumes**: User data, logs, metrics
- **Object Storage**: Uploaded files, artifacts
- **Disaster Recovery**: Full platform restoration procedures

## Backup Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   PostgreSQL    │────▶│     Velero      │────▶│  Object Storage │
│   Databases     │     │  (K8s Backup)   │     │   (S3/Minio)   │
└─────────────────┘     └─────────────────┘     └─────────────────┘
         │                       │                         │
         │              ┌─────────────────┐               │
         └─────────────▶│    WAL-G/PGX    │───────────────┘
                        │  (DB Streaming)  │
                        └─────────────────┘
```

## Backup Components

### 1. Database Backups

#### PostgreSQL with CloudNativePG

```yaml
# postgres-backup-config.yaml
apiVersion: v1
kind: Secret
metadata:
  name: backup-credentials
  namespace: hexabase-system
stringData:
  ACCESS_KEY_ID: "your-s3-access-key"
  SECRET_ACCESS_KEY: "your-s3-secret-key"
---
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: hexabase-db
  namespace: hexabase-system
spec:
  instances: 3
  
  postgresql:
    parameters:
      archive_mode: "on"
      archive_timeout: "5min"
      max_wal_size: "4GB"
      min_wal_size: "1GB"
  
  backup:
    # Retention policy
    retentionPolicy: "30d"
    
    # S3-compatible object store
    barmanObjectStore:
      destinationPath: "s3://hexabase-backups/postgres"
      endpointURL: "https://s3.amazonaws.com"
      s3Credentials:
        accessKeyId:
          name: backup-credentials
          key: ACCESS_KEY_ID
        secretAccessKey:
          name: backup-credentials
          key: SECRET_ACCESS_KEY
      
      # WAL archive configuration
      wal:
        compression: gzip
        encryption: AES256
        maxParallel: 8
      
      # Base backup configuration
      data:
        compression: gzip
        encryption: AES256
        immediateCheckpoint: false
        jobs: 4
```

#### Manual PostgreSQL Backup

```bash
#!/bin/bash
# backup-postgres.sh

# Variables
DB_HOST="hexabase-db-rw.hexabase-system.svc.cluster.local"
DB_NAME="hexabase"
DB_USER="hexabase"
BACKUP_DIR="/backups/postgres"
S3_BUCKET="s3://hexabase-backups/postgres/manual"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup
kubectl exec -n hexabase-system hexabase-db-1 -- \
  pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME \
  --format=custom \
  --verbose \
  --no-password \
  --compress=9 \
  > ${BACKUP_DIR}/hexabase_${DATE}.dump

# Encrypt backup
openssl enc -aes-256-cbc -salt \
  -in ${BACKUP_DIR}/hexabase_${DATE}.dump \
  -out ${BACKUP_DIR}/hexabase_${DATE}.dump.enc \
  -pass file:/etc/backup/encryption.key

# Upload to S3
aws s3 cp ${BACKUP_DIR}/hexabase_${DATE}.dump.enc \
  ${S3_BUCKET}/hexabase_${DATE}.dump.enc \
  --storage-class GLACIER_IR

# Cleanup old local backups
find ${BACKUP_DIR} -name "*.dump*" -mtime +7 -delete
```

### 2. Kubernetes Resource Backups

#### Install Velero

```bash
# Download Velero CLI
wget https://github.com/vmware-tanzu/velero/releases/download/v1.13.0/velero-v1.13.0-linux-amd64.tar.gz
tar -xvf velero-v1.13.0-linux-amd64.tar.gz
sudo mv velero-v1.13.0-linux-amd64/velero /usr/local/bin/

# Create S3 credentials
cat > credentials-velero <<EOF
[default]
aws_access_key_id=your-access-key
aws_secret_access_key=your-secret-key
EOF

# Install Velero in cluster
velero install \
  --provider aws \
  --plugins velero/velero-plugin-for-aws:v1.9.0 \
  --bucket hexabase-velero-backups \
  --secret-file ./credentials-velero \
  --backup-location-config \
    region=us-east-1,s3ForcePathStyle=false,s3Url=https://s3.amazonaws.com \
  --snapshot-location-config \
    region=us-east-1 \
  --use-node-agent \
  --default-volumes-to-fs-backup
```

#### Configure Backup Schedules

```bash
# Daily backup of control plane
velero schedule create control-plane-daily \
  --schedule="0 2 * * *" \
  --include-namespaces hexabase-system,hexabase-workspaces \
  --exclude-resources pods,events \
  --ttl 720h \
  --storage-location default

# Hourly backup of critical configs
velero schedule create configs-hourly \
  --schedule="0 * * * *" \
  --include-resources \
    configmaps,secrets,ingresses,services,deployments,statefulsets \
  --ttl 168h

# Weekly full cluster backup
velero schedule create full-weekly \
  --schedule="0 3 * * 0" \
  --ttl 2160h \
  --exclude-namespaces kube-system,kube-public,kube-node-lease
```

### 3. Persistent Volume Backups

#### Configure Volume Snapshots

```yaml
# volume-snapshot-class.yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotClass
metadata:
  name: csi-snapclass
driver: ebs.csi.aws.com
deletionPolicy: Retain
parameters:
  type: "gp3"
  encrypted: "true"
---
# Create volume snapshot schedule
apiVersion: batch/v1
kind: CronJob
metadata:
  name: volume-snapshot-scheduler
  namespace: hexabase-system
spec:
  schedule: "0 */6 * * *"  # Every 6 hours
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName: volume-snapshot-sa
          containers:
          - name: snapshot-creator
            image: hexabase/volume-snapshot-tool:latest
            command:
            - /bin/sh
            - -c
            - |
              # Get all PVCs
              for pvc in $(kubectl get pvc -A -o jsonpath='{range .items[*]}{.metadata.namespace}{" "}{.metadata.name}{"\n"}{end}'); do
                namespace=$(echo $pvc | awk '{print $1}')
                name=$(echo $pvc | awk '{print $2}')
                
                # Create snapshot
                kubectl apply -f - <<EOF
              apiVersion: snapshot.storage.k8s.io/v1
              kind: VolumeSnapshot
              metadata:
                name: ${name}-$(date +%Y%m%d-%H%M%S)
                namespace: ${namespace}
              spec:
                volumeSnapshotClassName: csi-snapclass
                source:
                  persistentVolumeClaimName: ${name}
              EOF
              done
              
              # Cleanup old snapshots (keep last 5)
              kubectl get volumesnapshot -A --sort-by=.metadata.creationTimestamp \
                | tail -n +6 | awk '{print $1" "$2}' \
                | xargs -n2 sh -c 'kubectl delete volumesnapshot -n $0 $1'
          restartPolicy: OnFailure
```

### 4. Application Data Backup

#### Redis Backup

```yaml
# redis-backup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: redis-backup
  namespace: hexabase-system
spec:
  schedule: "0 */4 * * *"  # Every 4 hours
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: redis-backup
            image: redis:7-alpine
            command:
            - /bin/sh
            - -c
            - |
              # Create backup
              redis-cli -h redis-master --rdb /backup/dump.rdb
              
              # Compress
              gzip -9 /backup/dump.rdb
              
              # Upload to S3
              aws s3 cp /backup/dump.rdb.gz \
                s3://hexabase-backups/redis/dump-$(date +%Y%m%d-%H%M%S).rdb.gz
            volumeMounts:
            - name: backup
              mountPath: /backup
            - name: aws-credentials
              mountPath: /root/.aws
          volumes:
          - name: backup
            emptyDir: {}
          - name: aws-credentials
            secret:
              secretName: aws-credentials
          restartPolicy: OnFailure
```

#### Object Storage Sync

```yaml
# minio-backup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: object-storage-backup
  namespace: hexabase-system
spec:
  schedule: "0 1 * * *"  # Daily at 1 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: mc-backup
            image: minio/mc:latest
            command:
            - /bin/sh
            - -c
            - |
              # Configure MinIO client
              mc alias set source https://minio.hexabase.local \
                $SOURCE_ACCESS_KEY $SOURCE_SECRET_KEY
              
              mc alias set backup s3 \
                $BACKUP_ACCESS_KEY $BACKUP_SECRET_KEY
              
              # Mirror buckets with versioning
              for bucket in $(mc ls source | awk '{print $5}'); do
                mc mirror source/$bucket backup/hexabase-backup-$bucket \
                  --overwrite --remove \
                  --exclude "*.tmp" \
                  --encrypt-key "backup/=32byteslongsecretkeymustbegiven"
              done
            env:
            - name: SOURCE_ACCESS_KEY
              valueFrom:
                secretKeyRef:
                  name: minio-credentials
                  key: access-key
            - name: SOURCE_SECRET_KEY
              valueFrom:
                secretKeyRef:
                  name: minio-credentials
                  key: secret-key
            - name: BACKUP_ACCESS_KEY
              valueFrom:
                secretKeyRef:
                  name: backup-s3-credentials
                  key: access-key
            - name: BACKUP_SECRET_KEY
              valueFrom:
                secretKeyRef:
                  name: backup-s3-credentials
                  key: secret-key
          restartPolicy: OnFailure
```

## Disaster Recovery Procedures

### 1. Recovery Planning

#### Recovery Time Objectives (RTO)

| Component | RTO | RPO | Priority |
|-----------|-----|-----|----------|
| Control Plane API | 15 min | 5 min | Critical |
| PostgreSQL Database | 30 min | 5 min | Critical |
| Workspace vClusters | 1 hour | 1 hour | High |
| User Data (PVs) | 2 hours | 6 hours | High |
| Monitoring Stack | 4 hours | 24 hours | Medium |
| Log Data | 8 hours | 24 hours | Low |

#### Disaster Recovery Runbook

```bash
#!/bin/bash
# disaster-recovery.sh

# 1. Verify backup availability
echo "=== Verifying Backups ==="
velero backup get
aws s3 ls s3://hexabase-backups/postgres/ --recursive | tail -10

# 2. Prepare new cluster
echo "=== Preparing Recovery Cluster ==="
# Assumes new K3s cluster is ready

# 3. Install Velero in recovery cluster
velero install \
  --provider aws \
  --plugins velero/velero-plugin-for-aws:v1.9.0 \
  --bucket hexabase-velero-backups \
  --secret-file ./credentials-velero \
  --backup-location-config region=us-east-1 \
  --snapshot-location-config region=us-east-1 \
  --use-node-agent \
  --wait

# 4. Restore control plane
echo "=== Restoring Control Plane ==="
LATEST_BACKUP=$(velero backup get --output json | jq -r '.items[0].metadata.name')
velero restore create control-plane-restore \
  --from-backup $LATEST_BACKUP \
  --include-namespaces hexabase-system \
  --wait

# 5. Restore database
echo "=== Restoring PostgreSQL ==="
kubectl apply -f postgres-cluster.yaml
kubectl wait --for=condition=Ready cluster/hexabase-db -n hexabase-system --timeout=600s

# Restore from backup
kubectl exec -n hexabase-system hexabase-db-1 -- \
  barman-cloud-restore \
    --cloud-provider aws-s3 \
    s3://hexabase-backups/postgres \
    $(kubectl get cluster hexabase-db -n hexabase-system -o jsonpath='{.status.targetPrimary}') \
    latest

# 6. Restore workspaces
echo "=== Restoring Workspaces ==="
velero restore create workspaces-restore \
  --from-backup $LATEST_BACKUP \
  --include-namespaces hexabase-workspaces \
  --wait

# 7. Verify restoration
echo "=== Verifying Restoration ==="
kubectl get pods -n hexabase-system
kubectl get vcluster -A
```

### 2. Database Recovery

#### Point-in-Time Recovery (PITR)

```bash
# Restore PostgreSQL to specific point in time
RECOVERY_TIME="2024-01-15 14:30:00"

# Stop current cluster
kubectl scale cluster hexabase-db -n hexabase-system --replicas=0

# Restore with PITR
kubectl exec -n hexabase-system hexabase-db-recovery -- \
  barman-cloud-restore \
    --cloud-provider aws-s3 \
    --endpoint-url https://s3.amazonaws.com \
    s3://hexabase-backups/postgres \
    hexabase-db \
    latest \
    --target-time "$RECOVERY_TIME"

# Update cluster to use recovered data
kubectl patch cluster hexabase-db -n hexabase-system --type merge -p \
  '{"spec":{"bootstrap":{"recovery":{"source":"hexabase-db-recovery"}}}}'

# Scale back up
kubectl scale cluster hexabase-db -n hexabase-system --replicas=3
```

### 3. Workspace Recovery

#### Individual Workspace Restoration

```bash
#!/bin/bash
# restore-workspace.sh

WORKSPACE_ID=$1
BACKUP_NAME=$2

# Find workspace resources in backup
velero backup describe $BACKUP_NAME --details | grep $WORKSPACE_ID

# Restore specific workspace
velero restore create workspace-$WORKSPACE_ID-restore \
  --from-backup $BACKUP_NAME \
  --selector "workspace-id=$WORKSPACE_ID" \
  --include-namespaces hexabase-workspaces \
  --wait

# Restore vCluster
kubectl apply -f - <<EOF
apiVersion: vcluster.loft.sh/v1alpha1
kind: VCluster
metadata:
  name: $WORKSPACE_ID
  namespace: hexabase-workspaces
spec:
  # Restored configuration
  restore:
    backup: $BACKUP_NAME
    workspace: $WORKSPACE_ID
EOF

# Wait for vCluster to be ready
kubectl wait --for=condition=Ready vcluster/$WORKSPACE_ID \
  -n hexabase-workspaces --timeout=300s

# Restore persistent volumes
for pvc in $(kubectl get pvc -n hexabase-workspaces -l workspace-id=$WORKSPACE_ID -o name); do
  SNAPSHOT=$(kubectl get volumesnapshot -n hexabase-workspaces \
    -l pvc-name=$(basename $pvc) \
    --sort-by=.metadata.creationTimestamp \
    -o jsonpath='{.items[-1].metadata.name}')
  
  kubectl patch $pvc -n hexabase-workspaces --type merge -p \
    '{"spec":{"dataSource":{"name":"'$SNAPSHOT'","kind":"VolumeSnapshot","apiGroup":"snapshot.storage.k8s.io"}}}'
done
```

### 4. Data Validation

#### Post-Recovery Validation Script

```bash
#!/bin/bash
# validate-recovery.sh

echo "=== Validating Recovery ==="

# 1. Check API health
API_HEALTH=$(curl -s https://api.hexabase.ai/health | jq -r '.status')
if [ "$API_HEALTH" != "healthy" ]; then
  echo "ERROR: API is not healthy"
  exit 1
fi

# 2. Validate database
DB_CHECK=$(kubectl exec -n hexabase-system hexabase-db-1 -- \
  psql -U hexabase -d hexabase -c "SELECT COUNT(*) FROM workspaces;" -t)
echo "Database contains $DB_CHECK workspaces"

# 3. Check vClusters
VCLUSTERS=$(kubectl get vcluster -A --no-headers | wc -l)
echo "Found $VCLUSTERS vClusters"

# 4. Test workspace connectivity
for vcluster in $(kubectl get vcluster -A -o jsonpath='{range .items[*]}{.metadata.namespace}/{.metadata.name} {end}'); do
  ns=$(echo $vcluster | cut -d/ -f1)
  name=$(echo $vcluster | cut -d/ -f2)
  
  kubectl --kubeconfig=/tmp/kubeconfig-$name \
    --context vcluster-$name \
    get nodes > /dev/null 2>&1
  
  if [ $? -eq 0 ]; then
    echo "✓ vCluster $name is accessible"
  else
    echo "✗ vCluster $name is NOT accessible"
  fi
done

# 5. Verify persistent data
kubectl get pv -o json | jq -r '.items[] | select(.status.phase=="Bound") | .metadata.name' | while read pv; do
  echo "Checking PV: $pv"
  # Add specific data validation based on PV content
done

echo "=== Recovery Validation Complete ==="
```

## Backup Monitoring

### 1. Backup Status Dashboard

```yaml
# grafana-backup-dashboard.json
{
  "dashboard": {
    "title": "Backup Status",
    "panels": [
      {
        "title": "Backup Success Rate",
        "targets": [{
          "expr": "rate(velero_backup_success_total[24h]) / rate(velero_backup_attempt_total[24h])"
        }]
      },
      {
        "title": "Last Successful Backups",
        "targets": [{
          "expr": "time() - velero_backup_last_successful_timestamp"
        }]
      },
      {
        "title": "Backup Size Trend",
        "targets": [{
          "expr": "velero_backup_size_bytes"
        }]
      },
      {
        "title": "PostgreSQL WAL Lag",
        "targets": [{
          "expr": "pg_replication_lag_seconds"
        }]
      }
    ]
  }
}
```

### 2. Backup Alerts

```yaml
# backup-alerts.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: backup-alerts
  namespace: monitoring
spec:
  groups:
  - name: backup
    rules:
    - alert: BackupFailed
      expr: increase(velero_backup_failure_total[1h]) > 0
      for: 5m
      labels:
        severity: critical
        team: platform
      annotations:
        summary: "Backup failed"
        description: "Velero backup {{ $labels.schedule }} has failed"
    
    - alert: BackupDelayed
      expr: time() - velero_backup_last_successful_timestamp > 86400
      for: 1h
      labels:
        severity: warning
        team: platform
      annotations:
        summary: "Backup delayed"
        description: "No successful backup for {{ $labels.schedule }} in 24 hours"
    
    - alert: PostgreSQLReplicationLag
      expr: pg_replication_lag_seconds > 300
      for: 10m
      labels:
        severity: critical
        team: platform
      annotations:
        summary: "PostgreSQL replication lag"
        description: "Replication lag is {{ $value }} seconds"
```

## Backup Testing

### 1. Automated Recovery Testing

```yaml
# backup-test-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: backup-recovery-test
  namespace: hexabase-system
spec:
  schedule: "0 4 * * 0"  # Weekly on Sunday
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName: backup-tester
          containers:
          - name: recovery-test
            image: hexabase/backup-tester:latest
            command:
            - /bin/sh
            - -c
            - |
              # Create test namespace
              kubectl create namespace backup-test-$(date +%Y%m%d)
              
              # Restore latest backup to test namespace
              LATEST_BACKUP=$(velero backup get -o json | jq -r '.items[0].metadata.name')
              velero restore create test-restore-$(date +%Y%m%d) \
                --from-backup $LATEST_BACKUP \
                --namespace-mappings hexabase-system:backup-test-$(date +%Y%m%d) \
                --wait
              
              # Validate restoration
              kubectl wait --for=condition=Ready pods \
                -n backup-test-$(date +%Y%m%d) \
                -l app=hexabase-api \
                --timeout=300s
              
              # Run data validation
              kubectl exec -n backup-test-$(date +%Y%m%d) \
                deployment/hexabase-api -- \
                hexabase-cli validate --mode recovery
              
              # Cleanup
              kubectl delete namespace backup-test-$(date +%Y%m%d)
              
              # Report result
              if [ $? -eq 0 ]; then
                echo "Recovery test passed"
                curl -X POST $SLACK_WEBHOOK -d '{"text":"✅ Weekly backup recovery test passed"}'
              else
                echo "Recovery test failed"
                curl -X POST $SLACK_WEBHOOK -d '{"text":"❌ Weekly backup recovery test FAILED"}'
              fi
          restartPolicy: OnFailure
```

### 2. Backup Integrity Verification

```bash
#!/bin/bash
# verify-backup-integrity.sh

# Verify PostgreSQL backup
aws s3 ls s3://hexabase-backups/postgres/ --recursive | tail -5 | while read line; do
  FILE=$(echo $line | awk '{print $4}')
  SIZE=$(echo $line | awk '{print $3}')
  
  # Download and verify
  aws s3 cp s3://hexabase-backups/postgres/$FILE /tmp/
  
  # Check file integrity
  if [[ $FILE == *.enc ]]; then
    openssl enc -aes-256-cbc -d -salt \
      -in /tmp/$(basename $FILE) \
      -out /tmp/$(basename $FILE .enc) \
      -pass file:/etc/backup/encryption.key
  fi
  
  # Verify PostgreSQL dump
  pg_restore --list /tmp/$(basename $FILE .enc) > /dev/null 2>&1
  if [ $? -eq 0 ]; then
    echo "✓ Backup $FILE is valid"
  else
    echo "✗ Backup $FILE is CORRUPT"
  fi
  
  rm -f /tmp/$(basename $FILE)*
done
```

## Best Practices

### 1. Backup Strategy

- **3-2-1 Rule**: 3 copies, 2 different media, 1 offsite
- **Regular Testing**: Weekly automated recovery tests
- **Encryption**: All backups encrypted at rest and in transit
- **Versioning**: Keep multiple versions with retention policy
- **Documentation**: Maintain detailed recovery procedures

### 2. Security

```yaml
# backup-security-policy.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: velero
  namespace: velero
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: velero-backup-role
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["persistentvolumes", "persistentvolumeclaims"]
  verbs: ["get", "list", "watch", "create", "update", "patch"]
---
# Encrypt backup bucket
aws s3api put-bucket-encryption \
  --bucket hexabase-backups \
  --server-side-encryption-configuration '{
    "Rules": [{
      "ApplyServerSideEncryptionByDefault": {
        "SSEAlgorithm": "aws:kms",
        "KMSMasterKeyID": "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
      }
    }]
  }'
```

### 3. Cost Optimization

```bash
# Set lifecycle policies for backup storage
aws s3api put-bucket-lifecycle-configuration \
  --bucket hexabase-backups \
  --lifecycle-configuration file://lifecycle.json

# lifecycle.json
{
  "Rules": [
    {
      "ID": "TransitionToGlacier",
      "Status": "Enabled",
      "Transitions": [
        {
          "Days": 30,
          "StorageClass": "GLACIER"
        },
        {
          "Days": 90,
          "StorageClass": "DEEP_ARCHIVE"
        }
      ]
    },
    {
      "ID": "DeleteOldBackups",
      "Status": "Enabled",
      "Expiration": {
        "Days": 365
      }
    }
  ]
}
```

## Troubleshooting

### Common Issues

**Velero backup stuck in "InProgress"**
```bash
# Check backup logs
velero backup logs <backup-name>

# Check node agent logs
kubectl logs -n velero -l name=node-agent

# Force completion
velero backup delete <backup-name> --confirm
```

**PostgreSQL WAL archiving failing**
```bash
# Check archive status
kubectl exec -n hexabase-system hexabase-db-1 -- \
  psql -U postgres -c "SELECT * FROM pg_stat_archiver;"

# Check S3 connectivity
kubectl exec -n hexabase-system hexabase-db-1 -- \
  aws s3 ls s3://hexabase-backups/postgres/
```

**Recovery validation failing**
```bash
# Check restored resources
kubectl get all -n <restored-namespace>

# Verify data integrity
kubectl exec -n <restored-namespace> deployment/hexabase-api -- \
  hexabase-cli validate --verbose
```

## Documentation

Maintain the following documentation:

1. **Recovery Runbook**: Step-by-step procedures
2. **Contact List**: Who to call during disasters
3. **Architecture Diagrams**: Current and recovery architectures
4. **Credential Inventory**: Where to find backup credentials
5. **Test Results**: History of recovery tests

## Resources

- [Velero Documentation](https://velero.io/docs/)
- [CloudNativePG Backup Guide](https://cloudnative-pg.io/docs/backup/)
- [Kubernetes Volume Snapshots](https://kubernetes.io/docs/concepts/storage/volume-snapshots/)
- [AWS S3 Glacier](https://aws.amazon.com/glacier/)
- [Disaster Recovery Planning](https://aws.amazon.com/disaster-recovery/)