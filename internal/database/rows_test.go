package database

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestScanRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	t.Run("successful scan", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice").
			AddRow(2, "Bob").
			AddRow(3, "Charlie")
		mock.ExpectQuery("SELECT").WillReturnRows(rows)

		result, err := db.Query("SELECT id, name FROM users")
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer result.Close()

		type user struct {
			ID   int
			Name string
		}
		var users []user

		err = ScanRows(result, func(r *sql.Rows) error {
			var u user
			if err := r.Scan(&u.ID, &u.Name); err != nil {
				return err
			}
			users = append(users, u)
			return nil
		})

		if err != nil {
			t.Errorf("ScanRows returned error: %v", err)
		}
		if len(users) != 3 {
			t.Errorf("expected 3 users, got %d", len(users))
		}
		if users[0].Name != "Alice" {
			t.Errorf("expected first user Alice, got %s", users[0].Name)
		}
	})

	t.Run("scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id"}).
			AddRow("not-an-int") // This will cause scan error for int
		mock.ExpectQuery("SELECT").WillReturnRows(rows)

		result, err := db.Query("SELECT id FROM users")
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer result.Close()

		err = ScanRows(result, func(r *sql.Rows) error {
			var id int
			return r.Scan(&id)
		})

		if err == nil {
			t.Error("expected scan error, got nil")
		}
	})

	t.Run("empty result set", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id"})
		mock.ExpectQuery("SELECT").WillReturnRows(rows)

		result, err := db.Query("SELECT id FROM users")
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer result.Close()

		var count int
		err = ScanRows(result, func(r *sql.Rows) error {
			count++
			return nil
		})

		if err != nil {
			t.Errorf("ScanRows returned error: %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0 iterations, got %d", count)
		}
	})
}

func TestCollectRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	t.Run("collect structs", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice").
			AddRow(2, "Bob")
		mock.ExpectQuery("SELECT").WillReturnRows(rows)

		result, err := db.Query("SELECT id, name FROM users")
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer result.Close()

		type user struct {
			ID   int
			Name string
		}

		users, err := CollectRows(result, func(r *sql.Rows) (user, error) {
			var u user
			err := r.Scan(&u.ID, &u.Name)
			return u, err
		})

		if err != nil {
			t.Errorf("CollectRows returned error: %v", err)
		}
		if len(users) != 2 {
			t.Errorf("expected 2 users, got %d", len(users))
		}
	})

	t.Run("collect pointers", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"name"}).
			AddRow("Alice").
			AddRow("Bob")
		mock.ExpectQuery("SELECT").WillReturnRows(rows)

		result, err := db.Query("SELECT name FROM users")
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		defer result.Close()

		type user struct {
			Name string
		}

		users, err := CollectRows(result, func(r *sql.Rows) (*user, error) {
			var u user
			err := r.Scan(&u.Name)
			return &u, err
		})

		if err != nil {
			t.Errorf("CollectRows returned error: %v", err)
		}
		if len(users) != 2 {
			t.Errorf("expected 2 users, got %d", len(users))
		}
		if users[0].Name != "Alice" {
			t.Errorf("expected first user Alice, got %s", users[0].Name)
		}
	})
}

func TestCollectStrings(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"name"}).
		AddRow("Alice").
		AddRow("Bob").
		AddRow("Charlie")
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	result, err := db.Query("SELECT name FROM users")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer result.Close()

	names, err := CollectStrings(result)
	if err != nil {
		t.Errorf("CollectStrings returned error: %v", err)
	}
	if len(names) != 3 {
		t.Errorf("expected 3 names, got %d", len(names))
	}
	if names[1] != "Bob" {
		t.Errorf("expected second name Bob, got %s", names[1])
	}
}

func TestCollectInts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1).
		AddRow(2).
		AddRow(3)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	result, err := db.Query("SELECT id FROM users")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer result.Close()

	ids, err := CollectInts(result)
	if err != nil {
		t.Errorf("CollectInts returned error: %v", err)
	}
	if len(ids) != 3 {
		t.Errorf("expected 3 ids, got %d", len(ids))
	}
	if ids[0] != 1 || ids[2] != 3 {
		t.Errorf("unexpected ids: %v", ids)
	}
}

func TestCollectInt64s(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(int64(1000000000001)).
		AddRow(int64(1000000000002))
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	result, err := db.Query("SELECT id FROM users")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer result.Close()

	ids, err := CollectInt64s(result)
	if err != nil {
		t.Errorf("CollectInt64s returned error: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 ids, got %d", len(ids))
	}
}

func TestMustCloseRows(t *testing.T) {
	t.Run("no existing error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create mock: %v", err)
		}
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
		mock.ExpectQuery("SELECT").WillReturnRows(rows)

		result, _ := db.Query("SELECT id FROM users")
		var funcErr error
		MustCloseRows(result, &funcErr)

		if funcErr != nil {
			t.Errorf("expected no error, got: %v", funcErr)
		}
	})

	t.Run("preserves existing error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create mock: %v", err)
		}
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
		mock.ExpectQuery("SELECT").WillReturnRows(rows)

		result, _ := db.Query("SELECT id FROM users")
		existingErr := errors.New("existing error")
		funcErr := existingErr
		MustCloseRows(result, &funcErr)

		if funcErr != existingErr {
			t.Errorf("expected existing error to be preserved")
		}
	})
}

func TestRowsErr(t *testing.T) {
	t.Run("returns existing error when set", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create mock: %v", err)
		}
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
		mock.ExpectQuery("SELECT").WillReturnRows(rows)

		result, _ := db.Query("SELECT id FROM users")
		defer result.Close()
		// Consume rows
		for result.Next() {
			var id int
			_ = result.Scan(&id)
		}

		existingErr := errors.New("existing error")
		err = RowsErr(result, existingErr)
		if err != existingErr {
			t.Errorf("expected existing error, got: %v", err)
		}
	})

	t.Run("returns nil when no errors", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create mock: %v", err)
		}
		defer db.Close()

		rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
		mock.ExpectQuery("SELECT").WillReturnRows(rows)

		result, _ := db.Query("SELECT id FROM users")
		defer result.Close()
		// Consume rows
		for result.Next() {
			var id int
			_ = result.Scan(&id)
		}

		err = RowsErr(result, nil)
		if err != nil {
			t.Errorf("expected nil, got: %v", err)
		}
	})
}
