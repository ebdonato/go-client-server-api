package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"go-client-server-api/server/constants"
	"go-client-server-api/server/database"
	"go-client-server-api/server/models"
)

func main() {
	log.Println("App started")

	db := database.OpenDatabase()
	defer db.Close()

	http.HandleFunc("/health", handlerHealth)
	http.HandleFunc("/quotation", handlerQuotation)
	http.ListenAndServe(":8080", nil)
}

func handlerHealth(w http.ResponseWriter, r *http.Request) {
	defer log.Println("App health checked")
	w.Write([]byte("Server is running"))
}

func handlerQuotation(w http.ResponseWriter, r *http.Request) {
	log.Println("Request quotation started")
	defer log.Println("Request quotation ended")

	ctx := r.Context()

	response, err := requestExchange(ctx)
	if err != nil {
		log.Println("ERROR: External API Request Timeout")
		w.WriteHeader(http.StatusGatewayTimeout)
		return
	}

	err = persistExchange(ctx, database.GetDatabase(), response)
	if err != nil {
		log.Println("ERROR: Persist Timeout")
		w.WriteHeader(http.StatusGatewayTimeout)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func requestExchange(ctx context.Context) (*models.Exchange, error) {
	ctx, cancel := context.WithTimeout(ctx, constants.EXCHANGE_REQUEST_TIMEOUT)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", constants.EXCHANGE_URL, nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var data models.USDBRL
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}

	return &data.USDBRL, nil
}

func persistExchange(ctx context.Context, db *sql.DB, e *models.Exchange) error {
	ctx, cancel := context.WithTimeout(ctx, constants.EXCHANGE_PERSIST_TIMEOUT)
	defer cancel()

	stmt, err := db.PrepareContext(ctx, "INSERT INTO exchanges(code, code_in, name, high, low, var_bid, pct_change, bid, ask, timestamp, create_date) VALUES(?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		panic(err)
	}

	_, err = stmt.Exec(e.Code, e.CodeIn, e.Name, e.High, e.Low, e.VarBid, e.PctChange, e.Bid, e.Ask, e.Timestamp, e.CreateDate)
	if err != nil {
		return err
	}

	return nil
}
