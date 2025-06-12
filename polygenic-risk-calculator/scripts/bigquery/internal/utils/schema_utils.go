package utils

import (
	"fmt"

	"cloud.google.com/go/bigquery"
)

// PrintField prints a BigQuery field schema with proper indentation
func PrintField(field *bigquery.FieldSchema, indent int) {
	indentStr := ""
	for i := 0; i < indent; i++ {
		indentStr += "  "
	}

	fmt.Printf("%s%s (%s)", indentStr, field.Name, field.Type)
	if field.Repeated {
		fmt.Print(" [REPEATED]")
	}
	if field.Required {
		fmt.Print(" [REQUIRED]")
	}
	fmt.Println()

	for _, subField := range field.Schema {
		PrintField(subField, indent+1)
	}
}

// CheckRequiredField checks if a field exists in the required fields map
func CheckRequiredField(field *bigquery.FieldSchema, requiredFields map[string]bool) {
	if _, exists := requiredFields[field.Name]; exists {
		requiredFields[field.Name] = true
	}

	for _, subField := range field.Schema {
		CheckRequiredField(subField, requiredFields)
	}
}
