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
	whCert, err := os.Open("dev.crt")
	if err != nil {
		panic(err)
	}
	defer whCert.Close()
	bot, err := NewBot("https://api.telegram.org", os.Getenv("TELEGRAM_BOT_TOKEN"), "https://robot.open-ro.com/telegram", whCert)
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
		fmt.Printf("Starting http web server on :8443")
		httpServer := &http.Server{
			Addr:         ":8443",
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			Handler:      mux,
		}
		return httpServer.ListenAndServeTLS("dev.crt", "dev.key")
	})
	return g.Wait()
}
