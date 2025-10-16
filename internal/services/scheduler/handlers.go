package scheduler

import (
	"context"
	"strconv"
	"strings"

	"github.com/gotrs-io/gotrs-ce/internal/models"
)

func (s *Service) registerBuiltinHandlers() {
	s.RegisterHandler("ticket.autoClose", s.handleAutoClose)
	s.RegisterHandler("email.poll", s.handleEmailPoll)
	s.RegisterHandler("scheduler.housekeeping", s.handleHousekeeping)
}

func (s *Service) handleAutoClose(ctx context.Context, job *models.ScheduledJob) error {
	if s.ticketRepo == nil {
		s.logger.Printf("scheduler: ticket repository unavailable, skipping autoClose")
		return nil
	}
	transitions := map[string]string{
		"pending auto close+": "closed successful",
		"pending auto close-": "closed unsuccessful",
	}
	if cfg := transitionsFromConfig(job.Config); len(cfg) > 0 {
		transitions = cfg
	}

	systemUserID := intFromConfig(job.Config, "system_user_id", 1)
	result, err := s.ticketRepo.AutoClosePendingTickets(ctx, s.now(), transitions, systemUserID)
	if err != nil {
		return err
	}
	if result != nil {
		s.logger.Printf("scheduler: autoClose transitioned %d ticket(s) %+v", result.Total, result.Transitions)
	}
	return nil
}

func (s *Service) handleEmailPoll(ctx context.Context, job *models.ScheduledJob) error {
	s.logger.Printf("scheduler: email poll starting")
	if s.emailRepo == nil {
		s.logger.Printf("scheduler: email repository unavailable, skipping poll")
		return nil
	}
	accounts, err := s.emailRepo.GetActiveAccounts()
	if err != nil {
		return err
	}
	if len(accounts) == 0 {
		s.logger.Printf("scheduler: email poll found no active accounts")
		return nil
	}

	limit := intFromConfig(job.Config, "max_accounts", 5)
	count := len(accounts)
	if limit > 0 && count > limit {
		count = limit
	}
	// Placeholder until POP3/IMAP fetchers land.
	s.logger.Printf("scheduler: email poll queued %d account(s) for fetch", count)
	return nil
}

func (s *Service) handleHousekeeping(ctx context.Context, job *models.ScheduledJob) error {
	s.logger.Printf("scheduler: housekeeping placeholder running")
	return nil
}

func defaultJobs() []*models.ScheduledJob {
	return []*models.ScheduledJob{
		{
			Name:           "Auto-close Pending Tickets",
			Slug:           "pending-auto-close",
			Handler:        "ticket.autoClose",
			Schedule:       "*/5 * * * *",
			TimeoutSeconds: 120,
			Config: map[string]any{
				"transitions": map[string]string{
					"pending auto close+": "closed successful",
					"pending auto close-": "closed unsuccessful",
				},
				"system_user_id": 1,
			},
		},
		{
			Name:           "Email Account Poller",
			Slug:           "email-ingest",
			Handler:        "email.poll",
			Schedule:       "*/2 * * * *",
			TimeoutSeconds: 300,
			Config: map[string]any{
				"max_accounts": 5,
			},
		},
		{
			Name:           "Scheduler Housekeeping",
			Slug:           "scheduler-housekeeping",
			Handler:        "scheduler.housekeeping",
			Schedule:       "0 3 * * *",
			TimeoutSeconds: 600,
			Config: map[string]any{
				"retention_days": 30,
			},
		},
	}
}

func intFromConfig(cfg map[string]any, key string, def int) int {
	if cfg == nil {
		return def
	}
	val, ok := cfg[key]
	if !ok {
		return def
	}
	switch v := val.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err == nil {
			return n
		}
	}
	return def
}

func transitionsFromConfig(cfg map[string]any) map[string]string {
	result := make(map[string]string)
	if cfg == nil {
		return result
	}
	raw, ok := cfg["transitions"]
	if !ok {
		return result
	}
	switch t := raw.(type) {
	case map[string]any:
		for k, v := range t {
			name := strings.TrimSpace(k)
			if name == "" {
				continue
			}
			if str, ok := v.(string); ok {
				result[name] = strings.TrimSpace(str)
			}
		}
	case map[string]string:
		for k, v := range t {
			result[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return result
}
