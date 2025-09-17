package api

type PostRequest struct {
	Events     []map[string]interface{} `json:"events,omitempty"`
	Host       string                   `json:"host,omitempty"`
	Source     string                   `json:"source,omitempty"`
	SourceType string                   `json:"source_type,omitempty"`
	Index      string                   `json:"index,omitempty"`
}

type PostResponse struct {
	Count int `json:"count"`
}
