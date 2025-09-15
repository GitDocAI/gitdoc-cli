package utils
import (
	"context"
	"encoding/json"
	"strings"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/YHVCorp/gitdocai-cli/schemas"
	"sync"
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

func ExtractRequestInfoFromOpenapi(path, method, request string, tryitbaseurl *string) (*schemas.OpenapiData, error) {
    openapiDic, err := ExtractOpenapiConfig(path)
    if err != nil {
        return nil, err
    }

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

    // parameters
    var params []schemas.Parameter
    if rawParams, ok := operation["parameters"].([]interface{}); ok {
        for _, p := range rawParams {
            if pmap, ok := p.(map[string]interface{}); ok {
                var ps schemas.ParameterSchema
                if schema := getMap(pmap, "schema"); schema != nil {
                    ps.Type = getString(schema, "type")
                    if f := getString(schema, "format"); f != "" {
                        ps.Format = &f
                    }
                }

                name := getString(pmap, "name")
                in := getString(pmap, "in")
                description := getString(pmap, "description")

                params = append(params, schemas.Parameter{
                    Name:        name,
                    In:          in,
                    Description: description,
                    Schema:      ps,
                })
            }
        }
    }

    // responses
    responses := make(map[string]schemas.ResponseContent)
    if rawRes, ok := operation["responses"].(map[string]interface{}); ok {
        for code, r := range rawRes {
            if rmap, ok := r.(map[string]interface{}); ok {
                desc := getString(rmap, "description")
                var content any
                if c, ok := rmap["content"]; ok {
                    content = c
                }
                responses[code] = schemas.ResponseContent{
                    Description: desc,
                    Content:     content,
                }
            }
        }
    }

    // title (summary or operationId)
    var title string
    if s, ok := operation["summary"].(string); ok {
        title = s
    } else if opid, ok := operation["operationId"].(string); ok {
        title = opid
    } else {
        title = fmt.Sprintf("%s %s", method, request)
    }

    // base URL
    var tryIt *string = tryitbaseurl
	if tryIt ==nil{
		if servers, ok := openapiDic["servers"].([]interface{}); ok && len(servers) > 0 {
			if first := getMap(servers[0].(map[string]interface{}), ""); first != nil {
				if url := getString(first, "url"); url != "" {
					tryIt = &url
				}
			}
		}
	}

    return &schemas.OpenapiData{
        Title:        title,
        Method:       strings.ToUpper(method),
        Path:         request,
        Parameters:   params,
        Responses:    responses,
        TryItBaseUrl: tryIt,
    }, nil
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
