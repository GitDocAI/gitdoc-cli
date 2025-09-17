package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/YHVCorp/gitdocai-cli/schemas"
	"github.com/getkin/kin-openapi/openapi3"
)

var openapiCache sync.Map

// Helper to safely get a string from a map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// Helper to safely get a map from a map
func getMap(m map[string]interface{}, key string) map[string]interface{} {
	if val, ok := m[key].(map[string]interface{}); ok {
		return val
	}
	return nil
}

// schemas ref resolver
var refResolutionCache sync.Map

func resolveSchemaRef(ref string, components map[string]interface{}, visitedRefs map[string]bool) map[string]interface{} {
	if !strings.HasPrefix(ref, "#/components/schemas/") {
		return nil
	}

	if visitedRefs == nil {
		visitedRefs = make(map[string]bool)
	}

	if visitedRefs[ref] {
		return map[string]interface{}{
			"type":        "object",
			"description": fmt.Sprintf("Circular reference to %s", ref),
		}
	}

	schemaName := strings.TrimPrefix(ref, "#/components/schemas/")

	if cached, ok := refResolutionCache.Load(ref); ok {
		return cached.(map[string]interface{})
	}

	schemas := getMap(components, "schemas")
	if schemas == nil {
		return nil
	}

	referencedSchema := getMap(schemas, schemaName)
	if referencedSchema == nil {
		return nil
	}

	visitedRefs[ref] = true

	expandedSchema := expandSchemaReferences(referencedSchema, components, visitedRefs)

	delete(visitedRefs, ref)

	refResolutionCache.Store(ref, expandedSchema)

	return expandedSchema
}

// schema ref expander
func expandSchemaReferences(schema map[string]interface{}, components map[string]interface{}, visitedRefs map[string]bool) map[string]interface{} {
	if schema == nil {
		return nil
	}

	if refVal := getString(schema, "$ref"); refVal != "" {
		return resolveSchemaRef(refVal, components, visitedRefs)
	}

	expandedSchema := make(map[string]interface{})
	for k, v := range schema {
		expandedSchema[k] = v
	}

	if properties := getMap(schema, "properties"); properties != nil {
		expandedProps := make(map[string]interface{})
		for propName, propSchema := range properties {
			if propMap, ok := propSchema.(map[string]interface{}); ok {
				expandedProps[propName] = expandSchemaReferences(propMap, components, visitedRefs)
			} else {
				expandedProps[propName] = propSchema
			}
		}
		expandedSchema["properties"] = expandedProps
	}

	if items := getMap(schema, "items"); items != nil {
		expandedSchema["items"] = expandSchemaReferences(items, components, visitedRefs)
	}

	if additionalProps := getMap(schema, "additionalProperties"); additionalProps != nil {
		expandedSchema["additionalProperties"] = expandSchemaReferences(additionalProps, components, visitedRefs)
	}

	if allOf := getSlice(schema, "allOf"); allOf != nil {
		expandedAllOf := make([]interface{}, 0, len(allOf))
		for _, subSchema := range allOf {
			if subMap, ok := subSchema.(map[string]interface{}); ok {
				expandedAllOf = append(expandedAllOf, expandSchemaReferences(subMap, components, visitedRefs))
			} else {
				expandedAllOf = append(expandedAllOf, subSchema)
			}
		}
		expandedSchema["allOf"] = expandedAllOf
	}

	if oneOf := getSlice(schema, "oneOf"); oneOf != nil {
		expandedOneOf := make([]interface{}, 0, len(oneOf))
		for _, subSchema := range oneOf {
			if subMap, ok := subSchema.(map[string]interface{}); ok {
				expandedOneOf = append(expandedOneOf, expandSchemaReferences(subMap, components, visitedRefs))
			} else {
				expandedOneOf = append(expandedOneOf, subSchema)
			}
		}
		expandedSchema["oneOf"] = expandedOneOf
	}

	if anyOf := getSlice(schema, "anyOf"); anyOf != nil {
		expandedAnyOf := make([]interface{}, 0, len(anyOf))
		for _, subSchema := range anyOf {
			if subMap, ok := subSchema.(map[string]interface{}); ok {
				expandedAnyOf = append(expandedAnyOf, expandSchemaReferences(subMap, components, visitedRefs))
			} else {
				expandedAnyOf = append(expandedAnyOf, subSchema)
			}
		}
		expandedSchema["anyOf"] = expandedAnyOf
	}

	if not := getMap(schema, "not"); not != nil {
		expandedSchema["not"] = expandSchemaReferences(not, components, visitedRefs)
	}

	return expandedSchema
}

// Helper to safely get a slice from a map
func getSlice(m map[string]interface{}, key string) []interface{} {
	if val, ok := m[key].([]interface{}); ok {
		return val
	}
	return nil
}

// Helper to safely get a bool from a map
func getBool(m map[string]interface{}, key string) *bool {
	if val, ok := m[key].(bool); ok {
		return &val
	}
	return nil
}

// Helper to safely get a float64 from a map (for numbers)
func getFloat64(m map[string]interface{}, key string) *float64 {
	if val, ok := m[key].(float64); ok {
		return &val
	}
	return nil
}

func cacheSet(key string, value map[string]interface{}) {
	openapiCache.Store(key, value)
}

func cacheGet(key string) (map[string]interface{}, bool) {
	val, ok := openapiCache.Load(key)
	if !ok {
		return nil, false
	}
	return val.(map[string]interface{}), true
}

// extracts full information abot a parameter
func extractParameterSchema(schema map[string]interface{}, components map[string]interface{}) schemas.ParameterSchema {
	expandedSchema := expandSchemaReferences(schema, components, nil)

	ps := schemas.ParameterSchema{
		Type: getString(expandedSchema, "type"),
	}

	if format := getString(expandedSchema, "format"); format != "" {
		ps.Format = &format
	}

	if description := getString(expandedSchema, "description"); description != "" {
		ps.Description = &description
	}

	ps.Required = getBool(expandedSchema, "required")
	ps.Nullable = getBool(expandedSchema, "nullable")
	ps.ReadOnly = getBool(expandedSchema, "readOnly")
	ps.WriteOnly = getBool(expandedSchema, "writeOnly")
	ps.Deprecated = getBool(expandedSchema, "deprecated")

	if defaultVal := expandedSchema["default"]; defaultVal != nil {
		ps.Default = &defaultVal
	}

	if example := expandedSchema["example"]; example != nil {
		ps.Example = &example
	}

	if examples := getSlice(expandedSchema, "examples"); examples != nil {
		ps.Examples = examples
	}

	ps.Minimum = getFloat64(expandedSchema, "minimum")
	ps.Maximum = getFloat64(expandedSchema, "maximum")
	ps.ExclusiveMinimum = getBool(expandedSchema, "exclusiveMinimum")
	ps.ExclusiveMaximum = getBool(expandedSchema, "exclusiveMaximum")
	ps.MinLength = getFloat64(expandedSchema, "minLength")
	ps.MaxLength = getFloat64(expandedSchema, "maxLength")
	ps.MinItems = getFloat64(expandedSchema, "minItems")
	ps.MaxItems = getFloat64(expandedSchema, "maxItems")
	ps.UniqueItems = getBool(expandedSchema, "uniqueItems")

	if pattern := getString(expandedSchema, "pattern"); pattern != "" {
		ps.Pattern = &pattern
	}

	if enumVals := getSlice(expandedSchema, "enum"); enumVals != nil {
		ps.Enum = enumVals
	}

	if items := getMap(expandedSchema, "items"); items != nil {
		itemSchema := extractParameterSchema(items, components)
		ps.Items = &itemSchema
	}

	if properties := getMap(expandedSchema, "properties"); properties != nil {
		ps.Properties = make(map[string]schemas.ParameterSchema)
		for propName, propSchema := range properties {
			if propSchemaMap, ok := propSchema.(map[string]interface{}); ok {
				ps.Properties[propName] = extractParameterSchema(propSchemaMap, components)
			}
		}
	}

	if additionalProps := expandedSchema["additionalProperties"]; additionalProps != nil {
		if addPropsMap, ok := additionalProps.(map[string]interface{}); ok {
			addPropsSchema := extractParameterSchema(addPropsMap, components)
			ps.AdditionalProperties = &addPropsSchema
		} else if addPropsBool, ok := additionalProps.(bool); ok {
			ps.AdditionalPropertiesAllowed = &addPropsBool
		}
	}

	if allOf := getSlice(expandedSchema, "allOf"); allOf != nil {
		ps.AllOf = make([]schemas.ParameterSchema, 0, len(allOf))
		for _, subSchema := range allOf {
			if subSchemaMap, ok := subSchema.(map[string]interface{}); ok {
				ps.AllOf = append(ps.AllOf, extractParameterSchema(subSchemaMap, components))
			}
		}
	}

	if oneOf := getSlice(expandedSchema, "oneOf"); oneOf != nil {
		ps.OneOf = make([]schemas.ParameterSchema, 0, len(oneOf))
		for _, subSchema := range oneOf {
			if subSchemaMap, ok := subSchema.(map[string]interface{}); ok {
				ps.OneOf = append(ps.OneOf, extractParameterSchema(subSchemaMap, components))
			}
		}
	}

	if anyOf := getSlice(expandedSchema, "anyOf"); anyOf != nil {
		ps.AnyOf = make([]schemas.ParameterSchema, 0, len(anyOf))
		for _, subSchema := range anyOf {
			if subSchemaMap, ok := subSchema.(map[string]interface{}); ok {
				ps.AnyOf = append(ps.AnyOf, extractParameterSchema(subSchemaMap, components))
			}
		}
	}

	if required := getSlice(expandedSchema, "required"); required != nil {
		requiredFields := make([]string, 0, len(required))
		for _, req := range required {
			if reqStr, ok := req.(string); ok {
				requiredFields = append(requiredFields, reqStr)
			}
		}
		ps.RequiredFields = requiredFields
	}

	return ps
}

func ClearRefResolutionCache() {
	refResolutionCache.Range(func(key, value interface{}) bool {
		refResolutionCache.Delete(key)
		return true
	})
}

func PreloadAndExpandAllSchemas(components map[string]interface{}) map[string]interface{} {
	schemas := getMap(components, "schemas")
	if schemas == nil {
		return nil
	}

	expandedSchemas := make(map[string]interface{})
	for schemaName, schemaData := range schemas {
		if schemaMap, ok := schemaData.(map[string]interface{}); ok {
			expandedSchemas[schemaName] = expandSchemaReferences(schemaMap, components, nil)
		}
	}

	return expandedSchemas
}
func extractRequestBody(requestBody, components map[string]interface{}) *schemas.RequestBody {
	if requestBody == nil {
		return nil
	}

	rb := &schemas.RequestBody{
		Description: getString(requestBody, "description"),
		Required:    getBool(requestBody, "required"),
	}

	if content := getMap(requestBody, "content"); content != nil {
		rb.Content = make(map[string]schemas.MediaType)
		for mediaType, mediaContent := range content {
			if mediaMap := getMap(map[string]interface{}{mediaType: mediaContent}, mediaType); mediaMap != nil {
				mt := schemas.MediaType{}
				if schema := getMap(mediaMap, "schema"); schema != nil {
					schemaData := extractParameterSchema(schema, components)
					mt.Schema = &schemaData
				}
				if example := mediaMap["example"]; example != nil {
					mt.Example = &example
				}
				if examples := getMap(mediaMap, "examples"); examples != nil {
					mt.Examples = examples
				}
				rb.Content[mediaType] = mt
			} else if mediaContentMap, ok := mediaContent.(map[string]interface{}); ok {
				mt := schemas.MediaType{}
				if schema := getMap(mediaContentMap, "schema"); schema != nil {
					schemaData := extractParameterSchema(schema, components)
					mt.Schema = &schemaData
				}
				if example := mediaContentMap["example"]; example != nil {
					mt.Example = &example
				}
				if examples := getMap(mediaContentMap, "examples"); examples != nil {
					mt.Examples = examples
				}
				rb.Content[mediaType] = mt
			}
		}
	}

	return rb
}

func ExtractRequestInfoFromOpenapi(title, path, method, request string) (*schemas.OpenapiData, error) {
	openapiDic, err := ExtractOpenapiConfig(path)
	if err != nil {
		return nil, err
	}

	components := getMap(openapiDic, "components")

	// paths
	paths := getMap(openapiDic, "paths")
	if paths == nil {
		return nil, fmt.Errorf("no paths found in OpenAPI")
	}

	pathItem := getMap(paths, request)
	if pathItem == nil {
		return nil, fmt.Errorf("path %s not found", request)
	}

	operation := getMap(pathItem, strings.ToLower(method))
	if operation == nil {
		return nil, fmt.Errorf("method %s not found for path %s", method, request)
	}

	//endpoint basic information
	result := &schemas.OpenapiData{
		Title:      title,
		Method:     strings.ToUpper(method),
		Path:       request,
		Parameters: []schemas.Parameter{},
		Responses:  make(map[string]schemas.ResponseContent),
	}

	//aditional information
	if desc := getString(operation, "description"); desc != "" {
		result.Description = &desc
	}

	if opId := getString(operation, "operationId"); opId != "" {
		result.OperationId = &opId
	}

	if tags := getSlice(operation, "tags"); tags != nil {
		tagStrings := make([]string, 0, len(tags))
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				tagStrings = append(tagStrings, tagStr)
			}
		}
		result.Tags = tagStrings
	}

	result.Deprecated = getBool(operation, "deprecated")

	// External docs
	if extDocs := getMap(operation, "externalDocs"); extDocs != nil {
		result.ExternalDocs = &schemas.ExternalDocs{
			Description: getString(extDocs, "description"),
			URL:         getString(extDocs, "url"),
		}
	}

	// Parameters
	if rawParams := getSlice(operation, "parameters"); rawParams != nil {
		for _, p := range rawParams {
			if pmap, ok := p.(map[string]interface{}); ok {
				param := schemas.Parameter{
					Name:            getString(pmap, "name"),
					In:              getString(pmap, "in"),
					Description:     getString(pmap, "description"),
					Required:        getBool(pmap, "required"),
					Deprecated:      getBool(pmap, "deprecated"),
					AllowEmptyValue: getBool(pmap, "allowEmptyValue"),
					Style:           getString(pmap, "style"),
					Explode:         getBool(pmap, "explode"),
					AllowReserved:   getBool(pmap, "allowReserved"),
				}

				if schema := getMap(pmap, "schema"); schema != nil {
					param.Schema = extractParameterSchema(schema, components)
				}

				if example := pmap["example"]; example != nil {
					param.Example = &example
				}

				if examples := getMap(pmap, "examples"); examples != nil {
					param.Examples = examples
				}

				result.Parameters = append(result.Parameters, param)
			}
		}
	}

	// Request Body
	if requestBody := getMap(operation, "requestBody"); requestBody != nil {
		result.RequestBody = extractRequestBody(requestBody, components)
	}

	// Responses
	if rawRes := getMap(operation, "responses"); rawRes != nil {
		for code, r := range rawRes {
			if rmap, ok := r.(map[string]interface{}); ok {
				response := schemas.ResponseContent{
					Description: getString(rmap, "description"),
				}

				if content := getMap(rmap, "content"); content != nil {
					response.Content = make(map[string]schemas.MediaType)
					for mediaType, mediaContent := range content {
						if mediaMap, ok := mediaContent.(map[string]interface{}); ok {
							mt := schemas.MediaType{}
							if schema := getMap(mediaMap, "schema"); schema != nil {
								schemaData := extractParameterSchema(schema, components)
								mt.Schema = &schemaData
							}
							if example := mediaMap["example"]; example != nil {
								mt.Example = &example
							}
							if examples := getMap(mediaMap, "examples"); examples != nil {
								mt.Examples = examples
							}
							response.Content[mediaType] = mt
						}
					}
				} else if rmap["content"] != nil {
					response.Content = rmap["content"].(map[string]schemas.MediaType)
				}

				if headers := getMap(rmap, "headers"); headers != nil {
					response.Headers = make(map[string]schemas.Header)
					for headerName, headerContent := range headers {
						if headerMap, ok := headerContent.(map[string]interface{}); ok {
							header := schemas.Header{
								Description: getString(headerMap, "description"),
								Required:    getBool(headerMap, "required"),
								Deprecated:  getBool(headerMap, "deprecated"),
								Style:       getString(headerMap, "style"),
								Explode:     getBool(headerMap, "explode"),
							}
							if schema := getMap(headerMap, "schema"); schema != nil {
								headerSchema := extractParameterSchema(schema, components)
								header.Schema = &headerSchema
							}
							if example := headerMap["example"]; example != nil {
								header.Example = &example
							}
							response.Headers[headerName] = header
						}
					}
				}

				result.Responses[code] = response
			}
		}
	}

	// Security
	op_security := getSlice(operation, "security")
	global_security := getSlice(openapiDic, "security")

	if op_security != nil || global_security != nil {
		security := append(op_security, global_security...)
		result.Security = make([]map[string][]string, 0, len(security))
		for _, sec := range security {
			if secMap, ok := sec.(map[string]interface{}); ok {
				securityReq := make(map[string][]string)
				for secName, secScopes := range secMap {
					if scopesList, ok := secScopes.([]interface{}); ok {
						scopes := make([]string, 0, len(scopesList))
						for _, scope := range scopesList {
							if scopeStr, ok := scope.(string); ok {
								scopes = append(scopes, scopeStr)
							}
						}
						securityReq[secName] = scopes
					} else {
						securityReq[secName] = []string{}
					}
				}
				result.Security = append(result.Security, securityReq)
			}
		}
	}
	//security schemas
	if secSchemes := getMap(components, "securitySchemes"); secSchemes != nil {
		result.SecuritySchemas = secSchemes
	}

	// Servers
	if servers := getSlice(operation, "servers"); servers != nil {
		result.Servers = make([]schemas.Server, 0, len(servers))
		for _, srv := range servers {
			if srvMap, ok := srv.(map[string]interface{}); ok {
				server := schemas.Server{
					URL:         getString(srvMap, "url"),
					Description: getString(srvMap, "description"),
				}
				if variables := getMap(srvMap, "variables"); variables != nil {
					server.Variables = make(map[string]schemas.ServerVariable)
					for varName, varContent := range variables {
						if varMap, ok := varContent.(map[string]interface{}); ok {
							serverVar := schemas.ServerVariable{
								Default:     getString(varMap, "default"),
								Description: getString(varMap, "description"),
							}
							if enumVals := getSlice(varMap, "enum"); enumVals != nil {
								enumStrings := make([]string, 0, len(enumVals))
								for _, enumVal := range enumVals {
									if enumStr, ok := enumVal.(string); ok {
										enumStrings = append(enumStrings, enumStr)
									}
								}
								serverVar.Enum = enumStrings
							}
							server.Variables[varName] = serverVar
						}
					}
				}
				result.Servers = append(result.Servers, server)
			}
		}
	}

	return result, nil
}

func ExtractOpenapiConfig(path string) (map[string]any, error) {
	if cached, ok := cacheGet(path); ok {
		return cached, nil
	}

	loader := &openapi3.Loader{Context: context.Background()}
	openapiDoc, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI file: %w", err)
	}

	openapiMap, err := json.Marshal(openapiDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to convert OpenAPI document to map: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(openapiMap, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OpenAPI document: %w", err)
	}

	cacheSet(path, result)
	return result, nil
}
