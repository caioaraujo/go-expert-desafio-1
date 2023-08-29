package main

import (
	"context"
	"encoding/json"
	"errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"time"
)

type CotacaoDolarReal struct {
	USDBRL `json:"USDBRL"`
}

type USDBRL struct {
	ID         int    `gorm:"primaryKey"`
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
}

func main() {
	http.HandleFunc("/cotacao", CotacaoHandler)
	http.ListenAndServe(":8080", nil)
}

func CotacaoHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	cotacao, err := Cotacao()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = GravarNoBancoDeDados(cotacao)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotacao.USDBRL.Bid)

}

func Cotacao() (*CotacaoDolarReal, error) {
	log.Println("Request iniciada.")
	defer log.Println("Request finalizada")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)

	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			log.Println("Tempo limite atingido ao realizar consulta")
			return nil, err
		default:
			return nil, err
		}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var data CotacaoDolarReal
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil

}

func GravarNoBancoDeDados(data *CotacaoDolarReal) error {
	log.Println("Salvando dados na base...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	db, err := gorm.Open(sqlite.Open("cotacao.db"), &gorm.Config{})
	if err != nil {
		return err
	}
	db.AutoMigrate(&USDBRL{})
	db.WithContext(ctx).Create(&data.USDBRL)

	select {
	case <-ctx.Done():
		log.Println("Tempo limite atingido ao realizar inserção na base de dados")
		return errors.New("timeout reached")
	default:
		log.Println("Dados salvos com sucesso")
		return nil
	}

	// select all
	//var cotacaoDolar []USDBRL
	//db.Find(&cotacaoDolar)
	//for _, c := range cotacaoDolar {
	//	fmt.Println(c)
	//}
}
