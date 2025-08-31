package contracts

import (
	"net/http"
)

// SearchContracts defines the API contracts for search endpoints
var SearchContracts = []Contract{
	{
		Name:        "SearchAll",
		Description: "Search across all entity types",
		Method:      "POST",
		Path:        "/api/v1/search",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
			"Content-Type":  "application/json",
		},
		Body: map[string]interface{}{
			"query": "network issue",
			"types": []string{"ticket", "article", "customer"},
			"limit": 10,
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"query":      {Type: "string"},
					"total_hits": {Type: "number"},
					"took_ms":    {Type: "number"},
					"hits": {
						Type: "array",
						Items: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"id":      {Type: "string"},
								"type":    {Type: "string"},
								"score":   {Type: "number"},
								"title":   {Type: "string"},
								"content": {Type: "string"},
								"metadata": {Type: "object"},
								"highlights": {
									Type: "object",
									Properties: map[string]Schema{
										"*": {
											Type:  "array",
											Items: &Schema{Type: "string"},
										},
									},
								},
							},
							Required: []string{"id", "type", "title"},
						},
					},
				},
				Required: []string{"query", "total_hits", "took_ms", "hits"},
			},
		},
	},
	{
		Name:        "SearchTickets",
		Description: "Search only tickets",
		Method:      "POST",
		Path:        "/api/v1/search",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
			"Content-Type":  "application/json",
		},
		Body: map[string]interface{}{
			"query": "email",
			"types": []string{"ticket"},
			"filters": map[string]string{
				"queue_id": "1",
				"state_id": "1",
			},
			"limit":  20,
			"offset": 0,
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"hits": {
						Type: "array",
						Items: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"type": {
									Type: "string",
									Enum: []interface{}{"ticket"},
								},
							},
						},
					},
				},
			},
		},
	},
	{
		Name:        "SearchWithHighlighting",
		Description: "Search with result highlighting",
		Method:      "POST",
		Path:        "/api/v1/search",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
			"Content-Type":  "application/json",
		},
		Body: map[string]interface{}{
			"query":     "password reset",
			"types":     []string{"ticket", "article"},
			"highlight": true,
			"limit":     5,
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"hits": {
						Type: "array",
						Items: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"highlights": {Type: "object"},
							},
						},
					},
				},
			},
		},
	},
	{
		Name:        "SearchWithPagination",
		Description: "Search with pagination",
		Method:      "POST",
		Path:        "/api/v1/search",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
			"Content-Type":  "application/json",
		},
		Body: map[string]interface{}{
			"query":  "customer",
			"types":  []string{"customer"},
			"offset": 10,
			"limit":  10,
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"hits": {
						Type: "array",
						MaxItems: func(v interface{}) error {
							if hits, ok := v.([]interface{}); ok && len(hits) > 10 {
								return &ValidationError{Field: "hits", Message: "Too many results"}
							}
							return nil
						},
					},
				},
			},
		},
	},
	{
		Name:        "SearchSuggestions",
		Description: "Get search suggestions",
		Method:      "GET",
		Path:        "/api/v1/search/suggestions?q=net&type=ticket",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"query": {Type: "string"},
					"type":  {Type: "string"},
					"suggestions": {
						Type:  "array",
						Items: &Schema{Type: "string"},
					},
				},
				Required: []string{"query", "type", "suggestions"},
			},
		},
	},
	{
		Name:        "SearchHealth",
		Description: "Check search backend health",
		Method:      "GET",
		Path:        "/api/v1/search/health",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"status": {
						Type: "string",
						Enum: []interface{}{"healthy", "unhealthy"},
					},
					"backend": {Type: "string"},
				},
				Required: []string{"status", "backend"},
			},
		},
	},
	{
		Name:        "Reindex",
		Description: "Trigger search index rebuild",
		Method:      "POST",
		Path:        "/api/v1/search/reindex",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
			"Content-Type":  "application/json",
		},
		Body: map[string]interface{}{
			"types": []string{"ticket", "article"},
			"force": false,
		},
		Expected: Response{
			StatusOptions: []int{http.StatusOK, http.StatusAccepted},
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"message": {Type: "string"},
					"backend": {Type: "string"},
				},
				Required: []string{"message"},
			},
		},
	},
	{
		Name:        "EmptySearchQuery",
		Description: "Search with empty query should fail",
		Method:      "POST",
		Path:        "/api/v1/search",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
			"Content-Type":  "application/json",
		},
		Body: map[string]interface{}{
			"query": "",
			"types": []string{"ticket"},
		},
		Expected: Response{
			Status: http.StatusBadRequest,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"error": {Type: "string"},
				},
				Required: []string{"error"},
			},
		},
	},
}

// RegisterSearchContracts registers search contracts for testing
func RegisterSearchContracts() {
	for _, contract := range SearchContracts {
		RegisterContract(contract)
	}
}