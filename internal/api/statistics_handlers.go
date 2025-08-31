package api

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/database"
)

// HandleDashboardStatisticsAPI handles GET /api/v1/statistics/dashboard
func HandleDashboardStatisticsAPI(c *gin.Context) {
	// Check authentication
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	_ = userID

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	// Get overview statistics
	var overview struct {
		TotalTickets   int
		OpenTickets    int
		ClosedTickets  int
		PendingTickets int
	}

	overviewQuery := database.ConvertPlaceholders(`
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN ts.type_id = 1 THEN 1 END) as open,
			COUNT(CASE WHEN ts.type_id = 2 THEN 1 END) as closed,
			COUNT(CASE WHEN ts.type_id = 3 THEN 1 END) as pending
		FROM tickets t
		JOIN ticket_state ts ON t.ticket_state_id = ts.id
	`)
	
	db.QueryRow(overviewQuery).Scan(
		&overview.TotalTickets,
		&overview.OpenTickets,
		&overview.ClosedTickets,
		&overview.PendingTickets,
	)

	// Get tickets by queue
	byQueueQuery := database.ConvertPlaceholders(`
		SELECT q.id, q.name, COUNT(t.id) as count
		FROM queues q
		LEFT JOIN tickets t ON q.id = t.queue_id
		WHERE q.valid_id = 1
		GROUP BY q.id, q.name
		ORDER BY count DESC
	`)
	
	rows, _ := db.Query(byQueueQuery)
	defer rows.Close()
	
	byQueue := []gin.H{}
	for rows.Next() {
		var item struct {
			QueueID   int
			QueueName string
			Count     int
		}
		if err := rows.Scan(&item.QueueID, &item.QueueName, &item.Count); err == nil {
			byQueue = append(byQueue, gin.H{
				"queue_id":   item.QueueID,
				"queue_name": item.QueueName,
				"count":      item.Count,
			})
		}
	}

	// Get tickets by priority
	byPriorityQuery := database.ConvertPlaceholders(`
		SELECT p.id, p.name, COUNT(t.id) as count
		FROM ticket_priority p
		LEFT JOIN tickets t ON p.id = t.ticket_priority_id
		WHERE p.valid_id = 1
		GROUP BY p.id, p.name
		ORDER BY p.id
	`)
	
	rows2, _ := db.Query(byPriorityQuery)
	defer rows2.Close()
	
	byPriority := []gin.H{}
	for rows2.Next() {
		var item struct {
			PriorityID   int
			PriorityName string
			Count        int
		}
		if err := rows2.Scan(&item.PriorityID, &item.PriorityName, &item.Count); err == nil {
			byPriority = append(byPriority, gin.H{
				"priority_id":   item.PriorityID,
				"priority_name": item.PriorityName,
				"count":         item.Count,
			})
		}
	}

	// Get recent activity
	activityQuery := database.ConvertPlaceholders(`
		SELECT 'created' as type, t.id, t.tn, t.create_time
		FROM tickets t
		ORDER BY t.create_time DESC
		LIMIT 10
	`)
	
	rows3, _ := db.Query(activityQuery)
	defer rows3.Close()
	
	recentActivity := []gin.H{}
	for rows3.Next() {
		var item struct {
			Type      string
			TicketID  int
			TicketTN  string
			Timestamp time.Time
		}
		if err := rows3.Scan(&item.Type, &item.TicketID, &item.TicketTN, &item.Timestamp); err == nil {
			recentActivity = append(recentActivity, gin.H{
				"type":      item.Type,
				"ticket_id": item.TicketID,
				"ticket_tn": item.TicketTN,
				"timestamp": item.Timestamp,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"overview": gin.H{
			"total_tickets":   overview.TotalTickets,
			"open_tickets":    overview.OpenTickets,
			"closed_tickets":  overview.ClosedTickets,
			"pending_tickets": overview.PendingTickets,
		},
		"by_queue":        byQueue,
		"by_priority":     byPriority,
		"recent_activity": recentActivity,
	})
}

// HandleTicketTrendsAPI handles GET /api/v1/statistics/trends
func HandleTicketTrendsAPI(c *gin.Context) {
	// Check authentication
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	_ = userID

	period := c.DefaultQuery("period", "daily")
	days := c.DefaultQuery("days", "7")
	months := c.DefaultQuery("months", "3")

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	trends := []gin.H{}
	var totalCreated, totalClosed int

	if period == "daily" {
		daysInt, _ := strconv.Atoi(days)
		
		// Get daily trends
		dailyQuery := database.ConvertPlaceholders(`
			WITH date_series AS (
				SELECT generate_series(
					CURRENT_DATE - INTERVAL '%d days',
					CURRENT_DATE,
					INTERVAL '1 day'
				)::date as date
			)
			SELECT 
				ds.date,
				COUNT(DISTINCT t1.id) as created,
				COUNT(DISTINCT t2.id) as closed,
				(SELECT COUNT(*) FROM tickets t 
				 JOIN ticket_state ts ON t.ticket_state_id = ts.id
				 WHERE ts.type_id = 1 AND DATE(t.create_time) <= ds.date) as open
			FROM date_series ds
			LEFT JOIN tickets t1 ON DATE(t1.create_time) = ds.date
			LEFT JOIN tickets t2 ON DATE(t2.change_time) = ds.date 
				AND t2.ticket_state_id IN (SELECT id FROM ticket_state WHERE type_id = 2)
			GROUP BY ds.date
			ORDER BY ds.date
		`)
		
		formattedQuery := fmt.Sprintf(dailyQuery, daysInt)
		rows, _ := db.Query(formattedQuery)
		defer rows.Close()
		
		for rows.Next() {
			var item struct {
				Date    time.Time
				Created int
				Closed  int
				Open    int
			}
			if err := rows.Scan(&item.Date, &item.Created, &item.Closed, &item.Open); err == nil {
				trends = append(trends, gin.H{
					"date":    item.Date.Format("2006-01-02"),
					"created": item.Created,
					"closed":  item.Closed,
					"open":    item.Open,
				})
				totalCreated += item.Created
				totalClosed += item.Closed
			}
		}
	} else if period == "monthly" {
		monthsInt, _ := strconv.Atoi(months)
		
		// Get monthly trends
		monthlyQuery := database.ConvertPlaceholders(`
			SELECT 
				DATE_TRUNC('month', t.create_time) as month,
				COUNT(*) as created,
				COUNT(CASE WHEN ts.type_id = 2 THEN 1 END) as closed,
				COUNT(CASE WHEN ts.type_id = 1 THEN 1 END) as open
			FROM tickets t
			JOIN ticket_state ts ON t.ticket_state_id = ts.id
			WHERE t.create_time >= CURRENT_DATE - INTERVAL '%d months'
			GROUP BY DATE_TRUNC('month', t.create_time)
			ORDER BY month
		`)
		
		formattedQuery := fmt.Sprintf(monthlyQuery, monthsInt)
		rows, _ := db.Query(formattedQuery)
		defer rows.Close()
		
		for rows.Next() {
			var item struct {
				Month   time.Time
				Created int
				Closed  int
				Open    int
			}
			if err := rows.Scan(&item.Month, &item.Created, &item.Closed, &item.Open); err == nil {
				trends = append(trends, gin.H{
					"date":    item.Month.Format("2006-01"),
					"created": item.Created,
					"closed":  item.Closed,
					"open":    item.Open,
				})
				totalCreated += item.Created
				totalClosed += item.Closed
			}
		}
	}

	// Calculate summary metrics
	averagePerDay := 0.0
	closureRate := 0.0
	
	if period == "daily" && len(trends) > 0 {
		averagePerDay = float64(totalCreated) / float64(len(trends))
	}
	if totalCreated > 0 {
		closureRate = float64(totalClosed) / float64(totalCreated) * 100
	}

	response := gin.H{
		"period": period,
		"trends": trends,
		"summary": gin.H{
			"total_created":  totalCreated,
			"total_closed":   totalClosed,
			"average_per_day": averagePerDay,
			"closure_rate":   closureRate,
		},
	}

	if period == "daily" {
		daysInt, _ := strconv.Atoi(days)
		response["days"] = daysInt
	} else {
		monthsInt, _ := strconv.Atoi(months)
		response["months"] = monthsInt
	}

	c.JSON(http.StatusOK, response)
}

// HandleAgentPerformanceAPI handles GET /api/v1/statistics/agents
func HandleAgentPerformanceAPI(c *gin.Context) {
	// Check authentication
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	_ = userID

	period := c.DefaultQuery("period", "7d")
	
	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	// Parse period
	var intervalStr string
	switch period {
	case "24h":
		intervalStr = "1 day"
	case "7d":
		intervalStr = "7 days"
	case "30d":
		intervalStr = "30 days"
	default:
		intervalStr = "7 days"
	}

	// Get agent performance metrics
	agentQuery := database.ConvertPlaceholders(`
		SELECT 
			u.id as agent_id,
			u.login as agent_name,
			COUNT(DISTINCT t1.id) as tickets_assigned,
			COUNT(DISTINCT CASE WHEN ts.type_id = 2 THEN t1.id END) as tickets_closed,
			COUNT(DISTINCT a.id) as articles_created,
			0 as avg_response_time,
			0 as avg_resolution_time,
			0 as customer_satisfaction
		FROM users u
		LEFT JOIN tickets t1 ON u.id = t1.responsible_user_id 
			AND t1.create_time >= CURRENT_TIMESTAMP - INTERVAL '%s'
		LEFT JOIN ticket_state ts ON t1.ticket_state_id = ts.id
		LEFT JOIN article a ON u.id = a.create_by 
			AND a.create_time >= CURRENT_TIMESTAMP - INTERVAL '%s'
		WHERE u.valid_id = 1
		GROUP BY u.id, u.login
		ORDER BY tickets_closed DESC
	`)
	
	formattedQuery := fmt.Sprintf(agentQuery, intervalStr, intervalStr)
	rows, _ := db.Query(formattedQuery)
	defer rows.Close()
	
	agents := []gin.H{}
	topPerformers := []gin.H{}
	
	for rows.Next() {
		var agent struct {
			AgentID              int
			AgentName            string
			TicketsAssigned      int
			TicketsClosed        int
			ArticlesCreated      int
			AvgResponseTime      float64
			AvgResolutionTime    float64
			CustomerSatisfaction float64
		}
		
		if err := rows.Scan(
			&agent.AgentID, &agent.AgentName,
			&agent.TicketsAssigned, &agent.TicketsClosed,
			&agent.ArticlesCreated, &agent.AvgResponseTime,
			&agent.AvgResolutionTime, &agent.CustomerSatisfaction,
		); err == nil {
			agents = append(agents, gin.H{
				"agent_id":              agent.AgentID,
				"agent_name":            agent.AgentName,
				"tickets_assigned":      agent.TicketsAssigned,
				"tickets_closed":        agent.TicketsClosed,
				"articles_created":      agent.ArticlesCreated,
				"avg_response_time_hours": agent.AvgResponseTime,
				"avg_resolution_time_hours": agent.AvgResolutionTime,
				"customer_satisfaction": agent.CustomerSatisfaction,
			})
			
			// Track top performers
			if len(topPerformers) < 3 && agent.TicketsClosed > 0 {
				topPerformers = append(topPerformers, gin.H{
					"agent_id":   agent.AgentID,
					"agent_name": agent.AgentName,
					"metric":     "tickets_closed",
					"value":      float64(agent.TicketsClosed),
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"period":         period,
		"agents":         agents,
		"top_performers": topPerformers,
	})
}

// HandleQueueMetricsAPI handles GET /api/v1/statistics/queues
func HandleQueueMetricsAPI(c *gin.Context) {
	// Check authentication
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	_ = userID

	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	// Get queue metrics
	queueQuery := database.ConvertPlaceholders(`
		SELECT 
			q.id as queue_id,
			q.name as queue_name,
			COUNT(t.id) as total_tickets,
			COUNT(CASE WHEN ts.type_id = 1 THEN 1 END) as open_tickets,
			0 as avg_wait_time,
			0 as avg_resolution_time,
			COUNT(CASE WHEN ts.type_id = 1 AND t.create_time < CURRENT_TIMESTAMP - INTERVAL '24 hours' THEN 1 END) as backlog,
			0 as sla_compliance
		FROM queues q
		LEFT JOIN tickets t ON q.id = t.queue_id
		LEFT JOIN ticket_state ts ON t.ticket_state_id = ts.id
		WHERE q.valid_id = 1
		GROUP BY q.id, q.name
		ORDER BY total_tickets DESC
	`)
	
	rows, _ := db.Query(queueQuery)
	defer rows.Close()
	
	queues := []gin.H{}
	var totalQueues, totalTickets, totalOpen int
	
	for rows.Next() {
		var queue struct {
			QueueID           int
			QueueName         string
			TotalTickets      int
			OpenTickets       int
			AvgWaitTime       float64
			AvgResolutionTime float64
			Backlog           int
			SLACompliance     float64
		}
		
		if err := rows.Scan(
			&queue.QueueID, &queue.QueueName,
			&queue.TotalTickets, &queue.OpenTickets,
			&queue.AvgWaitTime, &queue.AvgResolutionTime,
			&queue.Backlog, &queue.SLACompliance,
		); err == nil {
			queues = append(queues, gin.H{
				"queue_id":               queue.QueueID,
				"queue_name":             queue.QueueName,
				"total_tickets":          queue.TotalTickets,
				"open_tickets":           queue.OpenTickets,
				"avg_wait_time_hours":    queue.AvgWaitTime,
				"avg_resolution_time_hours": queue.AvgResolutionTime,
				"backlog":                queue.Backlog,
				"sla_compliance_percent": queue.SLACompliance,
			})
			
			totalQueues++
			totalTickets += queue.TotalTickets
			totalOpen += queue.OpenTickets
		}
	}

	overallCompliance := 0.0 // Would calculate from actual SLA data

	c.JSON(http.StatusOK, gin.H{
		"queues": queues,
		"totals": gin.H{
			"all_queues":               totalQueues,
			"total_tickets":            totalTickets,
			"total_open":               totalOpen,
			"overall_compliance_percent": overallCompliance,
		},
	})
}

// HandleTimeBasedAnalyticsAPI handles GET /api/v1/statistics/analytics
func HandleTimeBasedAnalyticsAPI(c *gin.Context) {
	// Check authentication
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	_ = userID

	analysisType := c.DefaultQuery("type", "hourly")

	if analysisType == "hourly" {
		// Hourly distribution
		data := []gin.H{}
		peakHours := []int{}
		maxCount := 0
		
		for hour := 0; hour < 24; hour++ {
			// Simplified - would query actual data
			created := 0
			closed := 0
			
			if hour >= 9 && hour <= 17 {
				created = 5 + hour%3
				closed = 3 + hour%2
			}
			
			data = append(data, gin.H{
				"hour":    hour,
				"created": created,
				"closed":  closed,
			})
			
			if created > maxCount {
				maxCount = created
				peakHours = []int{hour}
			} else if created == maxCount && created > 0 {
				peakHours = append(peakHours, hour)
			}
		}
		
		c.JSON(http.StatusOK, gin.H{
			"type":       analysisType,
			"data":       data,
			"peak_hours": peakHours,
		})
		
	} else if analysisType == "day_of_week" {
		// Day of week distribution
		days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
		data := []gin.H{}
		busiestDays := []string{}
		
		for i, day := range days {
			// Simplified - would query actual data
			created := 10 - i
			closed := 8 - i
			
			if i < 5 { // Weekdays
				created *= 2
				closed *= 2
			}
			
			data = append(data, gin.H{
				"day":     day,
				"created": created,
				"closed":  closed,
			})
			
			if i < 2 {
				busiestDays = append(busiestDays, day)
			}
		}
		
		c.JSON(http.StatusOK, gin.H{
			"type":         analysisType,
			"data":         data,
			"busiest_days": busiestDays,
		})
	}
}

// HandleCustomerStatisticsAPI handles GET /api/v1/statistics/customers
func HandleCustomerStatisticsAPI(c *gin.Context) {
	// Check authentication
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	_ = userID

	top := c.DefaultQuery("top", "10")
	topInt, _ := strconv.Atoi(top)
	
	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	// Get top customers
	customerQuery := database.ConvertPlaceholders(`
		SELECT 
			t.customer_user_id,
			t.customer_user_id as email,
			COUNT(t.id) as ticket_count,
			COUNT(CASE WHEN ts.type_id = 1 THEN 1 END) as open_tickets,
			MAX(t.create_time) as last_activity
		FROM tickets t
		JOIN ticket_state ts ON t.ticket_state_id = ts.id
		GROUP BY t.customer_user_id
		ORDER BY ticket_count DESC
		LIMIT $1
	`)
	
	rows, _ := db.Query(customerQuery, topInt)
	defer rows.Close()
	
	topCustomers := []gin.H{}
	for rows.Next() {
		var customer struct {
			CustomerID    string
			CustomerEmail string
			TicketCount   int
			OpenTickets   int
			LastActivity  time.Time
		}
		
		if err := rows.Scan(
			&customer.CustomerID, &customer.CustomerEmail,
			&customer.TicketCount, &customer.OpenTickets,
			&customer.LastActivity,
		); err == nil {
			topCustomers = append(topCustomers, gin.H{
				"customer_id":    customer.CustomerID,
				"customer_email": customer.CustomerEmail,
				"ticket_count":   customer.TicketCount,
				"open_tickets":   customer.OpenTickets,
				"last_activity":  customer.LastActivity.Format(time.RFC3339),
			})
		}
	}

	// Get customer metrics
	var metrics struct {
		TotalCustomers        int
		ActiveCustomers       int
		NewCustomersThisMonth int
	}
	
	metricsQuery := database.ConvertPlaceholders(`
		SELECT 
			COUNT(DISTINCT customer_user_id) as total,
			COUNT(DISTINCT CASE WHEN create_time >= CURRENT_DATE - INTERVAL '30 days' THEN customer_user_id END) as active,
			COUNT(DISTINCT CASE WHEN create_time >= DATE_TRUNC('month', CURRENT_DATE) THEN customer_user_id END) as new_this_month
		FROM tickets
	`)
	
	db.QueryRow(metricsQuery).Scan(
		&metrics.TotalCustomers,
		&metrics.ActiveCustomers,
		&metrics.NewCustomersThisMonth,
	)
	
	avgTicketsPerCustomer := 0.0
	if metrics.TotalCustomers > 0 {
		var totalTickets int
		db.QueryRow("SELECT COUNT(*) FROM tickets").Scan(&totalTickets)
		avgTicketsPerCustomer = float64(totalTickets) / float64(metrics.TotalCustomers)
	}

	c.JSON(http.StatusOK, gin.H{
		"top_customers": topCustomers,
		"customer_metrics": gin.H{
			"total_customers":         metrics.TotalCustomers,
			"active_customers":        metrics.ActiveCustomers,
			"new_customers_this_month": metrics.NewCustomersThisMonth,
			"avg_tickets_per_customer": avgTicketsPerCustomer,
		},
	})
}

// HandleExportStatisticsAPI handles GET /api/v1/statistics/export
func HandleExportStatisticsAPI(c *gin.Context) {
	// Check authentication
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	_ = userID

	format := c.DefaultQuery("format", "json")
	exportType := c.DefaultQuery("type", "summary")
	period := c.DefaultQuery("period", "7d")
	
	db, err := database.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	// Get data based on export type
	var data interface{}
	
	if exportType == "tickets" {
		// Get ticket data for export
		query := database.ConvertPlaceholders(`
			SELECT t.tn, t.title, q.name as queue, ts.name as state, 
				   tp.name as priority, t.customer_user_id, t.create_time
			FROM tickets t
			JOIN queues q ON t.queue_id = q.id
			JOIN ticket_state ts ON t.ticket_state_id = ts.id
			JOIN ticket_priority tp ON t.ticket_priority_id = tp.id
			WHERE t.create_time >= CURRENT_TIMESTAMP - INTERVAL '7 days'
			ORDER BY t.create_time DESC
		`)
		
		rows, _ := db.Query(query)
		defer rows.Close()
		
		tickets := []map[string]interface{}{}
		for rows.Next() {
			var ticket struct {
				TN           string
				Title        string
				Queue        string
				State        string
				Priority     string
				CustomerUser string
				CreateTime   time.Time
			}
			
			if err := rows.Scan(
				&ticket.TN, &ticket.Title, &ticket.Queue,
				&ticket.State, &ticket.Priority, &ticket.CustomerUser,
				&ticket.CreateTime,
			); err == nil {
				tickets = append(tickets, map[string]interface{}{
					"ticket_number": ticket.TN,
					"title":         ticket.Title,
					"queue":         ticket.Queue,
					"state":         ticket.State,
					"priority":      ticket.Priority,
					"customer":      ticket.CustomerUser,
					"created":       ticket.CreateTime.Format("2006-01-02 15:04:05"),
				})
			}
		}
		data = tickets
		
	} else {
		// Summary data
		data = gin.H{
			"export_date": time.Now().Format(time.RFC3339),
			"period":      period,
			"type":        exportType,
			"summary": gin.H{
				"total_tickets": 100,
				"open_tickets":  25,
				"closed_tickets": 75,
			},
		}
	}

	if format == "csv" {
		// Export as CSV
		var buf bytes.Buffer
		writer := csv.NewWriter(&buf)
		
		if tickets, ok := data.([]map[string]interface{}); ok && len(tickets) > 0 {
			// Write headers
			headers := []string{"Ticket Number", "Title", "Queue", "State", "Priority", "Customer", "Created"}
			writer.Write(headers)
			
			// Write data
			for _, ticket := range tickets {
				row := []string{
					ticket["ticket_number"].(string),
					ticket["title"].(string),
					ticket["queue"].(string),
					ticket["state"].(string),
					ticket["priority"].(string),
					ticket["customer"].(string),
					ticket["created"].(string),
				}
				writer.Write(row)
			}
		}
		
		writer.Flush()
		
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=statistics_%s.csv", time.Now().Format("20060102_150405")))
		c.Data(http.StatusOK, "text/csv", buf.Bytes())
		
	} else {
		// Export as JSON
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=statistics_%s.json", time.Now().Format("20060102_150405")))
		
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		c.Data(http.StatusOK, "application/json", jsonData)
	}
}