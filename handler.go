package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Handler struct {
	repo   *Repository
	cache  *StockCache
	events chan SaleEvent
}

func NewHandler(repo *Repository, cache *StockCache, events chan SaleEvent) *Handler {
	return &Handler{
		repo:   repo,
		cache:  cache,
		events: events,
	}
}

func (h *Handler) HandleSales(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var ev SaleEvent
	if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Queue event to background worker pool
	select {
	case h.events <- ev:
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "queued"})
	default:
		http.Error(w, "Ingestion buffer full", http.StatusServiceUnavailable)
	}
}

func (h *Handler) StartWorker() {
	for ev := range h.events {
		if err := h.processEvent(ev); err != nil {
			log.Printf("Error processing event %s: %v", ev.EventID, err)
		}
	}
}

func (h *Handler) processEvent(ev SaleEvent) error {
	for _, item := range ev.Items {
		newQty, err := h.repo.DecrementStockAtomic(ev.MachineID, item.ProductID, item.Quantity)
		if err != nil {
			return err
		}

		cacheKey := ev.MachineID + ":" + item.ProductID
		h.cache.Set(cacheKey, newQty)

		if err := h.repo.SaveSale(ev.EventID, ev.MachineID, item.ProductID, item.Quantity, item.PriceCents, ev.SoldAt); err != nil {
			log.Printf("Failed to record sale audit for event %s: %v", ev.EventID, err)
			return err
		}
	}
	return nil
}
