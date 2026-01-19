package schemas

type OpenapiData struct {
	Title           string                     `json:"title"`
	Description     *string                    `json:"description,omitempty"`
	Method          string                     `json:"method"`
	Path            string                     `json:"path"`
	OperationId     *string                    `json:"operationId,omitempty"`
	Tags            []string                   `json:"tags,omitempty"`
	Summary         *string                    `json:"summary,omitempty"`
	Deprecated      *bool                      `json:"deprecated,omitempty"`
	Parameters      []Parameter                `json:"parameters,omitempty"`
	RequestBody     *RequestBody               `json:"requestBody,omitempty"`
	Responses       map[string]ResponseContent `json:"responses,omitempty"`
	Security        []map[string][]string      `json:"security,omitempty"`
	SecuritySchemas map[string]any             `json:"securitySchemas,omitempty"`
	Servers         []Server                   `json:"servers,omitempty"`
	ExternalDocs    *ExternalDocs              `json:"externalDocs,omitempty"`
}

type Parameter struct {
	Name            string                 `json:"name"`
	In              string                 `json:"in"`
	Description     string                 `json:"description,omitempty"`
	Required        *bool                  `json:"required,omitempty"`
	Deprecated      *bool                  `json:"deprecated,omitempty"`
	AllowEmptyValue *bool                  `json:"allowEmptyValue,omitempty"`
	Style           string                 `json:"style,omitempty"`
	Explode         *bool                  `json:"explode,omitempty"`
	AllowReserved   *bool                  `json:"allowReserved,omitempty"`
	Schema          ParameterSchema        `json:"schema"`
	Example         *interface{}           `json:"example,omitempty"`
	Examples        map[string]interface{} `json:"examples,omitempty"`
}

type ParameterSchema struct {
	Type        string        `json:"type"`
	Format      *string       `json:"format,omitempty"`
	Description *string       `json:"description,omitempty"`
	Default     *interface{}  `json:"default,omitempty"`
	Example     *interface{}  `json:"example,omitempty"`
	Examples    []interface{} `json:"examples,omitempty"`
	Required    *bool         `json:"required,omitempty"`
	Nullable    *bool         `json:"nullable,omitempty"`
	ReadOnly    *bool         `json:"readOnly,omitempty"`
	WriteOnly   *bool         `json:"writeOnly,omitempty"`
	Deprecated  *bool         `json:"deprecated,omitempty"`

	Minimum          *float64 `json:"minimum,omitempty"`
	Maximum          *float64 `json:"maximum,omitempty"`
	ExclusiveMinimum *bool    `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *bool    `json:"exclusiveMaximum,omitempty"`

	MinLength *float64 `json:"minLength,omitempty"`
	MaxLength *float64 `json:"maxLength,omitempty"`
	Pattern   *string  `json:"pattern,omitempty"`

	MinItems    *float64         `json:"minItems,omitempty"`
	MaxItems    *float64         `json:"maxItems,omitempty"`
	UniqueItems *bool            `json:"uniqueItems,omitempty"`
	Items       *ParameterSchema `json:"items,omitempty"`

	Enum []interface{} `json:"enum,omitempty"`

	Properties                  map[string]ParameterSchema `json:"properties,omitempty"`
	RequiredFields              []string                   `json:"required,omitempty"`
	AdditionalProperties        *ParameterSchema           `json:"additionalProperties,omitempty"`
	AdditionalPropertiesAllowed *bool                      `json:"additionalPropertiesAllowed,omitempty"`
	OneOf                       []ParameterSchema          `json:"oneOf,omitempty"`
	AnyOf                       []ParameterSchema          `json:"anyOf,omitempty"`
	AllOf                       []ParameterSchema          `json:"allOf,omitempty"`
	Not                         *ParameterSchema           `json:"not,omitempty"`
}

type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Content     map[string]MediaType `json:"content"`
	Required    *bool                `json:"required,omitempty"`
}

type MediaType struct {
	Schema   *ParameterSchema       `json:"schema,omitempty"`
	Example  *interface{}           `json:"example,omitempty"`
	Examples map[string]interface{} `json:"examples,omitempty"`
}

type ResponseContent struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
	Headers     map[string]Header    `json:"headers,omitempty"`
}

type Header struct {
	Description string           `json:"description,omitempty"`
	Required    *bool            `json:"required,omitempty"`
	Deprecated  *bool            `json:"deprecated,omitempty"`
	Style       string           `json:"style,omitempty"`
	Explode     *bool            `json:"explode,omitempty"`
	Schema      *ParameterSchema `json:"schema,omitempty"`
	Example     *interface{}     `json:"example,omitempty"`
}

type Server struct {
	URL         string                    `json:"url"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

type ExternalDocs struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
}

