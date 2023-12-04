package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Cambio struct {
	Usdbrl Usdbrl `json:"USDBRL"`
}

type Usdbrl struct {
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

type Cotacao struct {
	Valor string `json:"valor"`
}

func UnmarshalCambio(data []byte) (Cambio, error) {
	var r Cambio
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Cambio) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func (c *Cotacao) Marshal() ([]byte, error) {
	return json.Marshal(c)
}

func main() {
	http.HandleFunc("/cotacao", handlerCotacao)
	http.ListenAndServe(":8080", nil)
}

func handlerCotacao(w http.ResponseWriter, r *http.Request) {
	log.Println("Request iniciada")
	defer log.Println("Request finalizada")

	cambio, err := getCambio()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Valor do dolar é: %s\n", cambio.Usdbrl.Bid)

	err = saveCambio(cambio)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Cambio salvo com sucesso")

	cotacao := Cotacao{Valor: cambio.Usdbrl.Bid}
	jsonResponse, err := cotacao.Marshal()
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func saveCambio(cambio *Cambio) error {
	db, err := sql.Open("sqlite3", "./cambio.db")
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(10*time.Millisecond))
	defer cancel()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS cambio (id INTEGER PRIMARY KEY AUTOINCREMENT, code TEXT, codein TEXT, name TEXT, high TEXT, low TEXT, varBid TEXT, pctChange TEXT, bid TEXT, ask TEXT, timestamp TEXT, createDate TEXT)")
	if err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO cambio (code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, createDate) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, cambio.Usdbrl.Code, cambio.Usdbrl.Codein, cambio.Usdbrl.Name, cambio.Usdbrl.High, cambio.Usdbrl.Low, cambio.Usdbrl.VarBid, cambio.Usdbrl.PctChange, cambio.Usdbrl.Bid, cambio.Usdbrl.Ask, cambio.Usdbrl.Timestamp, cambio.Usdbrl.CreateDate)
	if err != nil {
		return err
	}

	return nil
}

func getCambio() (*Cambio, error) {
	ctx := context.Background()
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(200*time.Millisecond))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	cambio, err := UnmarshalCambio(body)
	if err != nil {
		return nil, err
	}

	return &cambio, nil
}
