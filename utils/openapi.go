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
    paths, ok := openapiDic["paths"].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("no paths found in OpenAPI")
    }

    pathItem, ok := paths[request].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("path %s not found", request)
    }

    operation, ok := pathItem[strings.ToLower(method)].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("method %s not found for path %s", method, request)
    }

    // parameters
    var params []schemas.Parameter
    if rawParams, ok := operation["parameters"].([]interface{}); ok {
        for _, p := range rawParams {
            if pmap, ok := p.(map[string]interface{}); ok {
                var ps schemas.ParameterSchema
                if schema, ok := pmap["schema"].(map[string]interface{}); ok {
                    ps.Type, _ = schema["type"].(string)
                    if f, ok := schema["format"].(string); ok {
                        ps.Format = &f
                    }
                }

                params = append(params, schemas.Parameter{
                    Name:        pmap["name"].(string),
                    In:          pmap["in"].(string),
                    Description: pmap["description"].(string),
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
                desc, _ := rmap["description"].(string)
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
			if first, ok := servers[0].(map[string]interface{}); ok {
				if url, ok := first["url"].(string); ok {
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
