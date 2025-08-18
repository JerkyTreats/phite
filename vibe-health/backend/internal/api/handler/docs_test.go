package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSwaggerUIHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	w := httptest.NewRecorder()

	SwaggerUIHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "swagger-ui")
	assert.Contains(t, w.Body.String(), "/api/docs/openapi.json")
}

func TestReDocHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/redoc", nil)
	w := httptest.NewRecorder()

	ReDocHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "redoc")
	assert.Contains(t, w.Body.String(), "/api/docs/openapi.json")
}

func TestOpenAPIJSONHandler(t *testing.T) {
	// Create a temporary JSON file for testing
	tempDir := t.TempDir()
	docsDir := filepath.Join(tempDir, "docs", "api")
	err := os.MkdirAll(docsDir, 0755)
	require.NoError(t, err)

	testSpec := map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":   "Test API",
			"version": "1.0.0",
		},
	}
	specData, err := json.Marshal(testSpec)
	require.NoError(t, err)

	jsonFile := filepath.Join(docsDir, "openapi.json")
	err = os.WriteFile(jsonFile, specData, 0644)
	require.NoError(t, err)

	// Change to temp directory for the test
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/docs/openapi.json", nil)
	w := httptest.NewRecorder()

	OpenAPIJSONHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "3.0.3", response["openapi"])
}

func TestOpenAPIJSONHandler_FileNotFound(t *testing.T) {
	// Change to a temp directory where no docs exist
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/docs/openapi.json", nil)
	w := httptest.NewRecorder()

	OpenAPIJSONHandler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOpenAPIYAMLHandler(t *testing.T) {
	// Create a temporary YAML file for testing
	tempDir := t.TempDir()
	docsDir := filepath.Join(tempDir, "docs", "api")
	err := os.MkdirAll(docsDir, 0755)
	require.NoError(t, err)

	testYAML := `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0`

	yamlFile := filepath.Join(docsDir, "openapi.yaml")
	err = os.WriteFile(yamlFile, []byte(testYAML), 0644)
	require.NoError(t, err)

	// Change to temp directory for the test
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/docs/openapi.yaml", nil)
	w := httptest.NewRecorder()

	OpenAPIYAMLHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/x-yaml", w.Header().Get("Content-Type"))
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Body.String(), "openapi: 3.0.3")
}