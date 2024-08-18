package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sqliteConnection()
	if err != nil {
		log.Fatal("error connecting database:", err)
	}

	http.HandleFunc("/cotacao", handler(db))

	log.Println("Serving on port :8080")
	http.ListenAndServe(":8080", nil)
}

func sqliteConnection() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./exchange_rates.db")
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS exchange_rates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bid REAL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return nil, fmt.Errorf("error creating table database: %w", err)
	}

	return db, nil
}

func handler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		exchangeData, err := getExchangeRate(ctx)
		if err != nil {
			log.Println("error to get exchange data:", err)
			http.Error(w, "error to get exchange data", http.StatusInternalServerError)
			return
		}

		err = saveExchangeRate(ctx, db, exchangeData.USDBRL.Bid)
		if err != nil {
			log.Println("error saving exchange rate:", err)
			http.Error(w, "error to get exchange data", http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(exchangeData.USDBRL)
		if err != nil {
			log.Println("error to encode response:", err)
			http.Error(w, "error to encode response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
	}
}

type exchangeData struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func getExchangeRate(ctx context.Context) (*exchangeData, error) {
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, fmt.Errorf("error to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error to do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error to read body: %w", err)
	}

	var exchangeData exchangeData

	err = json.Unmarshal(body, &exchangeData)
	if err != nil {
		return nil, fmt.Errorf("error to Unmarshal body: %w", err)
	}

	return &exchangeData, nil
}

func saveExchangeRate(ctx context.Context, db *sql.DB, bid string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	query := "INSERT INTO exchange_rates (bid) VALUES (?)"
	_, err := db.ExecContext(ctx, query, bid)
	if err != nil {
		return fmt.Errorf("error saving bid: %w", err)
	}

	return nil
}
