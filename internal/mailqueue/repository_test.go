package mailqueue

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetFailedOrdersOldestFirstAndRespectsLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewMailQueueRepository(db)
	ctx := context.Background()

	oldTime := time.Now().Add(-7 * 24 * time.Hour)
	newerTime := time.Now().Add(-6 * 24 * time.Hour)

	q := `
        SELECT id, insert_fingerprint, article_id, attempts, sender, recipient,
               raw_message, due_time, last_smtp_code, last_smtp_message, create_time
        FROM mail_queue
        WHERE attempts >= ?
        ORDER BY create_time ASC
        LIMIT ?
    `

	rows := sqlmock.NewRows([]string{
		"id", "insert_fingerprint", "article_id", "attempts", "sender", "recipient",
		"raw_message", "due_time", "last_smtp_code", "last_smtp_message", "create_time",
	}).
		AddRow(int64(1), nil, nil, 5, nil, "a@example.com", []byte("raw1"), nil, nil, "fail1", oldTime).
		AddRow(int64(2), nil, nil, 5, nil, "b@example.com", []byte("raw2"), nil, nil, "fail2", newerTime)

	mock.ExpectQuery(regexp.QuoteMeta(q)).
		WithArgs(5, 2).
		WillReturnRows(rows)

	got, err := repo.GetFailed(ctx, 5, 2)
	if err != nil {
		t.Fatalf("GetFailed: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("unexpected count: %d", len(got))
	}
	if got[0].ID != 1 || got[1].ID != 2 {
		t.Fatalf("unexpected order: %v %v", got[0].ID, got[1].ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestBuildEmailMessageWithThreadingAddsThreadHeaders(t *testing.T) {
	body := "<p>Hello</p>"
	raw := BuildEmailMessageWithThreading("from@example.com", "to@example.com", "Hi", body, "example.com", "parent@example.com", "parent@example.com child@example.com")
	msg := string(raw)

	re := regexp.MustCompile(`Message-ID:\s*<[^>]+@example\.com>`)
	if !re.MatchString(msg) {
		t.Fatalf("missing or malformed Message-ID: %s", msg)
	}

	if !strings.Contains(msg, "In-Reply-To: parent@example.com") {
		t.Fatalf("missing In-Reply-To header: %s", msg)
	}
	if !strings.Contains(msg, "References: parent@example.com child@example.com") {
		t.Fatalf("missing References header: %s", msg)
	}
	if !strings.Contains(msg, "Content-Type: text/html; charset=UTF-8") {
		t.Fatalf("expected HTML content type: %s", msg)
	}
}

func TestBuildEmailMessageWithThreadingGeneratesMessageIDWhenThreadingEmpty(t *testing.T) {
	raw := BuildEmailMessageWithThreading("from@example.com", "to@example.com", "Hi", "plain", "example.com", "", "")
	msg := string(raw)

	re := regexp.MustCompile(`Message-ID:\s*<[^>]+@example\.com>`)
	if !re.MatchString(msg) {
		t.Fatalf("missing Message-ID: %s", msg)
	}
	if strings.Contains(msg, "In-Reply-To:") {
		t.Fatalf("unexpected In-Reply-To header: %s", msg)
	}
	if strings.Contains(msg, "References:") {
		t.Fatalf("unexpected References header: %s", msg)
	}
}

func TestExtractMessageIDFromRawMessage(t *testing.T) {
	raw := BuildEmailMessageWithThreading("from@example.com", "to@example.com", "Hi", "plain", "example.com", "", "")
	id := ExtractMessageIDFromRawMessage(raw)

	if id == "" {
		t.Fatalf("expected message id")
	}
	if strings.Contains(id, "<") || strings.Contains(id, ">") {
		t.Fatalf("expected trimmed id, got %s", id)
	}
	if !strings.HasSuffix(id, "@example.com") {
		t.Fatalf("unexpected domain in id: %s", id)
	}
}
