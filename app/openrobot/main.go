package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var log = logrus.New()

func WithLogging(h http.Handler) http.Handler {
	loggingFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		duration := time.Since(start)

		logrus.WithFields(logrus.Fields{
			"uri":      r.RequestURI,
			"method":   r.Method,
			"duration": duration,
		}).Info("request completed")
	}
	return http.HandlerFunc(loggingFn)
}

func newDB() (*sqlx.DB, error) {
	var (
		con *sqlx.DB
		err error
	)
	//dsn := "root:ragnarok@(localhost:9999)/rathena"
	if con, err = sqlx.Open("mysql", os.Getenv("DSN")); err != nil {
		return nil, fmt.Errorf("open connection to DB failed %v", err)
	}
	return con, nil
}

func main() {
	log.Out = os.Stdout
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{})
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
	bot := NewBot("https://api.telegram.org", os.Getenv("TELEGRAM_BOT_TOKEN"), os.Getenv("TELEGRAM_WEBHOOK_URL"))
	g, _ := errgroup.WithContext(context.Background())
	mux := http.NewServeMux()
	mux.HandleFunc("/telegram-webhook", bot.WebhookHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("200 Ok"))
	})

	g.Go(func() error {
		httpServer := &http.Server{
			Addr:         ":5555",
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			Handler:      WithLogging(mux),
		}
		log.Info("Started server on :5555")
		return httpServer.ListenAndServe()
	})
	return g.Wait()
}
