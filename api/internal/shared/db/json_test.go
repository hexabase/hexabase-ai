package db

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSON_ScanAndValue(t *testing.T) {
	t.Run("scan from bytes", func(t *testing.T) {
		var j JSON
		jsonData := []byte(`{"key":"value"}`)
		
		err := j.Scan(jsonData)
		assert.NoError(t, err)
		
		// Verify we can unmarshal it
		var result map[string]string
		err = json.Unmarshal(j, &result)
		assert.NoError(t, err)
		assert.Equal(t, "value", result["key"])
	})
	
	t.Run("scan from string", func(t *testing.T) {
		var j JSON
		jsonStr := `{"key":"value"}`
		
		err := j.Scan(jsonStr)
		assert.NoError(t, err)
		
		// Verify we can unmarshal it
		var result map[string]string
		err = json.Unmarshal(j, &result)
		assert.NoError(t, err)
		assert.Equal(t, "value", result["key"])
	})
	
	t.Run("value returns proper json", func(t *testing.T) {
		j := JSON(`{"key":"value"}`)
		
		val, err := j.Value()
		assert.NoError(t, err)
		
		// The value should be the JSON string
		jsonBytes, ok := val.([]byte)
		assert.True(t, ok)
		assert.Equal(t, `{"key":"value"}`, string(jsonBytes))
	})
}