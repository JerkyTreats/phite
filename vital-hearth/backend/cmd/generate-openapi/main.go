package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/JerkyTreats/PHITE/vital-hearth/backend/cmd/generate-openapi/analyzer"
	
	// Import packages to trigger init() functions that register routes
	_ "github.com/JerkyTreats/PHITE/vital-hearth/backend/internal/api/handler"
)

func main() {
	var (
		outputFile = flag.String("output", "docs/api/openapi.yaml", "Output file for OpenAPI specification")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	log.Printf("Starting OpenAPI specification generation...")
	log.Printf("Output file: %s", *outputFile)

	// Create analyzer
	gen := analyzer.NewGenerator()

	// Generate the OpenAPI specification
	spec, err := gen.GenerateSpec()
	if err != nil {
		log.Fatalf("Failed to generate OpenAPI spec: %v", err)
	}

	// Write YAML output file
	if err := os.WriteFile(*outputFile, []byte(spec), 0644); err != nil {
		log.Fatalf("Failed to write spec to file: %v", err)
	}

	// Also generate JSON version for broader tool compatibility
	jsonSpec, err := gen.GenerateJSONSpec()
	if err != nil {
		log.Printf("Warning: Failed to generate JSON spec: %v", err)
	} else {
		jsonOutputFile := strings.TrimSuffix(*outputFile, ".yaml") + ".json"
		if strings.HasSuffix(*outputFile, ".yml") {
			jsonOutputFile = strings.TrimSuffix(*outputFile, ".yml") + ".json"
		}
		if err := os.WriteFile(jsonOutputFile, []byte(jsonSpec), 0644); err != nil {
			log.Printf("Warning: Failed to write JSON spec: %v", err)
		} else {
			log.Printf("JSON specification generated at %s", jsonOutputFile)
		}
	}

	log.Printf("OpenAPI specification generated successfully at %s", *outputFile)
	fmt.Printf("Generated OpenAPI spec with %d routes\n", len(gen.GetDiscoveredRoutes()))
}