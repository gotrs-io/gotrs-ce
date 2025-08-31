package contracts

import (
	"net/http"
)

// TicketStateContracts defines the API contracts for ticket state endpoints
var TicketStateContracts = []Contract{
	{
		Name:        "ListTicketStates",
		Description: "List all ticket states",
		Method:      "GET",
		Path:        "/api/v1/ticket-states",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"states": {
						Type: "array",
						Items: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"id":        {Type: "number"},
								"name":      {Type: "string"},
								"type_id":   {Type: "number"},
								"valid_id":  {Type: "number"},
								"type_name": {Type: "string"},
							},
							Required: []string{"id", "name", "type_id", "valid_id"},
						},
					},
					"total": {Type: "number"},
				},
				Required: []string{"states", "total"},
			},
		},
	},
	{
		Name:        "GetTicketState",
		Description: "Get single ticket state by ID",
		Method:      "GET",
		Path:        "/api/v1/ticket-states/1",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"id":       {Type: "number"},
					"name":     {Type: "string"},
					"type_id":  {Type: "number"},
					"valid_id": {Type: "number"},
				},
				Required: []string{"id", "name", "type_id", "valid_id"},
			},
		},
	},
	{
		Name:        "CreateTicketState",
		Description: "Create new ticket state",
		Method:      "POST",
		Path:        "/api/v1/ticket-states",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
			"Content-Type":  "application/json",
		},
		Body: map[string]interface{}{
			"name":    "Test State",
			"type_id": 1, // open type
		},
		Expected: Response{
			Status: http.StatusCreated,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"id":       {Type: "number"},
					"name":     {Type: "string"},
					"type_id":  {Type: "number"},
					"valid_id": {Type: "number"},
				},
				Required: []string{"id", "name", "type_id", "valid_id"},
			},
		},
	},
	{
		Name:        "TicketStateStatistics",
		Description: "Get ticket state statistics",
		Method:      "GET",
		Path:        "/api/v1/ticket-states/statistics",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"statistics": {
						Type: "array",
						Items: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"state_id":     {Type: "number"},
								"state_name":   {Type: "string"},
								"type_id":      {Type: "number"},
								"ticket_count": {Type: "number"},
							},
							Required: []string{"state_id", "state_name", "type_id", "ticket_count"},
						},
					},
					"total_tickets": {Type: "number"},
				},
				Required: []string{"statistics", "total_tickets"},
			},
		},
	},
}

// SLAContracts defines the API contracts for SLA endpoints
var SLAContracts = []Contract{
	{
		Name:        "ListSLAs",
		Description: "List all SLAs",
		Method:      "GET",
		Path:        "/api/v1/slas",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"slas": {
						Type: "array",
						Items: &Schema{
							Type: "object",
							Properties: map[string]Schema{
								"id":                    {Type: "number"},
								"name":                  {Type: "string"},
								"calendar_id":           {Type: "number"},
								"first_response_time":   {Type: "number"},
								"first_response_notify": {Type: "number"},
								"update_time":           {Type: "number"},
								"update_notify":         {Type: "number"},
								"solution_time":         {Type: "number"},
								"solution_notify":       {Type: "number"},
								"valid_id":              {Type: "number"},
							},
							Required: []string{"id", "name", "first_response_time", "solution_time", "valid_id"},
						},
					},
					"total": {Type: "number"},
				},
				Required: []string{"slas", "total"},
			},
		},
	},
	{
		Name:        "GetSLA",
		Description: "Get single SLA by ID",
		Method:      "GET",
		Path:        "/api/v1/slas/1",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"id":                    {Type: "number"},
					"name":                  {Type: "string"},
					"calendar_id":           {Type: "number"},
					"first_response_time":   {Type: "number"},
					"first_response_notify": {Type: "number"},
					"update_time":           {Type: "number"},
					"update_notify":         {Type: "number"},
					"solution_time":         {Type: "number"},
					"solution_notify":       {Type: "number"},
					"valid_id":              {Type: "number"},
				},
				Required: []string{"id", "name", "first_response_time", "solution_time", "valid_id"},
			},
		},
	},
	{
		Name:        "CreateSLA",
		Description: "Create new SLA",
		Method:      "POST",
		Path:        "/api/v1/slas",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
			"Content-Type":  "application/json",
		},
		Body: map[string]interface{}{
			"name":                  "Test SLA",
			"calendar_id":           1,
			"first_response_time":   60,
			"first_response_notify": 50,
			"update_time":           120,
			"update_notify":         100,
			"solution_time":         480,
			"solution_notify":       400,
		},
		Expected: Response{
			Status: http.StatusCreated,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"id":                    {Type: "number"},
					"name":                  {Type: "string"},
					"first_response_time":   {Type: "number"},
					"solution_time":         {Type: "number"},
					"valid_id":              {Type: "number"},
				},
				Required: []string{"id", "name", "first_response_time", "solution_time", "valid_id"},
			},
		},
	},
	{
		Name:        "UpdateSLA",
		Description: "Update existing SLA",
		Method:      "PUT",
		Path:        "/api/v1/slas/1",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
			"Content-Type":  "application/json",
		},
		Body: map[string]interface{}{
			"name":               "Updated SLA",
			"first_response_time": 45,
			"solution_time":      360,
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"id":                  {Type: "number"},
					"name":                {Type: "string"},
					"first_response_time": {Type: "number"},
					"solution_time":       {Type: "number"},
				},
				Required: []string{"id", "name"},
			},
		},
	},
	{
		Name:        "DeleteSLA",
		Description: "Soft delete SLA",
		Method:      "DELETE",
		Path:        "/api/v1/slas/10",
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
	{
		Name:        "SLAMetrics",
		Description: "Get SLA performance metrics",
		Method:      "GET",
		Path:        "/api/v1/slas/1/metrics",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
		},
		Expected: Response{
			Status: http.StatusOK,
			Schema: Schema{
				Type: "object",
				Properties: map[string]Schema{
					"sla_id":   {Type: "number"},
					"sla_name": {Type: "string"},
					"metrics": {
						Type: "object",
						Properties: map[string]Schema{
							"total_tickets":      {Type: "number"},
							"met_first_response": {Type: "number"},
							"met_solution":       {Type: "number"},
							"breached_tickets":   {Type: "number"},
							"compliance_percent": {Type: "number"},
						},
						Required: []string{"total_tickets", "met_first_response", "met_solution"},
					},
				},
				Required: []string{"sla_id", "metrics"},
			},
		},
	},
}

// RegisterStateSLAContracts registers ticket state and SLA contracts for testing
func RegisterStateSLAContracts() {
	for _, contract := range TicketStateContracts {
		RegisterContract(contract)
	}
	for _, contract := range SLAContracts {
		RegisterContract(contract)
	}
}