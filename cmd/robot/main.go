package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	panic(RunServer())
}

func RunServer() error {
	/*var whCert *os.File
	if os.Getenv("TELEGRAM_WEBHOOK_CERTPATH") != "" {
		whCert, err := os.Open(os.Getenv("TELEGRAM_WEBHOOK_CERTPATH"))
		if err != nil {
			panic(err)
		}
		defer whCert.Close()
	}*/
	bot, err := NewBot("https://api.telegram.org", os.Getenv("TELEGRAM_BOT_TOKEN"), os.Getenv("TELEGRAM_WEBHOOK_URL"))
	if err != nil {
		panic(err)
	}
	g, _ := errgroup.WithContext(context.Background())
	mux := http.NewServeMux()
	mux.HandleFunc("/telegram", bot.WebhookHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("200 Ok"))
	})

	g.Go(func() error {
		fmt.Printf("Starting http web server on :5555")
		httpServer := &http.Server{
			Addr:         ":5555",
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			Handler:      mux,
		}
		return httpServer.ListenAndServe()
	})
	return g.Wait()
}
