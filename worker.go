package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	exchangeRateAPI    = "https://bank.gov.ua/NBUStatService/v1/statdirectory/exchange?json"
	googleAnalyticsURL = "https://www.google-analytics.com/mp/collect"
)

type (
	RateResponse []Rate

	Rate struct {
		R030         int     `json:"r030"`
		Txt          string  `json:"txt"`
		Rate         float64 `json:"rate"`
		Cc           string  `json:"cc"`
		Exchangedate string  `json:"exchangedate"`
	}

	GARequest struct {
		ClientID           string  `json:"client_id"`
		NonPersonalizedAds bool    `json:"non_personalized_ads"`
		Events             []Event `json:"events"`
	}

	EventParams struct {
		Items     []string `json:"items"`
		SessionID uint32   `json:"session_id"`
		Rate      float64  `json:"rate"`
	}

	Event struct {
		Name   string      `json:"name"`
		Params EventParams `json:"params"`
	}
)

func main() {
	var cfg Env
	if err := cfg.Parse(); err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		osCall := <-c
		log.Printf("Stop system call:%+v", osCall)
		cancel()
	}()

	gaURL, err := url.Parse(googleAnalyticsURL)
	if err != nil {
		fmt.Printf("Error %s", err)
		return
	}

	params := url.Values{}
	params.Add("api_secret", cfg.APISecret)
	params.Add("measurement_id", cfg.MeasurementID)
	gaURL.RawQuery = params.Encode()

	httpClient := http.Client{Timeout: time.Second * 5}

	ticker := time.NewTicker(time.Second)
	fmt.Printf("Starting GA rates collector with interval: ", cfg.TickInterval)
	for {
		select {
		case <-ticker.C:
			fmt.Println("Worker woke up, start rates collection:")

			resp, err := httpClient.Get(exchangeRateAPI)
			if err != nil {
				fmt.Printf("Error %s", err)
				return
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error %s", err)
				return
			}

			var response RateResponse
			if err := json.Unmarshal(body, &response); err != nil {
				fmt.Printf("Error %s", err)
				return
			}
			resp.Body.Close()

			var uahUsdRate *Rate
			for i := range response {
				if response[i].Txt == "Долар США" {
					uahUsdRate = &response[i]
					break
				}
			}

			fmt.Println("Found new UAH/USD rate: ", uahUsdRate.Rate)
			if uahUsdRate == nil {
				fmt.Printf("UAH/USD rate not found in response")
				return
			}

			gaEventPayload := GARequest{
				// used already created ID from cookies
				ClientID:           "93883486.1698489453",
				NonPersonalizedAds: true,
				Events: []Event{
					{
						Name: "uan_usd",
						Params: EventParams{
							Items: []string{},
							Rate:  uahUsdRate.Rate,
							// used already created session from cookies
							SessionID: 1698489452,
						},
					},
				},
			}

			rawPayload, err := json.Marshal(gaEventPayload)
			if err != nil {
				fmt.Printf("Error %s", err)
				return
			}

			resp, err = httpClient.Post(gaURL.String(), "application/json", bytes.NewBuffer(rawPayload))
			if err != nil {
				log.Fatal(err)
			}

			if resp.StatusCode != http.StatusNoContent {
				fmt.Printf("Failed to sent event response code is not OK %d", resp.StatusCode)
			}
		case <-ctx.Done():
			fmt.Println("Received signal to stop, exiting....")
			return
		}
	}
}
