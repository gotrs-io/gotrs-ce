package contracts

import (
	"net/http"
)

// ArticleContracts defines the API contracts for article endpoints
var ArticleContracts = []Contract{
	{
		Name:        "ListArticles",
		Description: "List all articles for a ticket",
		Method:      "GET",
		Path:        "/api/v1/tickets/1/articles",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"articles": {
						Type: "array",
						Items: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"id":                     {Type: "number"},
								"ticket_id":              {Type: "number"},
								"article_type_id":        {Type: "number"},
								"article_sender_type_id": {Type: "number"},
								"from_email":             {Type: "string"},
								"to_email":               {Type: "string"},
								"subject":                {Type: "string"},
								"body":                   {Type: "string"},
								"create_time":            {Type: "string"},
								"create_by":              {Type: "number"},
							},
							Required: []string{"id", "ticket_id", "subject", "body"},
						},
					},
					"total": {Type: "number"},
				},
				Required: []string{"articles", "total"},
			},
		},
	},
	{
		Name:        "GetArticle",
		Description: "Get single article by ID",
		Method:      "GET",
		Path:        "/api/v1/tickets/1/articles/1",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"id":                     {Type: "number"},
					"ticket_id":              {Type: "number"},
					"article_type_id":        {Type: "number"},
					"article_sender_type_id": {Type: "number"},
					"from_email":             {Type: "string"},
					"to_email":               {Type: "string"},
					"subject":                {Type: "string"},
					"body":                   {Type: "string"},
					"create_time":            {Type: "string"},
					"create_by":              {Type: "number"},
					"change_time":            {Type: "string"},
					"change_by":              {Type: "number"},
				},
				Required: []string{"id", "ticket_id", "subject", "body"},
			},
		},
	},
	{
		Name:        "GetArticleWithAttachments",
		Description: "Get article with attachment information",
		Method:      "GET",
		Path:        "/api/v1/tickets/1/articles/1?include_attachments=true",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"id":         {Type: "number"},
					"ticket_id":  {Type: "number"},
					"subject":    {Type: "string"},
					"body":       {Type: "string"},
					"attachments": {
						Type: "array",
						Items: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"id":           {Type: "number"},
								"filename":     {Type: "string"},
								"content_type": {Type: "string"},
								"size":         {Type: "number"},
							},
							Required: []string{"id", "filename", "content_type", "size"},
						},
					},
				},
				Required: []string{"id", "ticket_id", "subject", "body"},
			},
		},
	},
	{
		Name:        "CreateArticle",
		Description: "Create new article for ticket",
		Method:      "POST",
		Path:        "/api/v1/tickets/1/articles",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
			"Content-Type":  "application/json",
		},
		Body: map[string]interface{}{
			"subject":      "Test Article",
			"body":         "This is a test article body",
			"from_email":   "agent@example.com",
			"to_email":     "customer@example.com",
			"article_type": "email-external",
			"sender_type":  "agent",
			"is_visible":   true,
		},
		Expected: Response{
			Status: http.StatusCreated,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"id":                     {Type: "number"},
					"ticket_id":              {Type: "number"},
					"article_type_id":        {Type: "number"},
					"article_sender_type_id": {Type: "number"},
					"from_email":             {Type: "string"},
					"to_email":               {Type: "string"},
					"subject":                {Type: "string"},
					"body":                   {Type: "string"},
					"create_time":            {Type: "string"},
					"create_by":              {Type: "number"},
				},
				Required: []string{"id", "ticket_id", "subject", "body"},
			},
		},
	},
	{
		Name:        "UpdateArticle",
		Description: "Update existing article",
		Method:      "PUT",
		Path:        "/api/v1/tickets/1/articles/1",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
			"Content-Type":  "application/json",
		},
		Body: map[string]interface{}{
			"subject": "Updated Article Subject",
			"body":    "Updated article body content",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"id":        {Type: "number"},
					"ticket_id": {Type: "number"},
					"subject":   {Type: "string"},
					"body":      {Type: "string"},
				},
				Required: []string{"id", "ticket_id", "subject", "body"},
			},
		},
	},
	{
		Name:        "DeleteArticle",
		Description: "Delete article from ticket",
		Method:      "DELETE",
		Path:        "/api/v1/tickets/1/articles/2",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"message": {Type: "string"},
					"id":      {Type: "number"},
				},
				Required: []string{"message", "id"},
			},
		},
	},
}

// RegisterArticleContracts registers article contracts for testing
func RegisterArticleContracts() {
	for _, contract := range ArticleContracts {
		RegisterContract(contract)
	}
}