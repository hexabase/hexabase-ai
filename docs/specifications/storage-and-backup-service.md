# HKS Storage and Backup Service

## 1. Overview

This document specifies the design for the HKS Storage and Backup Service. This feature allows users on dedicated plans to provision persistent block storage for their workspaces (vClusters) and configure automated backup jobs. The service integrates with the underlying Proxmox virtualization platform for storage management and leverages AIOps for proactive monitoring and user assistance.

## 2. Target Audience and Prerequisites

- **Primary User**: Workspace Admin
- **Prerequisites**: The workspace must be subscribed to a "Dedicated Plan" that grants access to dedicated underlying hardware resources.

## 3. Core Components

### 3.1 Storage Volume

A `Storage Volume` is a block storage device (e.g., ZFS dataset, LVM-Thin volume) provisioned from the Proxmox host and attached to a specific workspace. From within the Kubernetes vCluster, this volume is represented as a dedicated `StorageClass` and `PersistentVolume` (PV), making it consumable by user applications via standard `PersistentVolumeClaim` (PVC) mechanisms.

### 3.2 Backup Object

A `Backup` is a custom HKS resource that defines a scheduled backup operation. It can be created under the "Application" entity and encapsulates:

- A schedule (in cron format).
- The source data to be backed up (e.g., a specific PVC).
- The destination path within a provisioned `Storage Volume`. It can create the folder if it does not exist.
- The backup logic, executed as a containerized function.

## 4. Architecture and Flows

### 4.1. Storage Provisioning Flow

1.  A Workspace Admin requests a new `Storage Volume` through the HKS UI, specifying the required size (e.g., 500GB).
2.  The HKS Control Plane receives the request and verifies that the workspace is on a "Dedicated Plan".
3.  The Control Plane communicates with the Proxmox API to provision the requested storage on the host node.
4.  Upon successful creation in Proxmox, the HKS Control Plane creates a corresponding `StorageClass` and `PersistentVolume` (PV) within the target vCluster.
5.  The workspace's billing information is updated to reflect the new storage cost.
6.  The `Storage Volume` is now available for use within the workspace.

### 4.2. Quota Management

- The total provisioned `Storage Volume` size acts as a hard limit for the workspace.
- To manage storage distribution _within_ the workspace, the Workspace Admin is responsible for creating standard Kubernetes `ResourceQuota` objects in each namespace.
- These quotas can limit the total size of `PersistentVolumeClaims` that can be created in a namespace, ensuring fair use among different projects or applications.

### 4.3. Backup Creation and Execution Flow

1.  A user creates a `Backup` object via the HKS UI under an Application. They define the schedule, source, destination, and the function for the backup logic.
2.  The HKS Control Plane translates this `Backup` object into a Kubernetes `CronJob`.
3.  The `CronJob` template will define a pod that:
    - Mounts the source PVC and the destination backup `Storage Volume`.
    - Runs a container executing the specified backup `function` (e.g., a script that runs `tar`, `rsync`, or a database dump tool).
4.  At the scheduled time, Kubernetes runs the job. If the backup fails due to exceeding the quota, this event is captured.

## 5. AIOps Integration and Monitoring

The AIOps system plays a crucial role in ensuring the reliability of the backup service.

- **Monitoring**: A specialized `StorageWorkerAgent` in the AIOps system continuously monitors PVC usage metrics from Prometheus.
- **Proactive Alerting**: When the total usage on a backup `Storage Volume` exceeds a predefined threshold (e.g., 85%), the `StorageWorkerAgent` flags this condition.
- **User Interaction**:
  1.  The `StorageWorkerAgent` passes the alert to the `OrchestrationAgent`.
  2.  The `OrchestrationAgent` instructs the `UserChatAgent` to notify the Workspace Admin.
  3.  The `UserChatAgent` sends a message: "Your backup storage volume 'backup-vol-1' has reached 87% capacity. Future backups may fail. Would you like to expand the volume?"
  4.  If the admin agrees, the agent can initiate the storage expansion workflow by calling the appropriate HKS API.
- **Failure Notification**: If a backup job fails due to exceeding the quota, the AIOps system will also notify the user of the specific failure and suggest remediation.

## 6. API and UI Considerations

- **HKS API**: New endpoints will be required to manage the lifecycle of `Storage Volume` and `Backup` objects.
- **UI**: The HKS Dashboard will be updated with:
  - A new "Storage" section for Workspace Admins to manage volumes.
  - A "Backup" tab within the "Application" view to configure backup jobs.
