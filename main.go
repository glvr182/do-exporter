package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/digitalocean/godo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	current_bill = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "do_current_bill",
			Help: "Current bill usage on digitalocean",
		}, nil,
	)
	billing_history = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "do_billing_history",
			Help: "Past bills on digitalocean",
		}, []string{"month", "year"},
	)
)

func init() {
	prometheus.MustRegister(current_bill, billing_history)
}

func main() {
	log.Printf("Starting exporter")
	key, found := os.LookupEnv("DO_KEY")
	if !found {
		log.Fatalln("Please provide an API Key in the environment (DO_KEY)")
	}
	client := godo.NewFromToken(key)
	go func() {
		for {
			if err := update(client); err != nil {
				log.Fatal(err)
			}
			time.Sleep(time.Minute)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func update(client *godo.Client) error {
	balance, _, err := client.Balance.Get(context.Background())
	if err != nil {
		return err
	}

	amount, err := strconv.ParseFloat(balance.MonthToDateUsage, 64)
	if err != nil {
		return err
	}

	current_bill.With(prometheus.Labels{}).Set(amount)

	bills, _, err := client.BillingHistory.List(context.Background(), nil)
	if err != nil {
		return err
	}

	for _, bill := range bills.BillingHistory {
		if bill.Type == "Payment" {
			continue
		}
		amount, err := strconv.ParseFloat(bill.Amount, 64)
		if err != nil {
			return err
		}
		billing_history.With(prometheus.Labels{"year": strconv.Itoa(bill.Date.Year()), "month": bill.Date.Month().String()}).Set(amount)
	}

	return nil
}
