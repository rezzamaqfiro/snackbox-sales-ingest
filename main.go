package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr == "" {
		dbConnStr = "postgres://snackbox:Sn4ckb0x!dev@localhost:5432/snackbox?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	cache := NewStockCache()

	// Initialize fixed worker pool (Worker Pool Pattern)
	eventsChan := make(chan SaleEvent, 1000)
	handler := NewHandler(repo, cache, eventsChan)

	const numWorkers = 5
	for i := 0; i < numWorkers; i++ {
		go handler.StartWorker()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/sales", handler.HandleSales)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("Server running on port :8080...")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
