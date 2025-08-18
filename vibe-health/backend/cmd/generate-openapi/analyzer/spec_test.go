package analyzer

import (
	"reflect"
	"testing"

	"github.com/JerkyTreats/PHITE/vibe-health/backend/internal/api/types"
	"github.com/stretchr/testify/assert"
)

func TestGenerateOperationID(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name     string
		route    types.RouteInfo
		expected string
	}{
		{
			name: "GET health endpoint",
			route: types.RouteInfo{
				Method: "GET",
				Path:   "/health",
			},
			expected: "gethealth",
		},
		{
			name: "POST create endpoint",
			route: types.RouteInfo{
				Method: "POST",
				Path:   "/create-user",
			},
			expected: "createcreateUser",
		},
		{
			name: "DELETE endpoint",
			route: types.RouteInfo{
				Method: "DELETE",
				Path:   "/delete-user",
			},
			expected: "deletedeleteUser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.generateOperationID(tt.route)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildResponses(t *testing.T) {
	gen := NewGenerator()

	// Test route without response type
	route := types.RouteInfo{
		Method: "GET",
		Path:   "/health",
	}

	responses := gen.buildResponses(route)

	// Should have standard responses
	assert.Contains(t, responses, "200")
	assert.Contains(t, responses, "400")
	assert.Contains(t, responses, "500")

	// GET should not have 422
	assert.NotContains(t, responses, "422")

	// Test POST route
	route.Method = "POST"
	responses = gen.buildResponses(route)

	// POST should have 422
	assert.Contains(t, responses, "422")
}

func TestBuildRequestBody(t *testing.T) {
	gen := NewGenerator()

	route := types.RouteInfo{
		Method:      "POST",
		Path:        "/create-user",
		Summary:     "Create a new user",
		RequestType: reflect.TypeOf(""),
	}

	requestBody := gen.buildRequestBody(route)

	assert.NotNil(t, requestBody)
	assert.True(t, requestBody.Required)
	assert.Contains(t, requestBody.Content, "application/json")
	assert.Contains(t, requestBody.Description, "Create a new user")
}
