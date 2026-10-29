package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Batch query product IDs (Fixes N+1 problem)
func (r *Repository) GetProductsForEvent(productIDs []string) (map[string]Product, error) {
	query := `SELECT id, name, price_cents FROM products WHERE id = ANY($1)`
	rows, err := r.db.Query(query, pq.Array(productIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make(map[string]Product)
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.PriceCents); err != nil {
			return nil, err
		}
		products[p.ID] = p
	}
	return products, nil
}

// Atomic stock decrement directly in SQL (Fixes race conditions)
func (r *Repository) DecrementStockAtomic(machineID, productID string, qty int) (int, error) {
	query := `
		UPDATE machine_stock 
		SET qty = GREATEST(0, qty - $1), updated_at = NOW() 
		WHERE machine_id = $2 AND product_id = $3 
		RETURNING qty;
	`
	var newQty int
	err := r.db.QueryRow(query, qty, machineID, productID).Scan(&newQty)
	if err != nil {
		return 0, fmt.Errorf("failed to atomic decrement stock: %w", err)
	}
	return newQty, nil
}

// Save sale event with idempotency and error logging
func (r *Repository) SaveSale(eventID, machineID, productID string, quantity, priceCents int, soldAt time.Time) error {
	query := `
		INSERT INTO sales (event_id, machine_id, product_id, quantity, price_cents, sold_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (event_id) DO NOTHING;
	`
	_, err := r.db.Exec(query, eventID, machineID, productID, quantity, priceCents, soldAt)
	if err != nil {
		return fmt.Errorf("failed to insert sale record: %w", err)
	}
	return nil
}
