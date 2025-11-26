package db

import (
	"context"
	"database/sql"
)

// product_dao.go is intentionally minimal for now to avoid parse errors during build.
// TODO: Implement Product DAO methods.

type ProductDaoImpl struct {
	db *sql.DB
}

// NewProductDao creates a GeneralProduct with the provided db instance.
func NewProductDao(db *sql.DB) ProductDao {
	return &ProductDaoImpl{db: db}
}

// GetAllProducts implements ProductDao.
func (g *ProductDaoImpl) GetAllProducts(ctx context.Context) (products []Product, err error) {
	rows, err := g.db.QueryContext(ctx, "SELECT id, name, price, category FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.Id, &p.Name, &p.Price, &p.Category); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

// GetProduct implements ProductDao.
func (g *ProductDaoImpl) GetProduct(ctx context.Context, id ID) (Product, error) {
	row := g.db.QueryRowContext(ctx, "SELECT id, name, price, category FROM products WHERE id = ?", id)
	if row == nil {
		return Product{}, ErrNotFound(sql.ErrNoRows)
	}
	var p Product
	if err := row.Scan(&p.Id, &p.Name, &p.Price, &p.Category); err != nil {
		return Product{}, err
	}
	return p, nil
}

var _ ProductDao = &ProductDaoImpl{}
