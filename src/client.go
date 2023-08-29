package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			log.Println("Tempo limite atingido no client")
		default:
			panic(err)
		}
		return
	}
	defer res.Body.Close()
	resp, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	file, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	_, err = file.WriteString(fmt.Sprintf("Dólar: %s", string(resp)))
	if err != nil {
		panic(err)
	}
	fmt.Println("Arquivo criado com sucesso!")
}
