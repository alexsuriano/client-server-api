package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

const (
	urlDollarExchange string = `https://economia.awesomeapi.com.br/json/last/USD-BRL`
	dbPath            string = `sqlite-database.db`
)

type DollarExchangeAPIResponse struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type ServerAPIResponse struct {
	Bid string `json:"bid"`
}

func init() {
	err := DBCreate(dbPath)
	if err != nil {
		log.Panic(err)
	}

	db, err := DBConnect(dbPath)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	err = CreateTable(db)
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	http.HandleFunc("/cotacao", HandleDollarExchange)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println(err)
	}
}

func getDollarExchange(ctx context.Context, url string) (*DollarExchangeAPIResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var dollar DollarExchangeAPIResponse

	err = json.Unmarshal(body, &dollar)
	if err != nil {
		return nil, err
	}

	return &dollar, nil
}

func HandleDollarExchange(w http.ResponseWriter, r *http.Request) {
	ctxRequest, cancelRequest := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancelRequest()

	ctxDBWrite, cancelDBWrite := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancelDBWrite()

	dollar, err := getDollarExchange(ctxRequest, urlDollarExchange)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Solicitação de cotação indiponível"))
		log.Println(err)
	} else {

		db, err := DBConnect(dbPath)
		if err != nil {
			log.Panic(err)
		}
		defer db.Close()

		err = InsertDollarQuote(ctxDBWrite, db, dollar)
		if err != nil {
			log.Println(err)
		}

		var serverResponse ServerAPIResponse
		serverResponse.Bid = dollar.Usdbrl.Bid

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(serverResponse)
	}
}

func DBConnect(path string) (*sql.DB, error) {
	DBCreate(path)

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func DBCreate(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		file, err := os.Create(path)
		defer file.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS dollar_exchanges(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		code VARCHAR(50) NOT NULL,
		code_in VARCHAR(50) NOT NULL,
		name VARCHAR(255) NOT NULL,
		high DECIMAL(9, 4) NOT NULL,
		low DECIMAL(9, 4) NOT NULL,
		var_bid DECIMAL (9, 4) NOT NULL,
		pct_change DECIMAL (9, 4) NOT NULL,
		bid DECIMAL (9, 4) NOT NULL,
		ask DECIMAL (9, 4) NOT NULL,
		timestamp VARCHAR NOT NULL,
		create_date DATETIME NOT NULL
	);`

	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func InsertDollarQuote(ctx context.Context, db *sql.DB, dollar *DollarExchangeAPIResponse) error {
	query := `insert into dollar_exchanges(code, code_in, name, high, low, var_bid, pct_change, 
		bid, ask, timestamp, create_date) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		dollar.Usdbrl.Code,
		dollar.Usdbrl.Codein,
		dollar.Usdbrl.Name,
		dollar.Usdbrl.High,
		dollar.Usdbrl.Low,
		dollar.Usdbrl.VarBid,
		dollar.Usdbrl.PctChange,
		dollar.Usdbrl.Bid,
		dollar.Usdbrl.Ask,
		dollar.Usdbrl.Timestamp,
		dollar.Usdbrl.CreateDate,
	)
	if err != nil {
		return err
	}

	return nil
}
