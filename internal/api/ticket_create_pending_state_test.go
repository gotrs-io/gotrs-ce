package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/database"
	"github.com/gotrs-io/gotrs-ce/internal/models"
	"github.com/gotrs-io/gotrs-ce/internal/repository"
	"github.com/stretchr/testify/require"
)

func TestHandleCreateTicket_PendingStateWithDueDate(t *testing.T) {
	t.Setenv("APP_ENV", "unit-real")
	t.Setenv("DB_DRIVER", "postgres")
	gin.SetMode(gin.TestMode)

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock setup failed: %v", err)
	}
	t.Cleanup(func() { _ = mockDB.Close() })
	database.SetDB(mockDB)
	t.Cleanup(database.ResetDB)

	repository.SetTicketNumberGenerator(stubGen{n: "202510050001"}, stubStore{})
	t.Cleanup(func() { repository.SetTicketNumberGenerator(nil, nil) })

	pendingStateID := 42
	pendingUntil := "2030-05-20T14:30"
	expectedPending, perr := time.Parse("2006-01-02T15:04", pendingUntil)
	require.NoError(t, perr)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM queue")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, type_id, valid_id")).
		WithArgs(pendingStateID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "type_id", "valid_id", "create_time", "create_by", "change_time", "change_by"}).
			AddRow(pendingStateID, "pending auto close", 5, 1, time.Now(), 1, time.Now(), 1))

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO ticket (")).
		WithArgs(
			"202510050001",
			"Pending Example",
			1,
			models.TicketUnlocked,
			nil,
			nil,
			nil,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			nil,
			nil,
			pendingStateID,
			3,
			0,
			int(expectedPending.Unix()),
			0,
			0,
			0,
			0,
			0,
			sqlmock.AnyArg(),
			1,
			sqlmock.AnyArg(),
			1,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(555))

	router := gin.New()
	router.POST("/api/tickets", handleCreateTicket)

	form := url.Values{
		"subject":        {"Pending Example"},
		"description":    {"Body"},
		"customer_email": {"customer@example.com"},
		"queue_id":       {"1"},
		"next_state":     {"pending-auto-close"},
		"next_state_id":  {"42"},
		"pending_until":  {pendingUntil},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/tickets", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code, "unexpected status: %s", w.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, true, resp["success"])
	require.Equal(t, float64(555), resp["id"])

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestHandleCreateTicket_PendingStateWithoutDueDateFails(t *testing.T) {
	t.Setenv("APP_ENV", "unit-real")
	t.Setenv("DB_DRIVER", "postgres")
	gin.SetMode(gin.TestMode)

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock setup failed: %v", err)
	}
	t.Cleanup(func() { _ = mockDB.Close() })
	database.SetDB(mockDB)
	t.Cleanup(database.ResetDB)

	repository.SetTicketNumberGenerator(stubGen{n: "202510050001"}, stubStore{})
	t.Cleanup(func() { repository.SetTicketNumberGenerator(nil, nil) })

	pendingStateID := 42

	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM queue")).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, type_id, valid_id")).
		WithArgs(pendingStateID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "type_id", "valid_id", "create_time", "create_by", "change_time", "change_by"}).
			AddRow(pendingStateID, "pending auto close", 5, 1, time.Now(), 1, time.Now(), 1))

	router := gin.New()
	router.POST("/api/tickets", handleCreateTicket)

	form := url.Values{
		"subject":        {"Pending Example"},
		"description":    {"Body"},
		"customer_email": {"customer@example.com"},
		"queue_id":       {"1"},
		"next_state_id":  {"42"},
	}

	req := httptest.NewRequest(http.MethodPost, "/api/tickets", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, false, resp["success"])
	require.Contains(t, strings.ToLower(resp["error"].(string)), "pending state requires pending until")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
