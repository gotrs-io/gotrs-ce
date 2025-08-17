package service

import (
	"fmt"
	"sort"
	"time"

	"github.com/gotrs-io/gotrs-ce/internal/models"
	"github.com/gotrs-io/gotrs-ce/internal/repository"
)

// WorkflowAnalytics provides analytics and insights for workflows
type WorkflowAnalytics struct {
	workflowRepo repository.WorkflowRepository
	ticketRepo   repository.TicketRepository
}

// NewWorkflowAnalytics creates a new workflow analytics service
func NewWorkflowAnalytics(workflowRepo repository.WorkflowRepository, ticketRepo repository.TicketRepository) *WorkflowAnalytics {
	return &WorkflowAnalytics{
		workflowRepo: workflowRepo,
		ticketRepo:   ticketRepo,
	}
}

// WorkflowMetrics contains overall workflow metrics
type WorkflowMetrics struct {
	TotalWorkflows       int                    `json:"total_workflows"`
	ActiveWorkflows      int                    `json:"active_workflows"`
	TotalExecutions      int                    `json:"total_executions"`
	SuccessfulExecutions int                    `json:"successful_executions"`
	FailedExecutions     int                    `json:"failed_executions"`
	SuccessRate          float64                `json:"success_rate"`
	AverageExecutionTime int64                  `json:"average_execution_time_ms"`
	TotalActionsSaved    int                    `json:"total_actions_saved"`
	TimeSaved            int64                  `json:"time_saved_hours"`
	TopWorkflows         []WorkflowSummary      `json:"top_workflows"`
	ExecutionTrends      []ExecutionTrend       `json:"execution_trends"`
	ErrorAnalysis        ErrorAnalysis          `json:"error_analysis"`
	ROIMetrics           ROIMetrics             `json:"roi_metrics"`
}

// WorkflowSummary provides summary statistics for a single workflow
type WorkflowSummary struct {
	ID                   int                    `json:"id"`
	Name                 string                 `json:"name"`
	Status               string                 `json:"status"`
	ExecutionCount       int                    `json:"execution_count"`
	SuccessRate          float64                `json:"success_rate"`
	AverageExecutionTime int64                  `json:"average_execution_time_ms"`
	LastExecuted         *time.Time             `json:"last_executed"`
	TriggerDistribution  map[string]int         `json:"trigger_distribution"`
	ActionDistribution   map[string]int         `json:"action_distribution"`
	ImpactScore          float64                `json:"impact_score"`
}

// ExecutionTrend shows execution trends over time
type ExecutionTrend struct {
	Date             time.Time              `json:"date"`
	ExecutionCount   int                    `json:"execution_count"`
	SuccessCount     int                    `json:"success_count"`
	FailureCount     int                    `json:"failure_count"`
	AverageTime      int64                  `json:"average_time_ms"`
}

// ErrorAnalysis provides insights into workflow errors
type ErrorAnalysis struct {
	TotalErrors          int                    `json:"total_errors"`
	ErrorRate            float64                `json:"error_rate"`
	CommonErrors         []ErrorPattern         `json:"common_errors"`
	ErrorsByComponent    map[string]int         `json:"errors_by_component"`
	ErrorsByTime         []ErrorTimeline        `json:"errors_by_time"`
	RecoveryRate         float64                `json:"recovery_rate"`
}

// ErrorPattern represents a common error pattern
type ErrorPattern struct {
	Pattern      string                 `json:"pattern"`
	Count        int                    `json:"count"`
	Percentage   float64                `json:"percentage"`
	AffectedFlows []string              `json:"affected_flows"`
	LastOccurred time.Time              `json:"last_occurred"`
}

// ErrorTimeline shows errors over time
type ErrorTimeline struct {
	Time       time.Time              `json:"time"`
	ErrorCount int                    `json:"error_count"`
	ErrorTypes map[string]int         `json:"error_types"`
}

// ROIMetrics calculates return on investment metrics
type ROIMetrics struct {
	AutomatedActions     int                    `json:"automated_actions"`
	ManualActionsAvoided int                    `json:"manual_actions_avoided"`
	TimeSavedHours       float64                `json:"time_saved_hours"`
	CostSavings          float64                `json:"cost_savings_usd"`
	EfficiencyGain       float64                `json:"efficiency_gain_percent"`
	TicketResolutionTime float64                `json:"avg_resolution_time_reduction_percent"`
}

// PerformanceAnalysis analyzes workflow performance
type PerformanceAnalysis struct {
	WorkflowID           int                    `json:"workflow_id"`
	WorkflowName         string                 `json:"workflow_name"`
	PerformanceTrend     string                 `json:"performance_trend"` // improving, stable, degrading
	Bottlenecks          []Bottleneck           `json:"bottlenecks"`
	OptimizationSuggestions []string            `json:"optimization_suggestions"`
	ResourceUsage        ResourceUsage          `json:"resource_usage"`
	SLACompliance        float64                `json:"sla_compliance_percent"`
}

// Bottleneck identifies performance bottlenecks
type Bottleneck struct {
	ComponentID   string                 `json:"component_id"`
	ComponentType string                 `json:"component_type"`
	AverageTime   int64                  `json:"average_time_ms"`
	MaxTime       int64                  `json:"max_time_ms"`
	Frequency     int                    `json:"frequency"`
	Impact        string                 `json:"impact"` // high, medium, low
}

// ResourceUsage tracks resource consumption
type ResourceUsage struct {
	CPUUsage         float64                `json:"cpu_usage_percent"`
	MemoryUsage      int64                  `json:"memory_usage_mb"`
	DatabaseQueries  int                    `json:"database_queries"`
	APICallCount     int                    `json:"api_call_count"`
	NetworkBandwidth int64                  `json:"network_bandwidth_kb"`
}

// TriggerAnalysis analyzes trigger patterns
type TriggerAnalysis struct {
	TotalTriggers        int                    `json:"total_triggers"`
	TriggerDistribution  map[string]int         `json:"trigger_distribution"`
	PeakTriggerTimes     []PeakTime             `json:"peak_trigger_times"`
	TriggerPatterns      []TriggerPattern       `json:"trigger_patterns"`
	PredictedTriggers    []PredictedTrigger     `json:"predicted_triggers"`
}

// PeakTime represents a peak activity time
type PeakTime struct {
	Hour         int                    `json:"hour"`
	DayOfWeek    string                 `json:"day_of_week"`
	TriggerCount int                    `json:"trigger_count"`
}

// TriggerPattern represents a recurring trigger pattern
type TriggerPattern struct {
	Pattern      string                 `json:"pattern"`
	Frequency    int                    `json:"frequency"`
	Periodicity  string                 `json:"periodicity"` // daily, weekly, monthly
	NextExpected time.Time              `json:"next_expected"`
}

// PredictedTrigger represents a predicted future trigger
type PredictedTrigger struct {
	TriggerType  string                 `json:"trigger_type"`
	PredictedAt  time.Time              `json:"predicted_at"`
	Probability  float64                `json:"probability"`
	BasedOn      string                 `json:"based_on"`
}

// ActionAnalysis analyzes action patterns and effectiveness
type ActionAnalysis struct {
	TotalActions         int                    `json:"total_actions"`
	ActionDistribution   map[string]int         `json:"action_distribution"`
	ActionSuccess        map[string]float64     `json:"action_success_rates"`
	MostEffective        []EffectiveAction      `json:"most_effective_actions"`
	LeastEffective       []EffectiveAction      `json:"least_effective_actions"`
	ActionChains         []ActionChain          `json:"common_action_chains"`
}

// EffectiveAction represents action effectiveness metrics
type EffectiveAction struct {
	ActionType    string                 `json:"action_type"`
	SuccessRate   float64                `json:"success_rate"`
	AverageTime   int64                  `json:"average_time_ms"`
	ImpactScore   float64                `json:"impact_score"`
	UsageCount    int                    `json:"usage_count"`
}

// ActionChain represents a common sequence of actions
type ActionChain struct {
	Actions      []string               `json:"actions"`
	Frequency    int                    `json:"frequency"`
	SuccessRate  float64                `json:"success_rate"`
	AverageTime  int64                  `json:"average_time_ms"`
}

// ComparisonAnalysis compares workflow performance
type ComparisonAnalysis struct {
	WorkflowA            WorkflowSummary        `json:"workflow_a"`
	WorkflowB            WorkflowSummary        `json:"workflow_b"`
	PerformanceDiff      float64                `json:"performance_diff_percent"`
	SuccessRateDiff      float64                `json:"success_rate_diff_percent"`
	ExecutionTimeDiff    int64                  `json:"execution_time_diff_ms"`
	Recommendations      []string               `json:"recommendations"`
}

// GetOverallMetrics returns overall workflow metrics
func (a *WorkflowAnalytics) GetOverallMetrics(period string) (*WorkflowMetrics, error) {
	startDate := a.getStartDate(period)
	
	// Get all workflows
	workflows, err := a.workflowRepo.GetAll()
	if err != nil {
		return nil, err
	}
	
	metrics := &WorkflowMetrics{
		TotalWorkflows:  len(workflows),
		TopWorkflows:    []WorkflowSummary{},
		ExecutionTrends: []ExecutionTrend{},
	}
	
	// Count active workflows
	for _, w := range workflows {
		if w.Status == models.WorkflowStatusActive {
			metrics.ActiveWorkflows++
		}
	}
	
	// Get execution data
	executions, err := a.workflowRepo.GetExecutionsSince(startDate)
	if err != nil {
		return nil, err
	}
	
	metrics.TotalExecutions = len(executions)
	
	var totalTime int64
	for _, exec := range executions {
		if exec.Status == "success" {
			metrics.SuccessfulExecutions++
		} else if exec.Status == "failed" {
			metrics.FailedExecutions++
		}
		
		if exec.CompletedAt != nil {
			duration := exec.CompletedAt.Sub(exec.StartedAt).Milliseconds()
			totalTime += duration
		}
		
		metrics.TotalActionsSaved += exec.ActionsRun
	}
	
	if metrics.TotalExecutions > 0 {
		metrics.SuccessRate = float64(metrics.SuccessfulExecutions) / float64(metrics.TotalExecutions) * 100
		metrics.AverageExecutionTime = totalTime / int64(metrics.TotalExecutions)
	}
	
	// Calculate time saved (assuming 2 minutes per automated action)
	metrics.TimeSaved = int64(metrics.TotalActionsSaved * 2 / 60)
	
	// Get top workflows
	metrics.TopWorkflows = a.getTopWorkflows(workflows, executions, 5)
	
	// Get execution trends
	metrics.ExecutionTrends = a.getExecutionTrends(executions, 7)
	
	// Analyze errors
	metrics.ErrorAnalysis = a.analyzeErrors(executions)
	
	// Calculate ROI
	metrics.ROIMetrics = a.calculateROI(metrics)
	
	return metrics, nil
}

// GetWorkflowAnalytics returns detailed analytics for a specific workflow
func (a *WorkflowAnalytics) GetWorkflowAnalytics(workflowID int, period string) (*WorkflowSummary, error) {
	workflow, err := a.workflowRepo.GetByID(workflowID)
	if err != nil {
		return nil, err
	}
	
	startDate := a.getStartDate(period)
	executions, err := a.workflowRepo.GetWorkflowExecutions(workflowID, startDate)
	if err != nil {
		return nil, err
	}
	
	summary := &WorkflowSummary{
		ID:                  workflow.ID,
		Name:                workflow.Name,
		Status:              string(workflow.Status),
		ExecutionCount:      len(executions),
		TriggerDistribution: make(map[string]int),
		ActionDistribution:  make(map[string]int),
	}
	
	var successCount int
	var totalTime int64
	
	for _, exec := range executions {
		if exec.Status == "success" {
			successCount++
		}
		
		if exec.CompletedAt != nil {
			duration := exec.CompletedAt.Sub(exec.StartedAt).Milliseconds()
			totalTime += duration
		}
		
		// Count trigger types
		summary.TriggerDistribution[string(exec.TriggerType)]++
		
		// Count action types from execution log
		for _, entry := range exec.ExecutionLog {
			if entry.ActionType != "" {
				summary.ActionDistribution[string(entry.ActionType)]++
			}
		}
		
		if summary.LastExecuted == nil || exec.StartedAt.After(*summary.LastExecuted) {
			summary.LastExecuted = &exec.StartedAt
		}
	}
	
	if summary.ExecutionCount > 0 {
		summary.SuccessRate = float64(successCount) / float64(summary.ExecutionCount) * 100
		summary.AverageExecutionTime = totalTime / int64(summary.ExecutionCount)
		summary.ImpactScore = a.calculateImpactScore(summary)
	}
	
	return summary, nil
}

// GetPerformanceAnalysis analyzes workflow performance
func (a *WorkflowAnalytics) GetPerformanceAnalysis(workflowID int) (*PerformanceAnalysis, error) {
	workflow, err := a.workflowRepo.GetByID(workflowID)
	if err != nil {
		return nil, err
	}
	
	executions, err := a.workflowRepo.GetWorkflowExecutions(workflowID, time.Now().AddDate(0, -1, 0))
	if err != nil {
		return nil, err
	}
	
	analysis := &PerformanceAnalysis{
		WorkflowID:   workflow.ID,
		WorkflowName: workflow.Name,
		Bottlenecks:  []Bottleneck{},
		OptimizationSuggestions: []string{},
		ResourceUsage: ResourceUsage{},
	}
	
	// Analyze performance trend
	analysis.PerformanceTrend = a.analyzePerformanceTrend(executions)
	
	// Identify bottlenecks
	analysis.Bottlenecks = a.identifyBottlenecks(executions)
	
	// Generate optimization suggestions
	analysis.OptimizationSuggestions = a.generateOptimizationSuggestions(analysis.Bottlenecks)
	
	// Calculate resource usage
	analysis.ResourceUsage = a.calculateResourceUsage(executions)
	
	// Calculate SLA compliance
	analysis.SLACompliance = a.calculateSLACompliance(workflow, executions)
	
	return analysis, nil
}

// GetTriggerAnalysis analyzes trigger patterns
func (a *WorkflowAnalytics) GetTriggerAnalysis(period string) (*TriggerAnalysis, error) {
	startDate := a.getStartDate(period)
	executions, err := a.workflowRepo.GetExecutionsSince(startDate)
	if err != nil {
		return nil, err
	}
	
	analysis := &TriggerAnalysis{
		TriggerDistribution: make(map[string]int),
		PeakTriggerTimes:    []PeakTime{},
		TriggerPatterns:     []TriggerPattern{},
		PredictedTriggers:   []PredictedTrigger{},
	}
	
	// Count triggers
	hourlyCount := make(map[int]int)
	dayCount := make(map[string]int)
	
	for _, exec := range executions {
		analysis.TotalTriggers++
		analysis.TriggerDistribution[string(exec.TriggerType)]++
		
		// Track hourly distribution
		hour := exec.StartedAt.Hour()
		hourlyCount[hour]++
		
		// Track daily distribution
		day := exec.StartedAt.Weekday().String()
		dayCount[day]++
	}
	
	// Find peak times
	for hour, count := range hourlyCount {
		if count > analysis.TotalTriggers/24 { // Above average
			analysis.PeakTriggerTimes = append(analysis.PeakTriggerTimes, PeakTime{
				Hour:         hour,
				TriggerCount: count,
			})
		}
	}
	
	// Identify patterns
	analysis.TriggerPatterns = a.identifyTriggerPatterns(executions)
	
	// Generate predictions
	analysis.PredictedTriggers = a.predictFutureTriggers(analysis.TriggerPatterns)
	
	return analysis, nil
}

// GetActionAnalysis analyzes action patterns and effectiveness
func (a *WorkflowAnalytics) GetActionAnalysis(period string) (*ActionAnalysis, error) {
	startDate := a.getStartDate(period)
	executions, err := a.workflowRepo.GetExecutionsSince(startDate)
	if err != nil {
		return nil, err
	}
	
	analysis := &ActionAnalysis{
		ActionDistribution: make(map[string]int),
		ActionSuccess:      make(map[string]float64),
		MostEffective:      []EffectiveAction{},
		LeastEffective:     []EffectiveAction{},
		ActionChains:       []ActionChain{},
	}
	
	actionStats := make(map[string]*actionStat)
	
	// Analyze actions from execution logs
	for _, exec := range executions {
		for _, entry := range exec.ExecutionLog {
			if entry.ActionType != "" {
				analysis.TotalActions++
				analysis.ActionDistribution[entry.ActionType]++
				
				// Track success rates
				if _, exists := actionStats[entry.ActionType]; !exists {
					actionStats[entry.ActionType] = &actionStat{}
				}
				
				stats := actionStats[entry.ActionType]
				stats.total++
				if entry.Status == "completed" {
					stats.success++
				}
				stats.totalTime += entry.Duration
			}
		}
	}
	
	// Calculate success rates
	for actionType, stats := range actionStats {
		if stats.total > 0 {
			analysis.ActionSuccess[actionType] = float64(stats.success) / float64(stats.total) * 100
			
			effectiveAction := EffectiveAction{
				ActionType:  actionType,
				SuccessRate: analysis.ActionSuccess[actionType],
				AverageTime: stats.totalTime / int64(stats.total),
				UsageCount:  stats.total,
				ImpactScore: a.calculateActionImpact(actionType, stats),
			}
			
			if effectiveAction.SuccessRate >= 90 {
				analysis.MostEffective = append(analysis.MostEffective, effectiveAction)
			} else if effectiveAction.SuccessRate < 70 {
				analysis.LeastEffective = append(analysis.LeastEffective, effectiveAction)
			}
		}
	}
	
	// Sort most and least effective
	sort.Slice(analysis.MostEffective, func(i, j int) bool {
		return analysis.MostEffective[i].ImpactScore > analysis.MostEffective[j].ImpactScore
	})
	
	sort.Slice(analysis.LeastEffective, func(i, j int) bool {
		return analysis.LeastEffective[i].SuccessRate < analysis.LeastEffective[j].SuccessRate
	})
	
	// Identify common action chains
	analysis.ActionChains = a.identifyActionChains(executions)
	
	return analysis, nil
}

// CompareWorkflows compares two workflows
func (a *WorkflowAnalytics) CompareWorkflows(workflowAID, workflowBID int, period string) (*ComparisonAnalysis, error) {
	summaryA, err := a.GetWorkflowAnalytics(workflowAID, period)
	if err != nil {
		return nil, err
	}
	
	summaryB, err := a.GetWorkflowAnalytics(workflowBID, period)
	if err != nil {
		return nil, err
	}
	
	comparison := &ComparisonAnalysis{
		WorkflowA:       *summaryA,
		WorkflowB:       *summaryB,
		Recommendations: []string{},
	}
	
	// Calculate differences
	if summaryB.AverageExecutionTime > 0 {
		comparison.PerformanceDiff = float64(summaryA.AverageExecutionTime-summaryB.AverageExecutionTime) / float64(summaryB.AverageExecutionTime) * 100
	}
	
	comparison.SuccessRateDiff = summaryA.SuccessRate - summaryB.SuccessRate
	comparison.ExecutionTimeDiff = summaryA.AverageExecutionTime - summaryB.AverageExecutionTime
	
	// Generate recommendations
	if comparison.SuccessRateDiff > 10 {
		comparison.Recommendations = append(comparison.Recommendations, 
			fmt.Sprintf("Consider applying patterns from %s to improve %s success rate", summaryA.Name, summaryB.Name))
	}
	
	if comparison.ExecutionTimeDiff < -1000 {
		comparison.Recommendations = append(comparison.Recommendations,
			fmt.Sprintf("%s is significantly faster - review for optimization opportunities", summaryB.Name))
	}
	
	return comparison, nil
}

// Helper methods

type actionStat struct {
	total     int
	success   int
	totalTime int64
}

func (a *WorkflowAnalytics) getStartDate(period string) time.Time {
	now := time.Now()
	switch period {
	case "day":
		return now.AddDate(0, 0, -1)
	case "week":
		return now.AddDate(0, 0, -7)
	case "month":
		return now.AddDate(0, -1, 0)
	case "quarter":
		return now.AddDate(0, -3, 0)
	case "year":
		return now.AddDate(-1, 0, 0)
	default:
		return now.AddDate(0, -1, 0) // Default to last month
	}
}

func (a *WorkflowAnalytics) getTopWorkflows(workflows []*models.Workflow, executions []*models.WorkflowExecution, limit int) []WorkflowSummary {
	workflowStats := make(map[int]*WorkflowSummary)
	
	// Initialize summaries
	for _, w := range workflows {
		workflowStats[w.ID] = &WorkflowSummary{
			ID:     w.ID,
			Name:   w.Name,
			Status: string(w.Status),
		}
	}
	
	// Aggregate execution data
	for _, exec := range executions {
		if summary, exists := workflowStats[exec.WorkflowID]; exists {
			summary.ExecutionCount++
			if exec.Status == "success" {
				summary.SuccessRate++
			}
		}
	}
	
	// Calculate success rates
	for _, summary := range workflowStats {
		if summary.ExecutionCount > 0 {
			summary.SuccessRate = summary.SuccessRate / float64(summary.ExecutionCount) * 100
			summary.ImpactScore = a.calculateImpactScore(summary)
		}
	}
	
	// Convert to slice and sort
	summaries := make([]WorkflowSummary, 0, len(workflowStats))
	for _, summary := range workflowStats {
		if summary.ExecutionCount > 0 {
			summaries = append(summaries, *summary)
		}
	}
	
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].ImpactScore > summaries[j].ImpactScore
	})
	
	if len(summaries) > limit {
		return summaries[:limit]
	}
	
	return summaries
}

func (a *WorkflowAnalytics) getExecutionTrends(executions []*models.WorkflowExecution, days int) []ExecutionTrend {
	trends := make(map[string]*ExecutionTrend)
	
	for _, exec := range executions {
		dateKey := exec.StartedAt.Format("2006-01-02")
		
		if _, exists := trends[dateKey]; !exists {
			trends[dateKey] = &ExecutionTrend{
				Date: exec.StartedAt.Truncate(24 * time.Hour),
			}
		}
		
		trend := trends[dateKey]
		trend.ExecutionCount++
		
		if exec.Status == "success" {
			trend.SuccessCount++
		} else if exec.Status == "failed" {
			trend.FailureCount++
		}
		
		if exec.CompletedAt != nil {
			duration := exec.CompletedAt.Sub(exec.StartedAt).Milliseconds()
			trend.AverageTime = (trend.AverageTime*int64(trend.ExecutionCount-1) + duration) / int64(trend.ExecutionCount)
		}
	}
	
	// Convert to slice and sort by date
	trendSlice := make([]ExecutionTrend, 0, len(trends))
	for _, trend := range trends {
		trendSlice = append(trendSlice, *trend)
	}
	
	sort.Slice(trendSlice, func(i, j int) bool {
		return trendSlice[i].Date.Before(trendSlice[j].Date)
	})
	
	// Limit to requested days
	if len(trendSlice) > days {
		return trendSlice[len(trendSlice)-days:]
	}
	
	return trendSlice
}

func (a *WorkflowAnalytics) analyzeErrors(executions []*models.WorkflowExecution) ErrorAnalysis {
	analysis := ErrorAnalysis{
		ErrorsByComponent: make(map[string]int),
		CommonErrors:      []ErrorPattern{},
		ErrorsByTime:      []ErrorTimeline{},
	}
	
	errorPatterns := make(map[string]*ErrorPattern)
	
	for _, exec := range executions {
		if exec.Status == "failed" {
			analysis.TotalErrors++
			
			// Track error patterns
			if exec.ErrorMessage != "" {
				if _, exists := errorPatterns[exec.ErrorMessage]; !exists {
					errorPatterns[exec.ErrorMessage] = &ErrorPattern{
						Pattern:       exec.ErrorMessage,
						AffectedFlows: []string{},
					}
				}
				
				pattern := errorPatterns[exec.ErrorMessage]
				pattern.Count++
				pattern.LastOccurred = exec.StartedAt
				pattern.AffectedFlows = append(pattern.AffectedFlows, exec.WorkflowName)
			}
			
			// Track errors by component
			for _, entry := range exec.ExecutionLog {
				if entry.Status == "failed" {
					analysis.ErrorsByComponent[entry.ActionType]++
				}
			}
		}
	}
	
	// Calculate error rate
	if len(executions) > 0 {
		analysis.ErrorRate = float64(analysis.TotalErrors) / float64(len(executions)) * 100
	}
	
	// Convert patterns to slice
	for _, pattern := range errorPatterns {
		if analysis.TotalErrors > 0 {
			pattern.Percentage = float64(pattern.Count) / float64(analysis.TotalErrors) * 100
		}
		analysis.CommonErrors = append(analysis.CommonErrors, *pattern)
	}
	
	// Sort by frequency
	sort.Slice(analysis.CommonErrors, func(i, j int) bool {
		return analysis.CommonErrors[i].Count > analysis.CommonErrors[j].Count
	})
	
	return analysis
}

func (a *WorkflowAnalytics) calculateROI(metrics *WorkflowMetrics) ROIMetrics {
	roi := ROIMetrics{
		AutomatedActions:     metrics.TotalActionsSaved,
		ManualActionsAvoided: metrics.TotalActionsSaved,
		TimeSavedHours:       float64(metrics.TimeSaved),
	}
	
	// Assume $30/hour for manual work
	roi.CostSavings = roi.TimeSavedHours * 30
	
	// Calculate efficiency gain
	if metrics.TotalExecutions > 0 {
		roi.EfficiencyGain = float64(metrics.TotalActionsSaved) / float64(metrics.TotalExecutions) * 100
	}
	
	// Estimate resolution time reduction (based on automation success rate)
	roi.TicketResolutionTime = metrics.SuccessRate * 0.3 // 30% reduction for successful automations
	
	return roi
}

func (a *WorkflowAnalytics) calculateImpactScore(summary *WorkflowSummary) float64 {
	// Impact score based on execution count, success rate, and time saved
	executionWeight := float64(summary.ExecutionCount) * 0.3
	successWeight := summary.SuccessRate * 0.5
	timeWeight := 0.0
	
	if summary.AverageExecutionTime > 0 {
		// Faster workflows have higher impact
		timeWeight = (10000 / float64(summary.AverageExecutionTime)) * 0.2
	}
	
	return executionWeight + successWeight + timeWeight
}

func (a *WorkflowAnalytics) analyzePerformanceTrend(executions []*models.WorkflowExecution) string {
	if len(executions) < 10 {
		return "insufficient_data"
	}
	
	// Compare first half vs second half performance
	midpoint := len(executions) / 2
	
	var firstHalfTime, secondHalfTime int64
	var firstCount, secondCount int
	
	for i, exec := range executions {
		if exec.CompletedAt != nil {
			duration := exec.CompletedAt.Sub(exec.StartedAt).Milliseconds()
			
			if i < midpoint {
				firstHalfTime += duration
				firstCount++
			} else {
				secondHalfTime += duration
				secondCount++
			}
		}
	}
	
	if firstCount == 0 || secondCount == 0 {
		return "stable"
	}
	
	firstAvg := firstHalfTime / int64(firstCount)
	secondAvg := secondHalfTime / int64(secondCount)
	
	percentChange := float64(secondAvg-firstAvg) / float64(firstAvg) * 100
	
	if percentChange < -10 {
		return "improving"
	} else if percentChange > 10 {
		return "degrading"
	}
	
	return "stable"
}

func (a *WorkflowAnalytics) identifyBottlenecks(executions []*models.WorkflowExecution) []Bottleneck {
	componentTimes := make(map[string]*componentTime)
	
	for _, exec := range executions {
		for _, entry := range exec.ExecutionLog {
			key := fmt.Sprintf("%s:%s", entry.ActionType, entry.ActionID)
			
			if _, exists := componentTimes[key]; !exists {
				componentTimes[key] = &componentTime{
					componentID:   entry.ActionID,
					componentType: entry.ActionType,
				}
			}
			
			ct := componentTimes[key]
			ct.totalTime += entry.Duration
			ct.count++
			
			if entry.Duration > ct.maxTime {
				ct.maxTime = entry.Duration
			}
		}
	}
	
	bottlenecks := []Bottleneck{}
	
	for _, ct := range componentTimes {
		if ct.count > 0 {
			avgTime := ct.totalTime / int64(ct.count)
			
			// Consider it a bottleneck if avg time > 5 seconds
			if avgTime > 5000 {
				impact := "low"
				if avgTime > 10000 {
					impact = "medium"
				}
				if avgTime > 30000 {
					impact = "high"
				}
				
				bottlenecks = append(bottlenecks, Bottleneck{
					ComponentID:   ct.componentID,
					ComponentType: ct.componentType,
					AverageTime:   avgTime,
					MaxTime:       ct.maxTime,
					Frequency:     ct.count,
					Impact:        impact,
				})
			}
		}
	}
	
	// Sort by impact and average time
	sort.Slice(bottlenecks, func(i, j int) bool {
		if bottlenecks[i].Impact != bottlenecks[j].Impact {
			return bottlenecks[i].Impact > bottlenecks[j].Impact
		}
		return bottlenecks[i].AverageTime > bottlenecks[j].AverageTime
	})
	
	return bottlenecks
}

type componentTime struct {
	componentID   string
	componentType string
	totalTime     int64
	maxTime       int64
	count         int
}

func (a *WorkflowAnalytics) generateOptimizationSuggestions(bottlenecks []Bottleneck) []string {
	suggestions := []string{}
	
	for _, bottleneck := range bottlenecks {
		if bottleneck.Impact == "high" {
			suggestions = append(suggestions, fmt.Sprintf(
				"Critical bottleneck in %s (%s): Consider optimizing or parallelizing this component",
				bottleneck.ComponentID, bottleneck.ComponentType,
			))
		}
		
		if bottleneck.MaxTime > bottleneck.AverageTime*3 {
			suggestions = append(suggestions, fmt.Sprintf(
				"High variance in %s performance: Investigate causes of occasional slowdowns",
				bottleneck.ComponentID,
			))
		}
		
		if bottleneck.ComponentType == "webhook" && bottleneck.AverageTime > 10000 {
			suggestions = append(suggestions, fmt.Sprintf(
				"Webhook %s is slow: Consider implementing timeout and retry logic",
				bottleneck.ComponentID,
			))
		}
	}
	
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Workflow is performing well. No critical optimizations needed.")
	}
	
	return suggestions
}

func (a *WorkflowAnalytics) calculateResourceUsage(executions []*models.WorkflowExecution) ResourceUsage {
	usage := ResourceUsage{}
	
	// Aggregate resource metrics from executions
	for _, exec := range executions {
		// These would come from actual monitoring data
		usage.DatabaseQueries += 10 // Placeholder
		usage.APICallCount += 5     // Placeholder
	}
	
	if len(executions) > 0 {
		usage.DatabaseQueries = usage.DatabaseQueries / len(executions)
		usage.APICallCount = usage.APICallCount / len(executions)
	}
	
	return usage
}

func (a *WorkflowAnalytics) calculateSLACompliance(workflow *models.Workflow, executions []*models.WorkflowExecution) float64 {
	if len(executions) == 0 {
		return 100.0
	}
	
	compliantCount := 0
	
	for _, exec := range executions {
		// Check if execution met SLA (placeholder logic)
		if exec.CompletedAt != nil {
			duration := exec.CompletedAt.Sub(exec.StartedAt)
			if duration < 30*time.Minute { // Example SLA: 30 minutes
				compliantCount++
			}
		}
	}
	
	return float64(compliantCount) / float64(len(executions)) * 100
}

func (a *WorkflowAnalytics) identifyTriggerPatterns(executions []*models.WorkflowExecution) []TriggerPattern {
	patterns := []TriggerPattern{}
	
	// Group executions by trigger type
	triggerGroups := make(map[string][]*models.WorkflowExecution)
	for _, exec := range executions {
		triggerGroups[string(exec.TriggerType)] = append(triggerGroups[string(exec.TriggerType)], exec)
	}
	
	// Analyze each trigger type for patterns
	for triggerType, execs := range triggerGroups {
		if len(execs) < 5 {
			continue
		}
		
		// Check for daily patterns
		hourCounts := make(map[int]int)
		for _, exec := range execs {
			hourCounts[exec.StartedAt.Hour()]++
		}
		
		// Find peak hour
		maxCount := 0
		peakHour := 0
		for hour, count := range hourCounts {
			if count > maxCount {
				maxCount = count
				peakHour = hour
			}
		}
		
		if maxCount > len(execs)/4 { // If >25% happen in same hour
			patterns = append(patterns, TriggerPattern{
				Pattern:      fmt.Sprintf("%s at %d:00", triggerType, peakHour),
				Frequency:    maxCount,
				Periodicity:  "daily",
				NextExpected: time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour).Add(time.Duration(peakHour) * time.Hour),
			})
		}
	}
	
	return patterns
}

func (a *WorkflowAnalytics) predictFutureTriggers(patterns []TriggerPattern) []PredictedTrigger {
	predictions := []PredictedTrigger{}
	
	for _, pattern := range patterns {
		prediction := PredictedTrigger{
			TriggerType:  pattern.Pattern,
			PredictedAt:  pattern.NextExpected,
			Probability:  0.75, // Placeholder probability
			BasedOn:      fmt.Sprintf("Historical pattern: %s", pattern.Periodicity),
		}
		
		predictions = append(predictions, prediction)
	}
	
	return predictions
}

func (a *WorkflowAnalytics) identifyActionChains(executions []*models.WorkflowExecution) []ActionChain {
	chainMap := make(map[string]*ActionChain)
	
	for _, exec := range executions {
		if len(exec.ExecutionLog) < 2 {
			continue
		}
		
		// Build action sequence
		actions := []string{}
		for _, entry := range exec.ExecutionLog {
			if entry.ActionType != "" {
				actions = append(actions, entry.ActionType)
			}
		}
		
		if len(actions) > 1 {
			chainKey := fmt.Sprintf("%v", actions)
			
			if _, exists := chainMap[chainKey]; !exists {
				chainMap[chainKey] = &ActionChain{
					Actions: actions,
				}
			}
			
			chain := chainMap[chainKey]
			chain.Frequency++
			
			if exec.Status == "success" {
				chain.SuccessRate++
			}
			
			if exec.CompletedAt != nil {
				duration := exec.CompletedAt.Sub(exec.StartedAt).Milliseconds()
				chain.AverageTime = (chain.AverageTime*int64(chain.Frequency-1) + duration) / int64(chain.Frequency)
			}
		}
	}
	
	// Convert to slice and calculate success rates
	chains := []ActionChain{}
	for _, chain := range chainMap {
		if chain.Frequency > 1 { // Only include chains that appear multiple times
			if chain.Frequency > 0 {
				chain.SuccessRate = chain.SuccessRate / float64(chain.Frequency) * 100
			}
			chains = append(chains, *chain)
		}
	}
	
	// Sort by frequency
	sort.Slice(chains, func(i, j int) bool {
		return chains[i].Frequency > chains[j].Frequency
	})
	
	// Return top 10
	if len(chains) > 10 {
		return chains[:10]
	}
	
	return chains
}

func (a *WorkflowAnalytics) calculateActionImpact(actionType string, stats *actionStat) float64 {
	// Calculate impact based on usage, success rate, and speed
	usageScore := float64(stats.total) * 0.3
	successScore := float64(stats.success) / float64(stats.total) * 100 * 0.5
	speedScore := 0.0
	
	if stats.total > 0 {
		avgTime := stats.totalTime / int64(stats.total)
		if avgTime > 0 {
			speedScore = (5000 / float64(avgTime)) * 20 // Bonus for fast actions
		}
	}
	
	return usageScore + successScore + speedScore
}