package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aiviaio/go-binance/v2"
	"github.com/aiviaio/go-binance/v2/futures"
)

// getSymbols retrieves a specified number of symbols from the Binance exchange.
func getSymbols(client *futures.Client, num int) ([]string, error) {
	exchangeInfo, err := client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		log.Fatalf("Error fetching exchange info: %v", err)
		return nil, err
	}

	var symbols []string
	for i := 0; i < num; i++ {
		symbols = append(symbols, exchangeInfo.Symbols[i].Symbol)
	}

	return symbols, nil
}

// fetchPrices concurrently retrieves prices for a list of symbols from the Binance exchange.
func fetchPrices(client *futures.Client, symbols []string, ch chan map[string]string) {
	var wg sync.WaitGroup
	wg.Add(len(symbols))

	for _, symbol := range symbols {
		go func(sym string) {
			defer wg.Done()
			ticker, err := client.NewListPricesService().Symbol(sym).Do(context.Background())
			if err != nil {
				log.Printf("Error fetching price for %s: %v", sym, err)
				return
			}
			ch <- map[string]string{sym: ticker[0].Price}
		}(symbol)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()
}

// printPrices prints out the symbol and corresponding price received from the channel.
func printPrices(ch chan map[string]string) {
	for data := range ch {
		for sym, price := range data {
			fmt.Printf("%s %s\n", sym, price)
		}
	}
}

func main() {
	client := binance.NewFuturesClient("", "")

	for {
		fmt.Println("Getting actual prices...")

		symbols, err := getSymbols(client, 5)
		if err != nil {
			fmt.Println("Error: ", err)
			fmt.Println("Application closing...")
			return
		}

		ch := make(chan map[string]string)

		go fetchPrices(client, symbols, ch)
		printPrices(ch)

		// Pause for 10 seconds before fetching prices again
		time.Sleep(10 * time.Second)
	}
}
