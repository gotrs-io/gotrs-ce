package contracts

import (
	"net/http"
)

// QueueContracts defines the API contracts for queue endpoints
var QueueContracts = []Contract{
	{
		Name:        "ListQueues",
		Description: "List all queues",
		Method:      "GET",
		Path:        "/api/v1/queues",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"queues": {
						Type: "array",
						Items: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"id": {Type: "number"},
								"name": {Type: "string"},
								"description": {Type: "string"},
								"valid_id": {Type: "number"},
								"group_access": {
									Type: "array",
									Items: &Schema{Type: "number"},
								},
							},
							Required: []string{"id", "name", "valid_id"},
						},
					},
					"total": {Type: "number"},
				},
				Required: []string{"queues", "total"},
			},
		},
	},
	{
		Name:        "GetQueue",
		Description: "Get single queue by ID",
		Method:      "GET",
		Path:        "/api/v1/queues/1",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"id": {Type: "number"},
					"name": {Type: "string"},
					"description": {Type: "string"},
					"valid_id": {Type: "number"},
					"group_access": {
						Type: "array",
						Items: &Schema{Type: "number"},
					},
				},
				Required: []string{"id", "name", "valid_id"},
			},
		},
	},
	{
		Name:        "CreateQueue",
		Description: "Create new queue",
		Method:      "POST",
		Path:        "/api/v1/queues",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
			"Content-Type":  "application/json",
		},
		Body: map[string]interface{}{
			"name":        "Test Queue",
			"description": "Test queue description",
			"group_access": []int{1, 2},
		},
		Expected: Response{
			Status: http.StatusCreated,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"id": {Type: "number"},
					"name": {Type: "string"},
					"description": {Type: "string"},
					"valid_id": {Type: "number"},
					"group_access": {
						Type: "array",
						Items: &Schema{Type: "number"},
					},
				},
				Required: []string{"id", "name", "valid_id"},
			},
		},
	},
	{
		Name:        "UpdateQueue",
		Description: "Update existing queue",
		Method:      "PUT",
		Path:        "/api/v1/queues/1",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
			"Content-Type":  "application/json",
		},
		Body: map[string]interface{}{
			"name":        "Updated Queue",
			"description": "Updated description",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"id": {Type: "number"},
					"name": {Type: "string"},
					"description": {Type: "string"},
					"valid_id": {Type: "number"},
				},
				Required: []string{"id", "name", "valid_id"},
			},
		},
	},
	{
		Name:        "DeleteQueue",
		Description: "Soft delete queue",
		Method:      "DELETE",
		Path:        "/api/v1/queues/10",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"message": {Type: "string"},
					"id": {Type: "number"},
				},
				Required: []string{"message", "id"},
			},
		},
	},
}

// PriorityContracts defines the API contracts for priority endpoints
var PriorityContracts = []Contract{
	{
		Name:        "ListPriorities",
		Description: "List all priorities",
		Method:      "GET",
		Path:        "/api/v1/priorities",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"priorities": {
						Type: "array",
						Items: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"id": {Type: "number"},
								"name": {Type: "string"},
								"valid_id": {Type: "number"},
							},
							Required: []string{"id", "name", "valid_id"},
						},
					},
					"total": {Type: "number"},
				},
				Required: []string{"priorities", "total"},
			},
		},
	},
	{
		Name:        "GetPriority",
		Description: "Get single priority by ID",
		Method:      "GET",
		Path:        "/api/v1/priorities/1",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"id": {Type: "number"},
					"name": {Type: "string"},
					"valid_id": {Type: "number"},
				},
				Required: []string{"id", "name", "valid_id"},
			},
		},
	},
	{
		Name:        "CreatePriority",
		Description: "Create new priority",
		Method:      "POST",
		Path:        "/api/v1/priorities",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
			"Content-Type":  "application/json",
		},
		Body: map[string]interface{}{
			"name": "Test Priority",
		},
		Expected: Response{
			Status: http.StatusCreated,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"id": {Type: "number"},
					"name": {Type: "string"},
					"valid_id": {Type: "number"},
				},
				Required: []string{"id", "name", "valid_id"},
			},
		},
	},
	{
		Name:        "UpdatePriority",
		Description: "Update existing priority",
		Method:      "PUT",
		Path:        "/api/v1/priorities/6",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
			"Content-Type":  "application/json",
		},
		Body: map[string]interface{}{
			"name": "Updated Priority",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"id": {Type: "number"},
					"name": {Type: "string"},
					"valid_id": {Type: "number"},
				},
				Required: []string{"id", "name", "valid_id"},
			},
		},
	},
	{
		Name:        "DeletePriority",
		Description: "Soft delete priority",
		Method:      "DELETE",
		Path:        "/api/v1/priorities/10",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"message": {Type: "string"},
					"id": {Type: "number"},
				},
				Required: []string{"message", "id"},
			},
		},
	},
}

// RegisterQueuePriorityContracts registers queue and priority contracts for testing
func RegisterQueuePriorityContracts() {
	for _, contract := range QueueContracts {
		RegisterContract(contract)
	}
	for _, contract := range PriorityContracts {
		RegisterContract(contract)
	}
}