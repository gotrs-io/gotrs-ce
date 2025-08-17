package service

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gotrs-io/gotrs-ce/internal/models"
)

// WorkflowDebugger provides debugging and testing capabilities for workflows
type WorkflowDebugger struct {
	workflowService *WorkflowService
	triggerService  *TriggerService
	mu              sync.RWMutex
	breakpoints     map[string][]Breakpoint
	watchedVars     map[string][]string
	debugSessions   map[string]*DebugSession
	stepMode        map[string]bool
}

// Breakpoint represents a debug breakpoint in a workflow
type Breakpoint struct {
	ID          string                 `json:"id"`
	WorkflowID  int                    `json:"workflow_id"`
	ComponentID string                 `json:"component_id"`
	Type        string                 `json:"type"` // before, after, error
	Condition   string                 `json:"condition"`
	Enabled     bool                   `json:"enabled"`
	HitCount    int                    `json:"hit_count"`
}

// DebugSession represents an active debugging session
type DebugSession struct {
	ID              string                   `json:"id"`
	WorkflowID      int                      `json:"workflow_id"`
	WorkflowName    string                   `json:"workflow_name"`
	StartedAt       time.Time                `json:"started_at"`
	Status          string                   `json:"status"` // running, paused, completed, failed
	CurrentStep     string                   `json:"current_step"`
	ExecutionPath   []ExecutionStep          `json:"execution_path"`
	Variables       map[string]interface{}   `json:"variables"`
	Breakpoints     []Breakpoint             `json:"breakpoints"`
	WatchedVars     []string                 `json:"watched_vars"`
	Logs            []DebugLog               `json:"logs"`
	Performance     PerformanceMetrics       `json:"performance"`
	TestData        map[string]interface{}   `json:"test_data"`
	Coverage        *CoverageReport          `json:"coverage"`
}

// ExecutionStep represents a single step in workflow execution
type ExecutionStep struct {
	ID          string                 `json:"id"`
	ComponentID string                 `json:"component_id"`
	Type        string                 `json:"type"` // trigger, condition, action
	Name        string                 `json:"name"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Duration    int64                  `json:"duration_ms"`
	Status      string                 `json:"status"` // pending, running, completed, failed, skipped
	Input       map[string]interface{} `json:"input"`
	Output      map[string]interface{} `json:"output"`
	Error       string                 `json:"error,omitempty"`
	Variables   map[string]interface{} `json:"variables"`
}

// DebugLog represents a debug log entry
type DebugLog struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"` // debug, info, warning, error
	Component   string                 `json:"component"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// PerformanceMetrics tracks workflow performance
type PerformanceMetrics struct {
	TotalDuration   int64                  `json:"total_duration_ms"`
	StepDurations   map[string]int64       `json:"step_durations"`
	DatabaseQueries int                    `json:"database_queries"`
	APICallCount    int                    `json:"api_call_count"`
	MemoryUsage     int64                  `json:"memory_usage_bytes"`
	CPUTime         int64                  `json:"cpu_time_ms"`
}

// CoverageReport shows workflow test coverage
type CoverageReport struct {
	TotalComponents    int                    `json:"total_components"`
	ExecutedComponents int                    `json:"executed_components"`
	CoveragePercent    float64                `json:"coverage_percent"`
	PathsCovered       []string               `json:"paths_covered"`
	PathsNotCovered    []string               `json:"paths_not_covered"`
	ConditionCoverage  map[string]bool        `json:"condition_coverage"`
}

// TestScenario represents a test scenario for a workflow
type TestScenario struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	WorkflowID      int                    `json:"workflow_id"`
	TestData        map[string]interface{} `json:"test_data"`
	ExpectedResults []ExpectedResult       `json:"expected_results"`
	Assertions      []Assertion            `json:"assertions"`
	CreatedAt       time.Time              `json:"created_at"`
	LastRun         *time.Time             `json:"last_run,omitempty"`
	LastResult      *TestResult            `json:"last_result,omitempty"`
}

// ExpectedResult defines expected outcome of a test
type ExpectedResult struct {
	ComponentID     string                 `json:"component_id"`
	ExpectedOutput  map[string]interface{} `json:"expected_output"`
	ExpectedStatus  string                 `json:"expected_status"`
}

// Assertion represents a test assertion
type Assertion struct {
	Type        string      `json:"type"` // equals, contains, greater_than, etc.
	Field       string      `json:"field"`
	Expected    interface{} `json:"expected"`
	Description string      `json:"description"`
}

// TestResult contains test execution results
type TestResult struct {
	ScenarioID      string                 `json:"scenario_id"`
	ExecutedAt      time.Time              `json:"executed_at"`
	Duration        int64                  `json:"duration_ms"`
	Status          string                 `json:"status"` // passed, failed, error
	PassedAssertions int                   `json:"passed_assertions"`
	FailedAssertions int                   `json:"failed_assertions"`
	Failures        []TestFailure          `json:"failures"`
	Coverage        *CoverageReport        `json:"coverage"`
	Performance     PerformanceMetrics     `json:"performance"`
}

// TestFailure describes a test failure
type TestFailure struct {
	AssertionID string      `json:"assertion_id"`
	Message     string      `json:"message"`
	Expected    interface{} `json:"expected"`
	Actual      interface{} `json:"actual"`
	Location    string      `json:"location"`
}

// NewWorkflowDebugger creates a new workflow debugger
func NewWorkflowDebugger(workflowService *WorkflowService, triggerService *TriggerService) *WorkflowDebugger {
	return &WorkflowDebugger{
		workflowService: workflowService,
		triggerService:  triggerService,
		breakpoints:     make(map[string][]Breakpoint),
		watchedVars:     make(map[string][]string),
		debugSessions:   make(map[string]*DebugSession),
		stepMode:        make(map[string]bool),
	}
}

// StartDebugSession starts a new debug session
func (d *WorkflowDebugger) StartDebugSession(workflowID int, testData map[string]interface{}) (*DebugSession, error) {
	workflow, err := d.workflowService.workflowRepo.GetByID(workflowID)
	if err != nil {
		return nil, fmt.Errorf("workflow not found: %w", err)
	}

	sessionID := fmt.Sprintf("debug_%d_%d", workflowID, time.Now().Unix())
	
	session := &DebugSession{
		ID:            sessionID,
		WorkflowID:    workflowID,
		WorkflowName:  workflow.Name,
		StartedAt:     time.Now(),
		Status:        "running",
		ExecutionPath: []ExecutionStep{},
		Variables:     make(map[string]interface{}),
		Logs:          []DebugLog{},
		TestData:      testData,
		Performance: PerformanceMetrics{
			StepDurations: make(map[string]int64),
		},
		Coverage: &CoverageReport{
			ConditionCoverage: make(map[string]bool),
		},
	}

	d.mu.Lock()
	d.debugSessions[sessionID] = session
	d.mu.Unlock()

	// Start execution with debugging hooks
	go d.executeWithDebug(session, workflow, testData)

	return session, nil
}

// executeWithDebug executes a workflow with debugging enabled
func (d *WorkflowDebugger) executeWithDebug(session *DebugSession, workflow *models.Workflow, testData map[string]interface{}) {
	startTime := time.Now()
	
	// Log start
	d.addLog(session, "info", "workflow", "Starting workflow execution", nil)
	
	// Initialize coverage tracking
	session.Coverage.TotalComponents = len(workflow.Triggers) + len(workflow.Conditions) + len(workflow.Actions)
	
	// Process triggers
	for _, trigger := range workflow.Triggers {
		if d.shouldBreak(session.ID, trigger.ID, "before") {
			d.pause(session, fmt.Sprintf("Breakpoint hit before trigger: %s", trigger.Type))
		}
		
		step := d.startStep(session, trigger.ID, "trigger", string(trigger.Type))
		
		// Simulate trigger execution
		d.simulateTrigger(session, trigger, testData)
		
		d.completeStep(session, step, nil)
		session.Coverage.ExecutedComponents++
		
		if d.shouldBreak(session.ID, trigger.ID, "after") {
			d.pause(session, fmt.Sprintf("Breakpoint hit after trigger: %s", trigger.Type))
		}
	}
	
	// Process conditions
	for _, condition := range workflow.Conditions {
		if d.shouldBreak(session.ID, condition.ID, "before") {
			d.pause(session, fmt.Sprintf("Breakpoint hit before condition: %s", condition.Field))
		}
		
		step := d.startStep(session, condition.ID, "condition", condition.Field)
		
		// Evaluate condition
		result := d.evaluateCondition(session, condition, testData)
		step.Output = map[string]interface{}{"result": result}
		
		d.completeStep(session, step, nil)
		session.Coverage.ExecutedComponents++
		session.Coverage.ConditionCoverage[condition.ID] = result
		
		if !result {
			d.addLog(session, "info", "condition", fmt.Sprintf("Condition %s evaluated to false, skipping branch", condition.Field), nil)
			break // Skip remaining actions if condition fails
		}
		
		if d.shouldBreak(session.ID, condition.ID, "after") {
			d.pause(session, fmt.Sprintf("Breakpoint hit after condition: %s", condition.Field))
		}
	}
	
	// Process actions
	for _, action := range workflow.Actions {
		if d.shouldBreak(session.ID, action.ID, "before") {
			d.pause(session, fmt.Sprintf("Breakpoint hit before action: %s", action.Type))
		}
		
		step := d.startStep(session, action.ID, "action", string(action.Type))
		
		// Simulate action execution
		err := d.simulateAction(session, action, testData)
		
		if err != nil {
			step.Error = err.Error()
			d.completeStep(session, step, err)
			
			if d.shouldBreak(session.ID, action.ID, "error") {
				d.pause(session, fmt.Sprintf("Breakpoint hit on error in action: %s", action.Type))
			}
			
			if !action.ContinueOnErr {
				session.Status = "failed"
				d.addLog(session, "error", "action", fmt.Sprintf("Action %s failed, stopping execution", action.Type), nil)
				break
			}
		} else {
			d.completeStep(session, step, nil)
			session.Coverage.ExecutedComponents++
		}
		
		if d.shouldBreak(session.ID, action.ID, "after") {
			d.pause(session, fmt.Sprintf("Breakpoint hit after action: %s", action.Type))
		}
	}
	
	// Calculate final metrics
	session.Performance.TotalDuration = time.Since(startTime).Milliseconds()
	session.Coverage.CoveragePercent = float64(session.Coverage.ExecutedComponents) / float64(session.Coverage.TotalComponents) * 100
	
	if session.Status == "running" {
		session.Status = "completed"
	}
	
	d.addLog(session, "info", "workflow", fmt.Sprintf("Workflow execution completed in %dms", session.Performance.TotalDuration), nil)
}

// simulateTrigger simulates trigger execution
func (d *WorkflowDebugger) simulateTrigger(session *DebugSession, trigger models.Trigger, testData map[string]interface{}) {
	d.addLog(session, "debug", "trigger", fmt.Sprintf("Processing trigger: %s", trigger.Type), trigger.Config)
	
	// Add trigger data to variables
	session.Variables["trigger_type"] = trigger.Type
	session.Variables["trigger_time"] = time.Now()
	
	// Merge test data into variables
	for k, v := range testData {
		session.Variables[k] = v
	}
}

// evaluateCondition evaluates a condition with test data
func (d *WorkflowDebugger) evaluateCondition(session *DebugSession, condition models.Condition, testData map[string]interface{}) bool {
	d.addLog(session, "debug", "condition", fmt.Sprintf("Evaluating: %s %s %v", condition.Field, condition.Operator, condition.Value), nil)
	
	// Get field value from test data
	fieldValue := d.getFieldValue(condition.Field, testData)
	
	// Evaluate based on operator
	result := d.performComparison(fieldValue, condition.Operator, condition.Value)
	
	d.addLog(session, "debug", "condition", fmt.Sprintf("Result: %v", result), map[string]interface{}{
		"field_value": fieldValue,
		"operator":    condition.Operator,
		"expected":    condition.Value,
	})
	
	return result
}

// simulateAction simulates action execution
func (d *WorkflowDebugger) simulateAction(session *DebugSession, action models.Action, testData map[string]interface{}) error {
	d.addLog(session, "debug", "action", fmt.Sprintf("Executing action: %s", action.Type), action.Config)
	
	// Simulate delay if configured
	if action.DelaySeconds > 0 {
		d.addLog(session, "info", "action", fmt.Sprintf("Waiting %d seconds", action.DelaySeconds), nil)
		time.Sleep(time.Duration(action.DelaySeconds) * time.Second)
	}
	
	// Simulate different action types
	switch action.Type {
	case models.ActionTypeAssignTicket:
		session.Variables["ticket_assigned"] = true
		session.Variables["assigned_to"] = action.Config["assign_to"]
		
	case models.ActionTypeSendEmail:
		session.Variables["email_sent"] = true
		session.Variables["email_template"] = action.Config["template"]
		
	case models.ActionTypeChangeStatus:
		session.Variables["status_changed"] = true
		session.Variables["new_status"] = action.Config["status"]
		
	case models.ActionTypeEscalate:
		session.Variables["escalated"] = true
		session.Variables["escalation_level"] = action.Config["level"]
		
	default:
		// Generic action simulation
		session.Variables[fmt.Sprintf("action_%s_executed", action.Type)] = true
	}
	
	// Simulate random failures for testing
	if testData["simulate_failures"] == true {
		if time.Now().Unix()%5 == 0 {
			return fmt.Errorf("simulated action failure")
		}
	}
	
	return nil
}

// Test scenario execution

// RunTestScenario runs a test scenario
func (d *WorkflowDebugger) RunTestScenario(scenario *TestScenario) (*TestResult, error) {
	startTime := time.Now()
	
	// Start debug session with test data
	session, err := d.StartDebugSession(scenario.WorkflowID, scenario.TestData)
	if err != nil {
		return nil, err
	}
	
	// Wait for execution to complete or timeout
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("test execution timeout")
		case <-ticker.C:
			d.mu.RLock()
			status := session.Status
			d.mu.RUnlock()
			
			if status == "completed" || status == "failed" {
				goto TestComplete
			}
		}
	}
	
TestComplete:
	// Evaluate assertions
	result := &TestResult{
		ScenarioID: scenario.ID,
		ExecutedAt: time.Now(),
		Duration:   time.Since(startTime).Milliseconds(),
		Status:     "passed",
		Failures:   []TestFailure{},
		Coverage:   session.Coverage,
		Performance: session.Performance,
	}
	
	for _, assertion := range scenario.Assertions {
		if d.evaluateAssertion(session, assertion) {
			result.PassedAssertions++
		} else {
			result.FailedAssertions++
			result.Status = "failed"
			
			failure := TestFailure{
				AssertionID: assertion.Field,
				Message:     assertion.Description,
				Expected:    assertion.Expected,
				Actual:      session.Variables[assertion.Field],
				Location:    fmt.Sprintf("Field: %s", assertion.Field),
			}
			result.Failures = append(result.Failures, failure)
		}
	}
	
	return result, nil
}

// evaluateAssertion evaluates a test assertion
func (d *WorkflowDebugger) evaluateAssertion(session *DebugSession, assertion Assertion) bool {
	actual := session.Variables[assertion.Field]
	
	switch assertion.Type {
	case "equals":
		return actual == assertion.Expected
	case "not_equals":
		return actual != assertion.Expected
	case "contains":
		if str, ok := actual.(string); ok {
			if expected, ok := assertion.Expected.(string); ok {
				return len(str) > 0 && len(expected) > 0 && str == expected
			}
		}
		return false
	case "greater_than":
		// Type assertions and comparisons would be needed
		return false
	case "exists":
		return actual != nil
	case "not_exists":
		return actual == nil
	default:
		return false
	}
}

// Helper methods

// startStep starts tracking a workflow step
func (d *WorkflowDebugger) startStep(session *DebugSession, componentID, stepType, name string) *ExecutionStep {
	step := &ExecutionStep{
		ID:          fmt.Sprintf("%s_%d", componentID, time.Now().UnixNano()),
		ComponentID: componentID,
		Type:        stepType,
		Name:        name,
		StartTime:   time.Now(),
		Status:      "running",
		Input:       make(map[string]interface{}),
		Output:      make(map[string]interface{}),
		Variables:   make(map[string]interface{}),
	}
	
	// Copy current variables
	for k, v := range session.Variables {
		step.Variables[k] = v
	}
	
	d.mu.Lock()
	session.ExecutionPath = append(session.ExecutionPath, *step)
	session.CurrentStep = step.ID
	d.mu.Unlock()
	
	return step
}

// completeStep completes tracking a workflow step
func (d *WorkflowDebugger) completeStep(session *DebugSession, step *ExecutionStep, err error) {
	endTime := time.Now()
	step.EndTime = &endTime
	step.Duration = endTime.Sub(step.StartTime).Milliseconds()
	
	if err != nil {
		step.Status = "failed"
		step.Error = err.Error()
	} else {
		step.Status = "completed"
	}
	
	d.mu.Lock()
	// Update step in execution path
	for i, s := range session.ExecutionPath {
		if s.ID == step.ID {
			session.ExecutionPath[i] = *step
			break
		}
	}
	
	// Track performance
	session.Performance.StepDurations[step.ComponentID] = step.Duration
	d.mu.Unlock()
}

// addLog adds a debug log entry
func (d *WorkflowDebugger) addLog(session *DebugSession, level, component, message string, data map[string]interface{}) {
	log := DebugLog{
		Timestamp: time.Now(),
		Level:     level,
		Component: component,
		Message:   message,
		Data:      data,
	}
	
	d.mu.Lock()
	session.Logs = append(session.Logs, log)
	d.mu.Unlock()
}

// shouldBreak checks if execution should break at a breakpoint
func (d *WorkflowDebugger) shouldBreak(sessionID, componentID, breakType string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	if d.stepMode[sessionID] {
		return true
	}
	
	breakpoints, exists := d.breakpoints[sessionID]
	if !exists {
		return false
	}
	
	for _, bp := range breakpoints {
		if bp.Enabled && bp.ComponentID == componentID && bp.Type == breakType {
			bp.HitCount++
			return true
		}
	}
	
	return false
}

// pause pauses execution at a breakpoint
func (d *WorkflowDebugger) pause(session *DebugSession, reason string) {
	d.mu.Lock()
	session.Status = "paused"
	d.mu.Unlock()
	
	d.addLog(session, "info", "debugger", fmt.Sprintf("Execution paused: %s", reason), nil)
	
	// Wait for resume signal
	// In a real implementation, this would wait for a resume command
	time.Sleep(1 * time.Second)
	
	d.mu.Lock()
	session.Status = "running"
	d.mu.Unlock()
}

// getFieldValue extracts a field value from test data
func (d *WorkflowDebugger) getFieldValue(field string, data map[string]interface{}) interface{} {
	// Handle nested fields with dot notation
	parts := []string{}
	current := data
	
	for _, part := range parts {
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return current[part]
		}
	}
	
	return nil
}

// performComparison performs a comparison operation
func (d *WorkflowDebugger) performComparison(value interface{}, operator models.ConditionOperator, expected interface{}) bool {
	switch operator {
	case models.OperatorEquals:
		return value == expected
	case models.OperatorNotEquals:
		return value != expected
	case models.OperatorGreaterThan:
		// Would need type assertions and numeric comparison
		return false
	case models.OperatorLessThan:
		// Would need type assertions and numeric comparison
		return false
	case models.OperatorContains:
		// Would need string comparison
		return false
	case models.OperatorIsEmpty:
		return value == nil || value == ""
	case models.OperatorIsNotEmpty:
		return value != nil && value != ""
	default:
		return false
	}
}

// Breakpoint management

// SetBreakpoint sets a breakpoint
func (d *WorkflowDebugger) SetBreakpoint(sessionID string, breakpoint Breakpoint) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if _, exists := d.breakpoints[sessionID]; !exists {
		d.breakpoints[sessionID] = []Breakpoint{}
	}
	
	d.breakpoints[sessionID] = append(d.breakpoints[sessionID], breakpoint)
}

// RemoveBreakpoint removes a breakpoint
func (d *WorkflowDebugger) RemoveBreakpoint(sessionID, breakpointID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if breakpoints, exists := d.breakpoints[sessionID]; exists {
		filtered := []Breakpoint{}
		for _, bp := range breakpoints {
			if bp.ID != breakpointID {
				filtered = append(filtered, bp)
			}
		}
		d.breakpoints[sessionID] = filtered
	}
}

// EnableStepMode enables step-by-step execution
func (d *WorkflowDebugger) EnableStepMode(sessionID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stepMode[sessionID] = true
}

// DisableStepMode disables step-by-step execution
func (d *WorkflowDebugger) DisableStepMode(sessionID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stepMode[sessionID] = false
}

// GetDebugSession returns a debug session
func (d *WorkflowDebugger) GetDebugSession(sessionID string) (*DebugSession, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	session, exists := d.debugSessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("debug session not found")
	}
	
	return session, nil
}

// ExportDebugReport exports a debug report
func (d *WorkflowDebugger) ExportDebugReport(sessionID string) ([]byte, error) {
	session, err := d.GetDebugSession(sessionID)
	if err != nil {
		return nil, err
	}
	
	report := map[string]interface{}{
		"session":     session,
		"exported_at": time.Now(),
		"version":     "1.0.0",
	}
	
	return json.MarshalIndent(report, "", "  ")
}