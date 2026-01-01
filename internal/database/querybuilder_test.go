package database

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelectBuilder_SimpleQuery(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("error closing db: %v", err)
		}
	}()

	qb, err := NewQueryBuilder(db)
	require.NoError(t, err)

	query, args, err := qb.NewSelect("id", "name").
		From("users").
		Where("active = ?", true).
		OrderBy("name ASC").
		Limit(10).
		Offset(5).
		ToSQL()

	require.NoError(t, err)
	assert.Contains(t, query, "SELECT id, name FROM users")
	assert.Contains(t, query, "WHERE active =")
	assert.Contains(t, query, "ORDER BY name ASC")
	assert.Len(t, args, 3) // true, 10, 5
}

func TestSelectBuilder_WithJoins(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("error closing db: %v", err)
		}
	}()

	qb, err := NewQueryBuilder(db)
	require.NoError(t, err)

	query, args, err := qb.NewSelect("u.id", "u.name", "r.name AS role").
		From("users u").
		LeftJoin("roles r ON u.role_id = r.id").
		Where("u.active = ?", true).
		Where("r.name = ?", "admin").
		ToSQL()

	require.NoError(t, err)
	assert.Contains(t, query, "LEFT JOIN roles r ON u.role_id = r.id")
	assert.Contains(t, query, "WHERE u.active =")
	assert.Len(t, args, 2)
}

func TestSelectBuilder_WhereIn(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("error closing db: %v", err)
		}
	}()

	qb, err := NewQueryBuilder(db)
	require.NoError(t, err)

	query, args, err := qb.NewSelect("*").
		From("users").
		WhereIn("id", []int{1, 2, 3}).
		ToSQL()

	require.NoError(t, err)
	// sqlx.In expands the slice
	assert.Len(t, args, 3)
	assert.Equal(t, 1, args[0])
	assert.Equal(t, 2, args[1])
	assert.Equal(t, 3, args[2])
	t.Logf("Generated query: %s", query)
}

func TestSelectBuilder_GroupByHaving(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("error closing db: %v", err)
		}
	}()

	qb, err := NewQueryBuilder(db)
	require.NoError(t, err)

	query, args, err := qb.NewSelect("queue_id", "COUNT(*) as count").
		From("tickets").
		GroupBy("queue_id").
		Having("COUNT(*) > ?", 5).
		OrderBy("count DESC").
		ToSQL()

	require.NoError(t, err)
	assert.Contains(t, query, "GROUP BY queue_id")
	assert.Contains(t, query, "HAVING COUNT(*) >")
	assert.Len(t, args, 1)
	assert.Equal(t, 5, args[0])
}

func TestSelectBuilder_NoTable(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("error closing db: %v", err)
		}
	}()

	qb, err := NewQueryBuilder(db)
	require.NoError(t, err)

	_, _, err = qb.NewSelect("id").ToSQL()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "table not specified")
}

func TestQueryBuilder_Rebind(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("error closing db: %v", err)
		}
	}()

	qb, err := NewQueryBuilder(db)
	require.NoError(t, err)

	// The rebind should convert ? to $ for postgres or keep ? for mysql
	query := "SELECT * FROM users WHERE id = ? AND name = ?"
	rebound := qb.Rebind(query)

	// Result depends on DB_DRIVER env var, but should be valid SQL
	assert.NotEmpty(t, rebound)
	t.Logf("Rebound query: %s", rebound)
}

func TestQueryBuilder_In(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("error closing db: %v", err)
		}
	}()

	qb, err := NewQueryBuilder(db)
	require.NoError(t, err)

	query, args, err := qb.In("SELECT * FROM users WHERE id IN (?)", []int{1, 2, 3, 4, 5})
	require.NoError(t, err)

	assert.Len(t, args, 5)
	t.Logf("Expanded query: %s with args %v", query, args)
}

func TestSelectBuilder_MultipleWheres(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("error closing db: %v", err)
		}
	}()

	qb, err := NewQueryBuilder(db)
	require.NoError(t, err)

	// Simulating the dynamic WHERE clause pattern from admin_customer_users_handlers.go
	sb := qb.NewSelect("cu.id", "cu.login", "cu.email").
		From("customer_user cu").
		LeftJoin("customer_company cc ON cu.customer_id = cc.customer_id")

	// Conditional filters (like the search/validFilter/customerFilter pattern)
	search := "john"
	if search != "" {
		searchPattern := "%" + search + "%"
		sb = sb.Where("(cu.login LIKE ? OR cu.email LIKE ?)", searchPattern, searchPattern)
	}

	validID := 1
	if validID != 0 {
		sb = sb.Where("cu.valid_id = ?", validID)
	}

	sb = sb.OrderBy("cu.last_name", "cu.first_name").Limit(50).Offset(0)

	query, args, err := sb.ToSQL()
	require.NoError(t, err)

	assert.Contains(t, query, "SELECT cu.id, cu.login, cu.email FROM customer_user cu")
	assert.Contains(t, query, "LEFT JOIN customer_company cc")
	assert.Contains(t, query, "WHERE")
	assert.Contains(t, query, "ORDER BY cu.last_name, cu.first_name")
	assert.Len(t, args, 5) // 2 search patterns + validID + limit + offset
	t.Logf("Generated query: %s", query)
	t.Logf("Args: %v", args)
}
