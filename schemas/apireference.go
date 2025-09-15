package schemas


type ApiReference struct{
  Title      string   `json:"title"`
  Method     string   `json:"method"` //"GET" | "POST" | "PUT" | "DELETE";
  Path       string   `json:"path"`
  Parameters []Parameter `json:"parameters"`
  Responses map[string]ResponseContent `json:"responses"`
  TryItBaseUrl *string `json:"try_it_base_url"`
}



type Parameter struct{
  Name string `json:"name"`
  In string `json:"in"`
  Description string `json:"description"`
  Schema ParameterSchema `json:"schema"`
}

type ParameterSchema struct{
 Type    string   `json:"type"`
 Format  *string  `json:"format"`
}

type ResponseContent struct {
  Description string `json:"description"`
  Content     any    `json:"content"`
}
