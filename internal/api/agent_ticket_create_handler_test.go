package api

import (
	"bytes"
	"context"
	"database/sql/driver"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/database"
	"github.com/gotrs-io/gotrs-ce/internal/repository"
	"github.com/gotrs-io/gotrs-ce/internal/ticketnumber"
)

type ticketNumberStub struct{ value string }

func (s ticketNumberStub) Name() string      { return "Date" }
func (s ticketNumberStub) IsDateBased() bool { return true }
func (s ticketNumberStub) Next(ctx context.Context, store ticketnumber.CounterStore) (string, error) {
	return s.value, nil
}

type noopCounterStore struct{}

func (noopCounterStore) Add(ctx context.Context, dateScoped bool, offset int64) (int64, error) {
	return 1, nil
}

type equalsInt struct{ want int }

func (e equalsInt) Match(v driver.Value) bool {
	switch val := v.(type) {
	case int:
		return val == e.want
	case int32:
		return int(val) == e.want
	case int64:
		return int(val) == e.want
	case uint:
		return int(val) == e.want
	case uint32:
		return int(val) == e.want
	case uint64:
		return int(val) == e.want
	case string:
		if len(val) == 0 {
			return e.want == 0
		}
	case nil:
		return e.want == 0
	}
	return false
}

func TestHandleAgentCreateTicket_UsesSelectedNextState(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock setup failed: %v", err)
	}
	defer mockDB.Close()

	repository.SetTicketNumberGenerator(ticketNumberStub{value: "202510050001"}, noopCounterStore{})
	t.Cleanup(func() { repository.SetTicketNumberGenerator(nil, nil) })
	database.SetDB(mockDB)

	stateRows := sqlmock.NewRows([]string{"id", "name", "type_id", "valid_id", "create_time", "create_by", "change_time", "change_by"}).
		AddRow(5, "pending reminder", 4, 1, time.Now(), 1, time.Now(), 1)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, type_id, valid_id,
	       create_time, create_by, change_time, change_by
	FROM ticket_state
	WHERE id = $1`)).
		WithArgs(5).
		WillReturnRows(stateRows)

	pendingUntil := "2025-10-18T15:30"
	pendingUnix := parsePendingUntil(pendingUntil)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO ticket (")).
		WithArgs(
			"202510050001",
			"Pending state ticket",
			1,
			1,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			equalsInt{want: 5},
			3,
			sqlmock.AnyArg(),
			equalsInt{want: pendingUnix},
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			equalsInt{want: 42},
			sqlmock.AnyArg(),
			equalsInt{want: 42},
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(987))

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'article' AND column_name = 'article_type_id'")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'article' AND column_name = 'communication_channel_id'")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO article (")).
		WithArgs(
			987,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(321))

	mock.ExpectExec("INSERT INTO article_data_mime").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("UPDATE ticket ").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint(42))
		c.Next()
	})
	router.POST("/tickets", HandleAgentCreateTicket(mockDB))

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("subject", "Pending state ticket"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.WriteField("body", "Details"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.WriteField("queue_id", "1"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.WriteField("priority", "3"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.WriteField("type_id", "1"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.WriteField("interaction_type", "internal_note"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.WriteField("next_state_id", "5"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.WriteField("pending_until", pendingUntil); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer close: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/tickets", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusSeeOther)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleAgentCreateTicket_PendingStateRequiresPendingUntil(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock setup failed: %v", err)
	}
	defer mockDB.Close()

	repository.SetTicketNumberGenerator(ticketNumberStub{value: "202510050002"}, noopCounterStore{})
	t.Cleanup(func() { repository.SetTicketNumberGenerator(nil, nil) })
	database.SetDB(mockDB)

	stateRows := sqlmock.NewRows([]string{"id", "name", "type_id", "valid_id", "create_time", "create_by", "change_time", "change_by"}).
		AddRow(5, "pending auto close+", 5, 1, time.Now(), 1, time.Now(), 1)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, type_id, valid_id,
	       create_time, create_by, change_time, change_by
	FROM ticket_state
	WHERE id = $1`)).
		WithArgs(5).
		WillReturnRows(stateRows)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint(99))
		c.Next()
	})
	router.POST("/tickets", HandleAgentCreateTicket(mockDB))

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("subject", "Pending without until"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.WriteField("body", "Body"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.WriteField("queue_id", "1"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.WriteField("priority", "3"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.WriteField("type_id", "1"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.WriteField("next_state_id", "5"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer close: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/tickets", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when pending_until missing, got %d", rec.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHandleAgentCreateTicket_PendingStateSetsPendingUntil(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock setup failed: %v", err)
	}
	defer mockDB.Close()

	repository.SetTicketNumberGenerator(ticketNumberStub{value: "202510050003"}, noopCounterStore{})
	t.Cleanup(func() { repository.SetTicketNumberGenerator(nil, nil) })
	database.SetDB(mockDB)

	stateRows := sqlmock.NewRows([]string{"id", "name", "type_id", "valid_id", "create_time", "create_by", "change_time", "change_by"}).
		AddRow(5, "pending auto close+", 5, 1, time.Now(), 1, time.Now(), 1)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, type_id, valid_id,
	       create_time, create_by, change_time, change_by
	FROM ticket_state
	WHERE id = $1`)).
		WithArgs(5).
		WillReturnRows(stateRows)

	pendingUntil := "2025-10-18T15:30"
	pendingUnix := parsePendingUntil(pendingUntil)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO ticket (")).
		WithArgs(
			"202510050003",
			"Pending auto-close",
			1,
			1,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			equalsInt{want: 5},
			3,
			sqlmock.AnyArg(),
			equalsInt{want: pendingUnix},
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			equalsInt{want: 42},
			sqlmock.AnyArg(),
			equalsInt{want: 42},
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(654))

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'article' AND column_name = 'article_type_id'")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'article' AND column_name = 'communication_channel_id'")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO article (")).
		WithArgs(
			654,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(777))

	mock.ExpectExec("INSERT INTO article_data_mime").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("UPDATE ticket ").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint(42))
		c.Next()
	})
	router.POST("/tickets", HandleAgentCreateTicket(mockDB))

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("subject", "Pending auto-close"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.WriteField("body", "Body"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.WriteField("queue_id", "1"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.WriteField("priority", "3"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.WriteField("type_id", "1"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.WriteField("interaction_type", "internal_note"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.WriteField("next_state_id", "5"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.WriteField("pending_until", pendingUntil); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer close: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/tickets", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusSeeOther)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
