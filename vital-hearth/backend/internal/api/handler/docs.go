package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/JerkyTreats/PHITE/vital-hearth/backend/internal/logging"
)

// SwaggerUIHandler serves the Swagger UI interface
func SwaggerUIHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/api/docs/openapi.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// OpenAPIJSONHandler serves the OpenAPI specification in JSON format
func OpenAPIJSONHandler(w http.ResponseWriter, r *http.Request) {
	// Try to read the generated JSON file
	jsonPath := filepath.Join("docs", "api", "openapi.json")
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		// Fallback to swagger.json if openapi.json doesn't exist
		jsonPath = filepath.Join("docs", "api", "swagger.json")
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		logging.Error("Failed to read OpenAPI JSON file: %v", err)
		http.Error(w, "OpenAPI specification not found", http.StatusNotFound)
		return
	}

	// Validate it's valid JSON
	var spec map[string]interface{}
	if err := json.Unmarshal(data, &spec); err != nil {
		logging.Error("Invalid JSON in OpenAPI file: %v", err)
		http.Error(w, "Invalid OpenAPI specification", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// OpenAPIYAMLHandler serves the OpenAPI specification in YAML format
func OpenAPIYAMLHandler(w http.ResponseWriter, r *http.Request) {
	yamlPath := filepath.Join("docs", "api", "openapi.yaml")
	
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		logging.Error("Failed to read OpenAPI YAML file: %v", err)
		http.Error(w, "OpenAPI specification not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// ReDocHandler serves the ReDoc interface as an alternative to Swagger UI
func ReDocHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>API Documentation - ReDoc</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
    <style>
        body {
            margin: 0;
            padding: 0;
        }
    </style>
</head>
<body>
    <redoc spec-url='/api/docs/openapi.json'></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}