package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	serverURL   string = "http://127.0.0.1:8080/cotacao"
	cotacaoPath string = "cotacao.txt"
)

type ServerAPIResponse struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	resp, err := getCotacao(ctx)
	if err != nil {
		log.Panic(err)
	}

	dollar, err := strconv.ParseFloat(resp.Bid, 64)
	if err != nil {
		log.Panic(err)
	}

	err = saveDollarExchange(dollar)
	if err != nil {
		log.Panic(err)
	}

}

func getCotacao(ctx context.Context) (*ServerAPIResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", serverURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var dollar ServerAPIResponse

	err = json.Unmarshal(body, &dollar)
	if err != nil {
		return nil, err
	}

	return &dollar, err
}

func saveDollarExchange(dollar float64) error {

	_, err := os.Stat(cotacaoPath)
	if os.IsNotExist(err) {
		file, err := os.Create(cotacaoPath)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	file, err := os.OpenFile(cotacaoPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	data := fmt.Sprintf("Dólar: %v\n", dollar)

	_, err = file.WriteString(data)
	if err != nil {
		return err
	}

	return nil
}
