package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Exchange struct {
	Bid string `json:"bid"`
}

const EXCHANGE_URL = "http://localhost:8080/cotacao"
const EXCHANGE_REQUEST_TIMEOUT = 300 * time.Millisecond
const EXPORT_FILENAME = "cotacao.txt"

func main() {
	log.Println("App started")
	ctx := context.Background()

	exchange, err := requestExchange(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = exportExchange(*exchange)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("App finished successfully!")
}

func requestExchange(ctx context.Context) (*Exchange, error) {
	log.Println("Requesting quotation...")

	ctx, cancel := context.WithTimeout(ctx, EXCHANGE_REQUEST_TIMEOUT)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", EXCHANGE_URL, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusGatewayTimeout {
		return nil, errors.New("quotation request timeout")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data Exchange
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	log.Println("Quotation Requested")

	return &data, nil
}

func exportExchange(exchange Exchange) error {
	log.Println("Exporting Quotation...")

	f, err := os.Create(EXPORT_FILENAME)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(f, "Dólar: %s\n", exchange.Bid)
	if err != nil {
		return err
	}

	log.Println("Quotation Exported")

	return nil
}
