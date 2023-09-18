package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const EXCHANGE_URL = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
const EXCHANGE_REQUEST_TIMEOUT = 200 * time.Millisecond
const EXCHANGE_PERSIST_TIMEOUT = 10 * time.Millisecond

type USDBRL struct {
	USDBRL Exchange `json:"USDBRL"`
}

type Exchange struct {
	Code       string `json:"code"`
	CodeIn     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

var db *sql.DB

func main() {
	log.Println("App started")

	OpenDatabase()
	defer db.Close()

	http.HandleFunc("/health", handlerHealth)
	http.HandleFunc("/quotation", handlerQuotation)
	http.HandleFunc("/cotacao", handlerQuotation)
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

	err = persistExchange(ctx, db, response)
	if err != nil {
		log.Println("ERROR: Persist Timeout")
		w.WriteHeader(http.StatusGatewayTimeout)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func OpenDatabase() {
	var err error

	db, err = sql.Open("sqlite3", "./db.sqlite")
	if err != nil {
		panic(err)
	}

	stmt := `
        CREATE TABLE IF NOT EXISTS exchanges(
            id INTEGER PRIMARY KEY,
            code TEXT,
            code_in TEXT,
            name TEXT,
            high TEXT,
            low TEXT,
            var_bid TEXT,
            pct_change TEXT,
            bid TEXT,
            ask TEXT,
            timestamp TEXT,
            create_date TEXT,
			persist_date DATETIME DEFAULT CURRENT_TIMESTAMP
        );
    `
	_, err = db.Exec(stmt)

	if err != nil {
		panic(err)
	}
}

func requestExchange(ctx context.Context) (*Exchange, error) {
	ctx, cancel := context.WithTimeout(ctx, EXCHANGE_REQUEST_TIMEOUT)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", EXCHANGE_URL, nil)
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

	var data USDBRL
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}

	return &data.USDBRL, nil
}

func persistExchange(ctx context.Context, db *sql.DB, e *Exchange) error {
	ctx, cancel := context.WithTimeout(ctx, EXCHANGE_PERSIST_TIMEOUT)
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
