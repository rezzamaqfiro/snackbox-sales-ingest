package main

import "time"

type SaleItem struct {
	ProductID  string `json:"product_id"`
	Quantity   int    `json:"quantity"`
	PriceCents int    `json:"price_cents"`
}

type SaleEvent struct {
	EventID   string     `json:"event_id"`
	MachineID string     `json:"machine_id"`
	SoldAt    time.Time  `json:"sold_at"`
	Items     []SaleItem `json:"items"`
}

type Product struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PriceCents int    `json:"price_cents"`
}
