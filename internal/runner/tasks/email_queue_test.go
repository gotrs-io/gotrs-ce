package tasks

import (
	"context"
	"errors"
	"io"
	"log"
	"net/textproto"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gotrs-io/gotrs-ce/internal/config"
	"github.com/gotrs-io/gotrs-ce/internal/mailqueue"
)

func TestCleanupFailedEmailsDeletesOnlyOld(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	repo := mailqueue.NewMailQueueRepository(db)
	task := &EmailQueueTask{repo: repo, cfg: &config.EmailConfig{Enabled: true}, logger: log.New(io.Discard, "", 0)}

	ctx := context.Background()
	oldTime := time.Now().Add(-8 * 24 * time.Hour)
	recentTime := time.Now().Add(-48 * time.Hour)

	rows := sqlmock.NewRows([]string{
		"id", "insert_fingerprint", "article_id", "attempts", "sender", "recipient",
		"raw_message", "due_time", "last_smtp_code", "last_smtp_message", "create_time",
	}).
		AddRow(int64(1), nil, nil, MaxRetries, nil, "old@example.com", []byte("raw"), nil, nil, "fail", oldTime).
		AddRow(int64(2), nil, nil, MaxRetries, nil, "recent@example.com", []byte("raw"), nil, nil, "fail", recentTime)

	mock.ExpectQuery("SELECT id, insert_fingerprint.*FROM mail_queue.*WHERE attempts >= .*ORDER BY create_time ASC.*LIMIT ?").
		WithArgs(MaxRetries, 100).
		WillReturnRows(rows)

	mock.ExpectExec("DELETE FROM mail_queue WHERE id = ?").
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := task.cleanupFailedEmails(ctx); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestCleanupFailedEmailsSkipsNonMaxAttempts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	repo := mailqueue.NewMailQueueRepository(db)
	task := &EmailQueueTask{repo: repo, cfg: &config.EmailConfig{Enabled: true}, logger: log.New(io.Discard, "", 0)}

	ctx := context.Background()
	createTime := time.Now().Add(-10 * 24 * time.Hour)

	rows := sqlmock.NewRows([]string{
		"id", "insert_fingerprint", "article_id", "attempts", "sender", "recipient",
		"raw_message", "due_time", "last_smtp_code", "last_smtp_message", "create_time",
	}).
		AddRow(int64(1), nil, nil, MaxRetries-1, nil, "keep@example.com", []byte("raw"), nil, nil, "fail", createTime)

	mock.ExpectQuery("SELECT id, insert_fingerprint.*FROM mail_queue.*WHERE attempts >= .*ORDER BY create_time ASC.*LIMIT ?").
		WithArgs(MaxRetries, 100).
		WillReturnRows(rows)

	if err := task.cleanupFailedEmails(ctx); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
func TestConstants(t *testing.T) {
	if MaxRetries != 5 {
		t.Errorf("expected MaxRetries 5, got %d", MaxRetries)
	}
	if RetryDelayBase != 5 {
		t.Errorf("expected RetryDelayBase 5, got %d", RetryDelayBase)
	}
}

func TestEmailQueueTask_Name(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	task := NewEmailQueueTask(db, &config.EmailConfig{})

	if name := task.Name(); name != "email-queue-processor" {
		t.Errorf("expected Name 'email-queue-processor', got %s", name)
	}
}

func TestEmailQueueTask_Schedule(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	task := NewEmailQueueTask(db, &config.EmailConfig{})

	schedule := task.Schedule()
	if schedule != "*/30 * * * * *" {
		t.Errorf("expected Schedule '*/30 * * * * *', got %s", schedule)
	}
}

func TestEmailQueueTask_Timeout(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	task := NewEmailQueueTask(db, &config.EmailConfig{})

	timeout := task.Timeout()
	if timeout != 5*time.Minute {
		t.Errorf("expected Timeout 5m, got %v", timeout)
	}
}

func TestEmailQueueTask_Run_Disabled(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	task := &EmailQueueTask{
		repo:   mailqueue.NewMailQueueRepository(db),
		cfg:    &config.EmailConfig{Enabled: false},
		logger: log.New(io.Discard, "", 0),
	}

	err = task.Run(context.Background())
	if err != nil {
		t.Errorf("expected no error when disabled, got %v", err)
	}
}

func TestCalculateNextRetryTime(t *testing.T) {
	task := &EmailQueueTask{cfg: &config.EmailConfig{}}

	tests := []struct {
		attempts     int
		expectedDelay time.Duration
		expectNil    bool
	}{
		{1, 5 * time.Minute, false},
		{2, 25 * time.Minute, false},
		{3, 125 * time.Minute, false},
		{4, 625 * time.Minute, false},
		{5, 0, true}, // MaxRetries reached
		{6, 0, true},
	}

	for _, tt := range tests {
		result := task.calculateNextRetryTime(tt.attempts)

		if tt.expectNil {
			if result != nil {
				t.Errorf("attempts=%d: expected nil, got %v", tt.attempts, result)
			}
		} else {
			if result == nil {
				t.Errorf("attempts=%d: expected non-nil", tt.attempts)
				continue
			}

			expectedTime := time.Now().Add(tt.expectedDelay)
			diff := result.Sub(expectedTime)
			if diff < -time.Second || diff > time.Second {
				t.Errorf("attempts=%d: expected delay ~%v, got diff %v", tt.attempts, tt.expectedDelay, diff)
			}
		}
	}
}

func TestStringPtr(t *testing.T) {
	s := "test string"
	ptr := stringPtr(s)

	if ptr == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *ptr != s {
		t.Errorf("expected %q, got %q", s, *ptr)
	}
}

func TestSmtpStatus_TextprotoError(t *testing.T) {
	err := &textproto.Error{Code: 550, Msg: "User unknown"}
	cfg := &config.EmailConfig{}

	code, msg := smtpStatus(err, cfg)

	if code == nil || *code != 550 {
		t.Errorf("expected code 550, got %v", code)
	}
	if msg != "550 User unknown" {
		t.Errorf("expected msg '550 User unknown', got %q", msg)
	}
}

func TestSmtpStatus_EOF(t *testing.T) {
	cfg := &config.EmailConfig{}
	cfg.SMTP.Port = 25

	code, msg := smtpStatus(io.EOF, cfg)

	if code == nil || *code != 421 {
		t.Errorf("expected code 421, got %v", code)
	}
	if msg != "421 unexpected EOF" {
		t.Errorf("expected '421 unexpected EOF', got %q", msg)
	}
}

func TestSmtpStatus_EOF_TestPort(t *testing.T) {
	cfg := &config.EmailConfig{}
	cfg.SMTP.Port = 25253

	code, msg := smtpStatus(io.EOF, cfg)

	if code == nil || *code != 550 {
		t.Errorf("expected code 550 for test port, got %v", code)
	}
	if msg != "550 unexpected EOF" {
		t.Errorf("expected '550 unexpected EOF', got %q", msg)
	}
}

func TestSmtpStatus_OtherError(t *testing.T) {
	err := errors.New("connection refused")
	cfg := &config.EmailConfig{}

	code, msg := smtpStatus(err, cfg)

	if code != nil {
		t.Errorf("expected nil code, got %v", code)
	}
	if msg != "connection refused" {
		t.Errorf("expected 'connection refused', got %q", msg)
	}
}

func TestSendError(t *testing.T) {
	originalErr := errors.New("network error")
	code := 421
	se := &sendError{code: &code, err: originalErr}

	if se.Error() != "network error" {
		t.Errorf("expected Error() 'network error', got %q", se.Error())
	}

	if se.Unwrap() != originalErr {
		t.Error("expected Unwrap() to return original error")
	}
}

func TestSendError_NilCode(t *testing.T) {
	originalErr := errors.New("timeout")
	se := &sendError{code: nil, err: originalErr}

	if se.Error() != "timeout" {
		t.Errorf("expected Error() 'timeout', got %q", se.Error())
	}
}

func TestLoginAuth_Start(t *testing.T) {
	auth := &loginAuth{username: "user@test.com", password: "secret"}

	proto, resp, err := auth.Start(nil)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if proto != "LOGIN" {
		t.Errorf("expected proto 'LOGIN', got %q", proto)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty response, got %v", resp)
	}
}

func TestLoginAuth_Next(t *testing.T) {
	auth := &loginAuth{username: "user@test.com", password: "secret123"}

	tests := []struct {
		fromServer string
		more       bool
		expected   string
		expectErr  bool
	}{
		{"Username:", true, "user@test.com", false},
		{"Password:", true, "secret123", false},
		{"Unknown:", true, "", true},
		{"", false, "", false},
	}

	for _, tt := range tests {
		resp, err := auth.Next([]byte(tt.fromServer), tt.more)

		if tt.expectErr {
			if err == nil {
				t.Errorf("fromServer=%q: expected error", tt.fromServer)
			}
		} else {
			if err != nil {
				t.Errorf("fromServer=%q: unexpected error: %v", tt.fromServer, err)
			}
			if string(resp) != tt.expected {
				t.Errorf("fromServer=%q: expected %q, got %q", tt.fromServer, tt.expected, string(resp))
			}
		}
	}
}