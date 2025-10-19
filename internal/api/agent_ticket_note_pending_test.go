package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/gotrs-io/gotrs-ce/internal/config"
	"github.com/stretchr/testify/require"
)

func TestHandleAgentTicketNotePendingReminderPersistsPendingUntil(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)

	if err := config.LoadFromFile(filepath.Join(baseDir, "testdata", "config_timezone.yaml")); err != nil {
		t.Fatalf("failed to load test config: %v", err)
	}
	t.Cleanup(func() {
		if err := config.LoadFromFile(filepath.Join(baseDir, "testdata", "config_timezone_utc.yaml")); err != nil {
			t.Fatalf("failed to restore config: %v", err)
		}
	})

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock setup failed: %v", err)
	}
	defer mockDB.Close()

	stateRows := sqlmock.NewRows([]string{"id", "name", "type_id", "comments", "valid_id", "create_time", "create_by", "change_time", "change_by"}).
		AddRow(4, "pending reminder", 4, "", 1, time.Now(), 1, time.Now(), 1)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, type_id, comments, valid_id,
       create_time, create_by, change_time, change_by
    FROM ticket_state
    WHERE id = $1`)).
		WithArgs(uint(4)).
		WillReturnRows(stateRows)

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO article (ticket_id, article_sender_type_id, communication_channel_id, is_visible_for_customer,
                                create_time, create_by, change_time, change_by)
            VALUES ($1, 1, $2, $3, CURRENT_TIMESTAMP, $4, CURRENT_TIMESTAMP, $4)
            RETURNING id`)).
		WithArgs("123", 3, 0, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(555))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO article_data_mime (article_id, a_from, a_subject, a_body, a_content_type, incoming_time, create_time, create_by, change_time, change_by)
            VALUES ($1, 'Agent', $2, $3, $4, $5, CURRENT_TIMESTAMP, $6, CURRENT_TIMESTAMP, $6)`)).
		WithArgs(int64(555), "Internal Note", "Pending reminder set", "text/plain", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	loc, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	pendingLocal := time.Date(2025, time.October, 18, 9, 30, 0, 0, loc)
	expectedUnix := pendingLocal.UTC().Unix()

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE ticket
                SET ticket_state_id = $1, until_time = $2, change_time = CURRENT_TIMESTAMP, change_by = $3
                WHERE id = $4`)).
		WithArgs(4, expectedUnix, sqlmock.AnyArg(), "123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	handler := handleAgentTicketNote(mockDB)

	form := url.Values{}
	form.Set("body", "Pending reminder set")
	form.Set("next_state_id", "4")
	form.Set("pending_until", "2025-10-18T09:30")

	req := httptest.NewRequest(http.MethodPost, "/agent/tickets/123/note", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "123"}}
	c.Set("user_id", uint(99))

	handler(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}
