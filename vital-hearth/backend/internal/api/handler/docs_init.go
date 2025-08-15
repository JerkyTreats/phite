package handler

import (
	"github.com/JerkyTreats/PHITE/vital-hearth/backend/internal/api/types"
)

func init() {
	// Register Swagger UI endpoint
	types.RegisterRoute(types.RouteInfo{
		Method:  "GET",
		Path:    "/docs",
		Handler: SwaggerUIHandler,
		Module:  "docs",
		Summary: "Swagger UI documentation interface",
	})

	// Register alternative ReDoc interface
	types.RegisterRoute(types.RouteInfo{
		Method:  "GET",
		Path:    "/redoc",
		Handler: ReDocHandler,
		Module:  "docs",
		Summary: "ReDoc documentation interface",
	})

	// Register OpenAPI JSON endpoint
	types.RegisterRoute(types.RouteInfo{
		Method:  "GET",
		Path:    "/api/docs/openapi.json",
		Handler: OpenAPIJSONHandler,
		Module:  "docs",
		Summary: "OpenAPI specification in JSON format",
	})

	// Register OpenAPI YAML endpoint
	types.RegisterRoute(types.RouteInfo{
		Method:  "GET",
		Path:    "/api/docs/openapi.yaml",
		Handler: OpenAPIYAMLHandler,
		Module:  "docs",
		Summary: "OpenAPI specification in YAML format",
	})

	// Convenience redirect from root docs path
	types.RegisterRoute(types.RouteInfo{
		Method:  "GET",
		Path:    "/api/docs",
		Handler: SwaggerUIHandler,
		Module:  "docs",
		Summary: "API documentation (redirects to Swagger UI)",
	})
}