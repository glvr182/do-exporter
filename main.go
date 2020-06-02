package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/digitalocean/godo"
)

func main() {
	log.Printf("Starting exporter")
	key, found := os.LookupEnv("DO_KEY")
	if !found {
		log.Fatalln("Please provide an API Key in the environment (DO_KEY)")
	}
	client := godo.NewFromToken(key)
	for {
		update(client)
		time.Sleep(time.Minute)
	}
}

func update(client *godo.Client) error {
	bills, resp, err := client.BillingHistory.List(context.Background(), nil)
	if err != nil {
		log.Fatalln(err)
	}

	for _, bill := range bills.BillingHistory {
		if bill.Type == "Payment" {
			continue
		}
		log.Printf("%v: %v\n", bill.Date.Format("02-01-2006"), bill.Amount)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Remote returned status: %d\n", resp.StatusCode)
	}

	balance, resp, err := client.Balance.Get(context.Background())
	log.Printf("Current bill: %v\n", balance.MonthToDateUsage)
	return nil
}
