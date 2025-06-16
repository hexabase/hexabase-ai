package httpauth

import "testing"

func TestHasBearerPrefix(t *testing.T) {
	cases := []struct {
		name     string
		header   string
		expected bool
	}{
		{
			name:     "valid bearer token",
			header:   "Bearer token",
			expected: true,
		},
		{
			name:     "valid bearer token with lowercase",
			header:   "bearer token",
			expected: true,
		},
		{
			name:     "valid bearer token with uppercase",
			header:   "BEARER token",
			expected: true,
		},
		{
			name:     "valid bearer token with mixed case",
			header:   "BeArEr token",
			expected: true,
		},
		{
			name:     "multiple spaces after bearer",
			header:   "Bearer  token",
			expected: true,
		},
		{
			name:     "valid bearer token without token",
			header:   "Bearer ",
			expected: true,
		},
		{
			name:     "invalid with other scheme",
			header:   "Basic token",
			expected: false,
		},
		{
			name:     "invalid without space",
			header:   "Bearer",
			expected: false,
		},
		{
			name:     "no space between bearer and token",
			header:   "Bearertoken",
			expected: false,
		},
		{
			name:     "invalid with leading space",
			header:   " Bearer token",
			expected: false,
		},
		{
			name:     "empty header",
			header:   "",
			expected: false,
		},
		{
			name:     "invalid with no scheme",
			header:   "token",
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := HasBearerPrefix(tc.header)
			if actual != tc.expected {
				t.Errorf("expected %v, but got %v", tc.expected, actual)
			}
		})
	}
}

func TestTrimBearerPrefix(t *testing.T) {
	cases := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "valid bearer token",
			header:   "Bearer token",
			expected: "token",
		},
		{
			name:     "valid bearer token with lowercase",
			header:   "bearer token",
			expected: "token",
		},
		{
			name:     "valid bearer token with uppercase",
			header:   "BEARER token",
			expected: "token",
		},
		{
			name:     "valid bearer token with mixed case",
			header:   "BeArEr token",
			expected: "token",
		},
		{
			name:     "no bearer prefix",
			header:   "token",
			expected: "token",
		},
		{
			name:     "empty header",
			header:   "",
			expected: "",
		},
		{
			name:     "bearer only",
			header:   "Bearer ",
			expected: "",
		},
		{
			name:     "bearer without space",
			header:   "BearerToken",
			expected: "BearerToken",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := TrimBearerPrefix(tc.header)
			if actual != tc.expected {
				t.Errorf("expected %q, but got %q", tc.expected, actual)
			}
		})
	}
} 