package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/YHVCorp/gitdocai-cli/schemas"
	"encoding/json"
	"strings"
)

// CreateFileWithApiReference creates an MDX file (and its directories) at the given path
// and writes the <ApiReference> component inside, using the provided complete OpenapiData.
func CreateFileWithApiReference(apiRef schemas.OpenapiData, path string) error {
	// Ensure directories exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Helper function to safely get string pointer value
	getStringValue := func(ptr *string) string {
		if ptr == nil {
			return ""
		}
		return *ptr
	}

	// Helper function to safely get bool pointer value
	getBoolValue := func(ptr *bool) bool {
		if ptr == nil {
			return false
		}
		return *ptr
	}

	// Helper function to convert slice to JSON string or return empty array
	sliceToJSON := func(slice any) string {
		if slice == nil {
			return "[]"
		}
		jsonBytes, err := json.Marshal(slice)
		if err != nil {
			return "[]"
		}
		return string(jsonBytes)
	}

	// Helper function to convert map/struct to JSON string or return empty object
	objectToJSON := func(obj any) string {
		if obj == nil {
			return "{}"
		}
		jsonBytes, err := json.Marshal(obj)
		if err != nil {
			return "{}"
		}
		return string(jsonBytes)
	}

	// Helper function to escape quotes in strings for JSX
	escapeString := func(s string) string {
		s = strings.ReplaceAll(s, "\"", "\\\"")
		s = strings.ReplaceAll(s, "\n", "\\n")
		s = strings.ReplaceAll(s, "\r", "\\r")
		s = strings.ReplaceAll(s, "\t", "\\t")
		return s
	}

	// Convert all data to JSON format
	parametersJSON := objectToJSON(apiRef.Parameters)
	responsesJSON := objectToJSON(apiRef.Responses)
	securitySchemasJson := objectToJSON(apiRef.SecuritySchemas)
	requestBodyJSON := objectToJSON(apiRef.RequestBody)
	securityJSON := sliceToJSON(apiRef.Security)
	serversJSON := sliceToJSON(apiRef.Servers)
	tagsJSON := sliceToJSON(apiRef.Tags)
	externalDocsJSON := objectToJSON(apiRef.ExternalDocs)

	// Build the component with all available data
	var componentBuilder strings.Builder
	componentBuilder.WriteString("<ApiReference\n")

	// Basic required fields
	componentBuilder.WriteString(fmt.Sprintf("  title=\"%s\"\n", escapeString(apiRef.Title)))
	componentBuilder.WriteString(fmt.Sprintf("  method=\"%s\"\n", apiRef.Method))
	componentBuilder.WriteString(fmt.Sprintf("  path=\"%s\"\n", escapeString(apiRef.Path)))

	// Optional string fields
	if description := getStringValue(apiRef.Description); description != "" {
		componentBuilder.WriteString(fmt.Sprintf("  description=\"%s\"\n", escapeString(description)))
	}

	if operationId := getStringValue(apiRef.OperationId); operationId != "" {
		componentBuilder.WriteString(fmt.Sprintf("  operationId=\"%s\"\n", escapeString(operationId)))
	}

	if summary := getStringValue(apiRef.Summary); summary != "" {
		componentBuilder.WriteString(fmt.Sprintf("  summary=\"%s\"\n", escapeString(summary)))
	}

	if tryItBaseUrl := getStringValue(apiRef.TryItBaseUrl); tryItBaseUrl != "" {
		componentBuilder.WriteString(fmt.Sprintf("  tryItBaseUrl=\"%s\"\n", escapeString(tryItBaseUrl)))
	}

	// Boolean fields
	if getBoolValue(apiRef.Deprecated) {
		componentBuilder.WriteString("  deprecated={true}\n")
	}

	// Array and object fields (as JSON)
	if len(apiRef.Tags) > 0 {
		componentBuilder.WriteString(fmt.Sprintf("  tags={%s}\n", tagsJSON))
	}

	if len(apiRef.Parameters) > 0 {
		componentBuilder.WriteString(fmt.Sprintf("  parameters={%s}\n", parametersJSON))
	}

	if apiRef.RequestBody != nil {
		componentBuilder.WriteString(fmt.Sprintf("  requestBody={%s}\n", requestBodyJSON))
	}

	if len(apiRef.Responses) > 0 {
		componentBuilder.WriteString(fmt.Sprintf("  responses={%s}\n", responsesJSON))
	}

	if len(apiRef.Security) > 0 {
		componentBuilder.WriteString(fmt.Sprintf("  security={%s}\n", securityJSON))
	}

	if apiRef.SecuritySchemas !=nil{
		componentBuilder.WriteString(fmt.Sprintf("  securitySchemas={%s}\n", securitySchemasJson))
	}

	if len(apiRef.Servers) > 0 {
		componentBuilder.WriteString(fmt.Sprintf("  servers={%s}\n", serversJSON))
	}

	if apiRef.ExternalDocs != nil {
		componentBuilder.WriteString(fmt.Sprintf("  externalDocs={%s}\n", externalDocsJSON))
	}

	componentBuilder.WriteString("/>")

	// Create the complete MDX file content
	mdxContent := fmt.Sprintf(`---
title: %s
description: %s
---

# %s

%s

`,
		escapeString(apiRef.Title),
		escapeString(getStringValue(apiRef.Description)),
		escapeString(apiRef.Title),
		componentBuilder.String())

	// Create and write to the file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(mdxContent); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}


