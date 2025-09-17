package schemas

type OpenapiData struct {
	Title           string                     `json:"title"`
	Description     *string                    `json:"description,omitempty"`
	Method          string                     `json:"method"`
	Path            string                     `json:"path"`
	OperationId     *string                    `json:"operation_id,omitempty"`
	Tags            []string                   `json:"tags,omitempty"`
	Summary         *string                    `json:"summary,omitempty"`
	Deprecated      *bool                      `json:"deprecated,omitempty"`
	Parameters      []Parameter                `json:"parameters"`
	RequestBody     *RequestBody               `json:"request_body,omitempty"`
	Responses       map[string]ResponseContent `json:"responses"`
	Security        []map[string][]string      `json:"security,omitempty"`
	SecuritySchemas map[string]any             `json:"securitySchemes,omitempty"`
	Servers         []Server                   `json:"servers,omitempty"`
	ExternalDocs    *ExternalDocs              `json:"external_docs,omitempty"`
}

type Parameter struct {
	Name            string                 `json:"name"`
	In              string                 `json:"in"`
	Description     string                 `json:"description,omitempty"`
	Required        *bool                  `json:"required,omitempty"`
	Deprecated      *bool                  `json:"deprecated,omitempty"`
	AllowEmptyValue *bool                  `json:"allow_empty_value,omitempty"`
	Style           string                 `json:"style,omitempty"`
	Explode         *bool                  `json:"explode,omitempty"`
	AllowReserved   *bool                  `json:"allow_reserved,omitempty"`
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
	ReadOnly    *bool         `json:"read_only,omitempty"`
	WriteOnly   *bool         `json:"write_only,omitempty"`
	Deprecated  *bool         `json:"deprecated,omitempty"`

	Minimum          *float64 `json:"minimum,omitempty"`
	Maximum          *float64 `json:"maximum,omitempty"`
	ExclusiveMinimum *bool    `json:"exclusive_minimum,omitempty"`
	ExclusiveMaximum *bool    `json:"exclusive_maximum,omitempty"`

	MinLength *float64 `json:"min_length,omitempty"`
	MaxLength *float64 `json:"max_length,omitempty"`
	Pattern   *string  `json:"pattern,omitempty"`

	MinItems    *float64         `json:"min_items,omitempty"`
	MaxItems    *float64         `json:"max_items,omitempty"`
	UniqueItems *bool            `json:"unique_items,omitempty"`
	Items       *ParameterSchema `json:"items,omitempty"`

	Enum []interface{} `json:"enum,omitempty"`

	Properties                  map[string]ParameterSchema `json:"properties,omitempty"`
	RequiredFields              []string                   `json:"required_fields,omitempty"`
	AdditionalProperties        *ParameterSchema           `json:"additional_properties,omitempty"`
	AdditionalPropertiesAllowed *bool                      `json:"additional_properties_allowed,omitempty"`
	OneOf                       []ParameterSchema          `json:"one_of,omitempty"`
	AnyOf                       []ParameterSchema          `json:"any_of,omitempty"`
	AllOf                       []ParameterSchema          `json:"all_of,omitempty"`
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
