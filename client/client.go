package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	Valor string `json:"valor"`
}

func UnmarshalCotacao(data []byte) (Cotacao, error) {
	var r Cotacao
	err := json.Unmarshal(data, &r)
	return r, err
}

func (c *Cotacao) Marshal() ([]byte, error) {
	return json.Marshal(c)
}

func main() {

	cotacao, err := getCotacao()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = writeCotacao(cotacao)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("A cotação recebida e salva foi de R$ %s\n", cotacao.Valor)

}

func writeCotacao(cotacao *Cotacao) error {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(fmt.Sprintf("Dólar: %s", cotacao.Valor))
	if err != nil {
		return err
	}
	return nil
}

func getCotacao() (*Cotacao, error) {
	ctx := context.Background()
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(300*time.Millisecond))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
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

	cotacao, err := UnmarshalCotacao(body)
	if err != nil {
		return nil, err
	}

	return &cotacao, nil
}
