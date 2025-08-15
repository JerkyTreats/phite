package analyzer

import (
	"reflect"
	"testing"

	"github.com/JerkyTreats/PHITE/vital-hearth/backend/internal/api/types"
	"github.com/stretchr/testify/assert"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator()
	assert.NotNil(t, gen)
	assert.NotNil(t, gen.fileSet)
	assert.NotNil(t, gen.typeSchemas)
	assert.Equal(t, 0, len(gen.routes))
}

func TestGetTypeName(t *testing.T) {
	gen := NewGenerator()
	
	tests := []struct {
		name     string
		input    reflect.Type
		expected string
	}{
		{
			name:     "simple string type",
			input:    reflect.TypeOf(""),
			expected: "string",
		},
		{
			name:     "simple int type",
			input:    reflect.TypeOf(0),
			expected: "int",
		},
		{
			name:     "struct type",
			input:    reflect.TypeOf(types.RouteInfo{}),
			expected: "RouteInfo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.getTypeName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAddStandardSchemas(t *testing.T) {
	gen := NewGenerator()
	gen.addStandardSchemas()
	
	assert.Contains(t, gen.typeSchemas, "ErrorResponse")
	
	errorSchema := gen.typeSchemas["ErrorResponse"]
	schemaMap, ok := errorSchema.(map[string]interface{})
	assert.True(t, ok)
	
	assert.Equal(t, "object", schemaMap["type"])
	assert.Contains(t, schemaMap, "properties")
	assert.Contains(t, schemaMap, "required")
}