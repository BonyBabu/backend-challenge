package db_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/go-playground/validator/v10"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"

	"backend-challenge/internal/db"
)

func setupProductTestDB(t *testing.T) *sql.DB {
	t.Helper()
	d, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	createStmt := `CREATE TABLE IF NOT EXISTS products (
		id TEXT PRIMARY KEY,
		name TEXT,
		price INTEGER,
		category TEXT
	);`
	if _, err := d.Exec(createStmt); err != nil {
		t.Fatalf("creating table failed: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return d
}

func CreateProducts(t *testing.T, products []db.Product, db *sql.DB) error {
	for _, p := range products {
		// Validate order structure
		validate := validator.New()
		if err := validate.Struct(p); err != nil {
			return err
		}
		_, err := db.ExecContext(context.Background(), "INSERT INTO products (id, name, price, category) VALUES (?, ?, ?, ?)", p.Id, p.Name, p.Price, p.Category)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestGeneralProduct_GetAllProducts(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for
		// et function.
		products []db.Product
		wantErr  bool
	}{
		{
			name: "two products in db",
			products: []db.Product{
				{Id: "prod-1", Name: "Product 1", Price: 100, Category: "cat"},
				{Id: "prod-2", Name: "Product 2", Price: 200, Category: "cat"},
			},
			wantErr: false,
		},
		{
			name:     "two products in db",
			products: []db.Product{},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sqlDB := setupProductTestDB(t)
			g := db.NewProductDao(sqlDB)
			gotErr := CreateProducts(t, tt.products, sqlDB)
			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}
			assert.NoError(t, gotErr)
			dbProducts, err := g.GetAllProducts(context.Background())
			assert.NoError(t, err)
			// Verify all inserted products are present
			productMap := make(map[string]db.Product)
			for _, p := range dbProducts {
				productMap[p.Id] = p
			}
			for _, want := range tt.products {
				got, exists := productMap[want.Id]
				assert.True(t, exists, "product ID %s not found in DB", want.Id)
				assert.Equal(t, want, got)
			}
		})
	}
}

func TestGeneralProduct_GetProduct(t *testing.T) {
	sqlDB := setupProductTestDB(t)
	sqlDB.ExecContext(context.Background(), "INSERT INTO products (id, name, price, category) VALUES (?, ?, ?, ?)",
		"1", "chicken waffle", 120, "Waffles")
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		id      db.ID
		want    db.Product
		wantErr bool
	}{
		{
			name:    "existing product",
			id:      "1",
			want:    db.Product{Id: "1", Name: "chicken waffle", Price: 120, Category: "Waffles"},
			wantErr: false,
		},
		{
			name:    "non-existing product",
			id:      "non-existent",
			want:    db.Product{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := db.NewProductDao(sqlDB)
			got, gotErr := g.GetProduct(context.Background(), tt.id)
			assert.Equal(t, tt.wantErr, gotErr != nil)
			if !tt.wantErr {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
