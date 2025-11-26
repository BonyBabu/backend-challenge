package db_test

import (
	"backend-challenge/internal/db"
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	d, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	sqlStmt := `CREATE TABLE IF NOT EXISTS orders (
		id TEXT PRIMARY KEY,
		items BLOB
	);`
	if _, err := d.Exec(sqlStmt); err != nil {
		t.Fatalf("creating table failed: %v", err)
	}
	return d
}

func GetOrder(t *testing.T, orderId db.ID, sqldb *sql.DB) (order db.Order, err error) {
	t.Helper()
	ctx := context.Background()
	query := "SELECT id, items FROM orders WHERE id = ?"
	row := sqldb.QueryRowContext(ctx, query, orderId)
	if row == nil {
		return db.Order{}, db.ErrNotFound(sql.ErrNoRows)
	}
	var id string
	var itemsJSON []byte
	if err := row.Scan(&id, &itemsJSON); err != nil {
		return db.Order{}, err
	}
	var items []db.Item
	if err := json.Unmarshal(itemsJSON, &items); err != nil {
		return db.Order{}, err
	}
	order = db.Order{ID: db.ID(id), Items: items}
	return order, nil
}

func TestGeneralOrder_CreateOrder(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		db *sql.DB
		// Named input parameters for target function.
		order   db.Order
		wantErr bool
	}{
		{
			name: "valid order",
			db:   setupTestDB(t),
			order: db.Order{
				ID:    "order-123",
				Items: []db.Item{{ProductID: "prod-1", Quantity: 2}},
			},
			wantErr: false,
		},
		{
			name: "empty product id",
			db:   setupTestDB(t),
			order: db.Order{
				ID:    "order-123",
				Items: []db.Item{{ProductID: "", Quantity: 2}},
			},
			wantErr: true,
		},
		{
			name: "empty order id",
			db:   setupTestDB(t),
			order: db.Order{
				ID:    "",
				Items: []db.Item{{ProductID: "prod-1", Quantity: 2}},
			},
			wantErr: true,
		},
		{
			name: "invalid quantity",
			db:   setupTestDB(t),
			order: db.Order{
				ID:    "order-124",
				Items: []db.Item{{ProductID: "prod-1", Quantity: 0}},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generalorder := db.NewOrderDao(tt.db)
			gotErr := generalorder.CreateOrder(context.Background(), tt.order)
			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}
			assert.NoError(t, gotErr)
			gotOrder, gotErr := GetOrder(t, tt.order.ID, tt.db)
			assert.NoError(t, gotErr)
			assert.Equal(t, tt.order, gotOrder)
		})
	}
}
