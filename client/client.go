package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	serverURL string = "http://127.0.0.1:8080/cotacao"
)

type ServerAPIResponse struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	dollar, err := getCotacao(ctx)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println(dollar)

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
