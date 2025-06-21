package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActivity_SetDetails(t *testing.T) {
	t.Run("sets details from struct", func(t *testing.T) {
		activity := &Activity{}
		data := struct {
			Role   string `json:"role"`
			UserID string `json:"user_id"`
		}{
			Role:   "admin",
			UserID: "user-123",
		}

		err := activity.SetDetails(data)
		require.NoError(t, err)
		assert.JSONEq(t, `{"role":"admin","user_id":"user-123"}`, activity.Details)
	})

	t.Run("sets details from map", func(t *testing.T) {
		activity := &Activity{}
		data := map[string]interface{}{
			"old_role": "member",
			"new_role": "admin",
		}

		err := activity.SetDetails(data)
		require.NoError(t, err)
		assert.JSONEq(t, `{"old_role":"member","new_role":"admin"}`, activity.Details)
	})

	t.Run("handles nil data", func(t *testing.T) {
		activity := &Activity{}
		err := activity.SetDetails(nil)
		require.NoError(t, err)
		assert.Equal(t, "", activity.Details)
	})

	t.Run("returns error for unmarshalable data", func(t *testing.T) {
		activity := &Activity{}
		// channels cannot be marshaled to JSON
		data := map[string]interface{}{
			"invalid": make(chan int),
		}

		err := activity.SetDetails(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal activity details")
	})
}

func TestActivity_GetDetails(t *testing.T) {
	t.Run("gets details into struct", func(t *testing.T) {
		activity := &Activity{
			Details: `{"role":"admin","user_id":"user-123"}`,
		}

		var result struct {
			Role   string `json:"role"`
			UserID string `json:"user_id"`
		}

		err := activity.GetDetails(&result)
		require.NoError(t, err)
		assert.Equal(t, "admin", result.Role)
		assert.Equal(t, "user-123", result.UserID)
	})

	t.Run("gets details into map", func(t *testing.T) {
		activity := &Activity{
			Details: `{"old_role":"member","new_role":"admin"}`,
		}

		var result map[string]interface{}
		err := activity.GetDetails(&result)
		require.NoError(t, err)
		assert.Equal(t, "member", result["old_role"])
		assert.Equal(t, "admin", result["new_role"])
	})

	t.Run("handles empty details", func(t *testing.T) {
		activity := &Activity{Details: ""}

		var result map[string]interface{}
		err := activity.GetDetails(&result)
		require.NoError(t, err)
		// result should remain nil/unchanged
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		activity := &Activity{
			Details: `{"invalid": json}`,
		}

		var result map[string]interface{}
		err := activity.GetDetails(&result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal activity details")
	})
}

func TestActivity_GetDetailsAsMap(t *testing.T) {
	t.Run("returns details as map", func(t *testing.T) {
		activity := &Activity{
			Details: `{"role":"admin","count":5,"active":true}`,
		}

		result, err := activity.GetDetailsAsMap()
		require.NoError(t, err)
		assert.Equal(t, "admin", result["role"])
		assert.Equal(t, float64(5), result["count"]) // JSON numbers become float64
		assert.Equal(t, true, result["active"])
	})

	t.Run("returns empty map for empty details", func(t *testing.T) {
		activity := &Activity{Details: ""}

		result, err := activity.GetDetailsAsMap()
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		activity := &Activity{
			Details: `{"invalid": json}`,
		}

		result, err := activity.GetDetailsAsMap()
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to unmarshal activity details as map")
	})
}

func TestActivity_SetDetailsFromMap(t *testing.T) {
	t.Run("sets details from map", func(t *testing.T) {
		activity := &Activity{}
		data := map[string]interface{}{
			"action":    "role_change",
			"old_role":  "member",
			"new_role":  "admin",
			"timestamp": time.Now().Unix(),
		}

		err := activity.SetDetailsFromMap(data)
		require.NoError(t, err)
		assert.NotEmpty(t, activity.Details)

		// Verify we can get it back
		result, err := activity.GetDetailsAsMap()
		require.NoError(t, err)
		assert.Equal(t, "role_change", result["action"])
		assert.Equal(t, "member", result["old_role"])
		assert.Equal(t, "admin", result["new_role"])
	})
}

func TestActivity_HelperMethods_Integration(t *testing.T) {
	t.Run("round trip conversion works correctly", func(t *testing.T) {
		activity := &Activity{
			ID:             "activity-123",
			OrganizationID: "org-456",
			UserID:         "user-789",
			Type:           "member",
			Action:         "role_updated",
			ResourceType:   "organization_user",
			ResourceID:     "user-789",
			Timestamp:      time.Now(),
		}

		// Set details using helper
		originalData := map[string]interface{}{
			"old_role":    "member",
			"new_role":    "admin",
			"changed_by":  "admin-user",
			"reason":      "promotion",
			"permissions": []string{"read", "write", "admin"},
		}

		err := activity.SetDetailsFromMap(originalData)
		require.NoError(t, err)

		// Get details back using helper
		retrievedData, err := activity.GetDetailsAsMap()
		require.NoError(t, err)

		// Verify all data is preserved
		assert.Equal(t, "member", retrievedData["old_role"])
		assert.Equal(t, "admin", retrievedData["new_role"])
		assert.Equal(t, "admin-user", retrievedData["changed_by"])
		assert.Equal(t, "promotion", retrievedData["reason"])
		
		// Arrays become []interface{} in JSON
		permissions := retrievedData["permissions"].([]interface{})
		assert.Len(t, permissions, 3)
		assert.Equal(t, "read", permissions[0])
		assert.Equal(t, "write", permissions[1])
		assert.Equal(t, "admin", permissions[2])
	})
} 