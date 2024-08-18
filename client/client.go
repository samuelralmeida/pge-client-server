package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		log.Println("error opening file")
	}
	defer file.Close()

	ctx := context.Background()
	exchangeData, err := requestExchangeData(ctx)
	if err != nil {
		log.Println("error requesting exchange data:", err)
		return
	}

	file.WriteString(fmt.Sprintf("Dólar: %s", exchangeData.Bid))
}

type exchangeData struct {
	Bid string `json:"bid"`
}

func requestExchangeData(ctx context.Context) (*exchangeData, error) {
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		return nil, fmt.Errorf("error to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

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
