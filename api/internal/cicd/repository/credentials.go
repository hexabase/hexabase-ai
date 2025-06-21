package repository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hexabase/hexabase-ai/api/internal/cicd/domain"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// KubernetesCredentialManager implements credential management using Kubernetes secrets
type KubernetesCredentialManager struct {
	kubeClient kubernetes.Interface
	namespace  string
}

// NewKubernetesCredentialManager creates a new Kubernetes credential manager
func NewKubernetesCredentialManager(kubeClient kubernetes.Interface, namespace string) domain.CredentialManager {
	return &KubernetesCredentialManager{
		kubeClient: kubeClient,
		namespace:  namespace,
	}
}

// StoreGitCredential stores Git credentials as a Kubernetes secret
func (m *KubernetesCredentialManager) StoreGitCredential(workspaceID string, cred *domain.GitCredential) (*domain.CredentialInfo, error) {
	credentialID := uuid.New().String()
	secretName := m.formatSecretName(workspaceID, "git", credentialID)
	
	var secret *corev1.Secret
	
	switch cred.Type {
	case "ssh-key":
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: m.namespace,
				Labels: map[string]string{
					"app.kubernetes.io/managed-by": "hexabase-ai",
					"hexabase.ai/workspace-id":     workspaceID,
					"hexabase.ai/credential-type":  "git",
					"hexabase.ai/credential-id":    credentialID,
				},
				Annotations: map[string]string{
					"hexabase.ai/created-at": time.Now().Format(time.RFC3339),
				},
			},
			Type: corev1.SecretTypeSSHAuth,
			Data: map[string][]byte{
				corev1.SSHAuthPrivateKey: []byte(cred.SSHKey),
			},
		}
		
	case "token":
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: m.namespace,
				Labels: map[string]string{
					"app.kubernetes.io/managed-by": "hexabase-ai",
					"hexabase.ai/workspace-id":     workspaceID,
					"hexabase.ai/credential-type":  "git",
					"hexabase.ai/credential-id":    credentialID,
				},
				Annotations: map[string]string{
					"hexabase.ai/created-at": time.Now().Format(time.RFC3339),
				},
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{
				"username": []byte(cred.Username),
				"password": []byte(cred.Token),
			},
		}
		
	default:
		return nil, fmt.Errorf("unsupported git credential type: %s", cred.Type)
	}
	
	ctx := context.Background()
	
	// Create the secret
	_, err := m.kubeClient.CoreV1().Secrets(m.namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to store git credential: %w", err)
	}
	
	return &domain.CredentialInfo{
		ID:          credentialID,
		WorkspaceID: workspaceID,
		Name:        secretName,
		Type:        "git",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// GetGitCredential retrieves Git credentials from Kubernetes secret
func (m *KubernetesCredentialManager) GetGitCredential(workspaceID, credentialID string) (*domain.GitCredential, error) {
	secretName := m.formatSecretName(workspaceID, "git", credentialID)
	
	ctx := context.Background()
	secret, err := m.kubeClient.CoreV1().Secrets(m.namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get git credential: %w", err)
	}
	
	cred := &domain.GitCredential{}
	
	if secret.Type == corev1.SecretTypeSSHAuth {
		cred.Type = "ssh-key"
		cred.SSHKey = string(secret.Data[corev1.SSHAuthPrivateKey])
	} else if secret.Type == corev1.SecretTypeOpaque {
		cred.Type = "token"
		cred.Username = string(secret.Data["username"])
		cred.Token = string(secret.Data["password"])
	}
	
	return cred, nil
}

// StoreRegistryCredential stores container registry credentials
func (m *KubernetesCredentialManager) StoreRegistryCredential(workspaceID string, cred *domain.RegistryCredential) (*domain.CredentialInfo, error) {
	credentialID := uuid.New().String()
	secretName := m.formatSecretName(workspaceID, "registry", credentialID)
	
	// Create Docker config JSON
	dockerConfig := map[string]map[string]map[string]string{
		"auths": {
			cred.Registry: {
				"username": cred.Username,
				"password": cred.Password,
				"email":    cred.Email,
				"auth":     base64.StdEncoding.EncodeToString([]byte(cred.Username + ":" + cred.Password)),
			},
		},
	}
	
	dockerConfigJSON, err := json.Marshal(dockerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal docker config: %w", err)
	}
	
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: m.namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "hexabase-ai",
				"hexabase.ai/workspace-id":     workspaceID,
				"hexabase.ai/credential-type":  "registry",
				"hexabase.ai/credential-id":    credentialID,
			},
			Annotations: map[string]string{
				"hexabase.ai/created-at": time.Now().Format(time.RFC3339),
				"hexabase.ai/registry":   cred.Registry,
			},
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: dockerConfigJSON,
		},
	}
	
	ctx := context.Background()
	
	// Create the secret
	_, err = m.kubeClient.CoreV1().Secrets(m.namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to store registry credential: %w", err)
	}
	
	return &domain.CredentialInfo{
		ID:          credentialID,
		WorkspaceID: workspaceID,
		Name:        secretName,
		Type:        "registry",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// GetRegistryCredential retrieves registry credentials from Kubernetes secret
func (m *KubernetesCredentialManager) GetRegistryCredential(workspaceID, credentialID string) (*domain.RegistryCredential, error) {
	secretName := m.formatSecretName(workspaceID, "registry", credentialID)
	
	ctx := context.Background()
	secret, err := m.kubeClient.CoreV1().Secrets(m.namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get registry credential: %w", err)
	}
	
	if secret.Type != corev1.SecretTypeDockerConfigJson {
		return nil, fmt.Errorf("invalid secret type for registry credential")
	}
	
	dockerConfigJSON := secret.Data[corev1.DockerConfigJsonKey]
	var dockerConfig map[string]map[string]map[string]string
	
	err = json.Unmarshal(dockerConfigJSON, &dockerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal docker config: %w", err)
	}
	
	// Find the first registry in the config
	for registry, auth := range dockerConfig["auths"] {
		return &domain.RegistryCredential{
			Registry: registry,
			Username: auth["username"],
			Password: auth["password"],
			Email:    auth["email"],
		}, nil
	}
	
	return nil, fmt.Errorf("no registry found in credential")
}

// ListCredentials lists available credentials for a workspace
func (m *KubernetesCredentialManager) ListCredentials(workspaceID string) ([]*domain.CredentialInfo, error) {
	ctx := context.Background()
	
	// List secrets with workspace label
	secrets, err := m.kubeClient.CoreV1().Secrets(m.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("hexabase.ai/workspace-id=%s", workspaceID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}
	
	credentials := []*domain.CredentialInfo{}
	for _, secret := range secrets.Items {
		credType := secret.Labels["hexabase.ai/credential-type"]
		credID := secret.Labels["hexabase.ai/credential-id"]
		
		if credType != "" && credID != "" {
			createdAt, _ := time.Parse(time.RFC3339, secret.Annotations["hexabase.ai/created-at"])
			
			credentials = append(credentials, &domain.CredentialInfo{
				ID:          credID,
				WorkspaceID: workspaceID,
				Name:        secret.Name,
				Type:        credType,
				CreatedAt:   createdAt,
				UpdatedAt:   secret.CreationTimestamp.Time,
			})
		}
	}
	
	return credentials, nil
}

// DeleteCredential deletes a stored credential
func (m *KubernetesCredentialManager) DeleteCredential(workspaceID, credentialID string) error {
	ctx := context.Background()
	
	// Try both git and registry secret patterns
	patterns := []string{
		m.formatSecretName(workspaceID, "git", credentialID),
		m.formatSecretName(workspaceID, "registry", credentialID),
	}
	
	var deleted bool
	for _, secretName := range patterns {
		err := m.kubeClient.CoreV1().Secrets(m.namespace).Delete(ctx, secretName, metav1.DeleteOptions{})
		if err == nil {
			deleted = true
			break
		}
	}
	
	if !deleted {
		return fmt.Errorf("credential not found: %s", credentialID)
	}
	
	return nil
}

// CreateKubernetesSecret creates a generic Kubernetes secret
func (m *KubernetesCredentialManager) CreateKubernetesSecret(workspaceID, secretName string, data map[string][]byte) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: m.namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "hexabase-ai",
				"hexabase.ai/workspace-id":     workspaceID,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}
	
	ctx := context.Background()
	_, err := m.kubeClient.CoreV1().Secrets(m.namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}
	
	return nil
}

// GetKubernetesSecret retrieves a generic Kubernetes secret
func (m *KubernetesCredentialManager) GetKubernetesSecret(workspaceID, secretName string) (map[string][]byte, error) {
	ctx := context.Background()
	secret, err := m.kubeClient.CoreV1().Secrets(m.namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}
	
	// Verify workspace ownership
	if secret.Labels["hexabase.ai/workspace-id"] != workspaceID {
		return nil, fmt.Errorf("access denied: secret does not belong to workspace")
	}
	
	return secret.Data, nil
}

// DeleteKubernetesSecret deletes a generic Kubernetes secret
func (m *KubernetesCredentialManager) DeleteKubernetesSecret(workspaceID, secretName string) error {
	ctx := context.Background()
	
	// First verify the secret belongs to the workspace
	secret, err := m.kubeClient.CoreV1().Secrets(m.namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}
	
	if secret.Labels["hexabase.ai/workspace-id"] != workspaceID {
		return fmt.Errorf("access denied: secret does not belong to workspace")
	}
	
	err = m.kubeClient.CoreV1().Secrets(m.namespace).Delete(ctx, secretName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	
	return nil
}

// formatSecretName formats a secret name with workspace and type prefix
func (m *KubernetesCredentialManager) formatSecretName(workspaceID, credType, credentialID string) string {
	return fmt.Sprintf("cicd-%s-%s-%s", workspaceID, credType, credentialID)
}