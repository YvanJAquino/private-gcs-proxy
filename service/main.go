package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"cloud.google.com/go/storage"
	"cloud.google.com/go/compute/metadata"
)

var (
	PORT = os.Getenv("PROXY_PORT")
	USE_TLS = os.Getenv("PROXY_USE_TLS")
	ProjectID = Must(metadata.ProjectID())
)

func main() {
	parent := context.Background()
	logger := log.Default()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	client, err := storage.NewClient(parent)
	if err != nil {
		logger.Fatal(err)
	}
	gcsProxy := New(client, logger)
	server := &http.Server{
		Addr:        ":" + PORT,
		Handler:     gcsProxy,
		BaseContext: func(l net.Listener) context.Context { return parent },
	}

	if USE_TLS != "" {
		SERVER_CRT := os.Getenv("SERVER_CRT")
		SERVER_KEY := os.Getenv("SERVER_KEY")
		if SERVER_CRT == "" || SERVER_KEY == "" {
			logger.Fatal("SERVER_CRT and SERVER_KEY must be set")
		}
		go func() {
			logger.Printf("Listening and serving HTTP(S) on :%s", PORT)
			err := server.ListenAndServeTLS(SERVER_CRT, SERVER_KEY)
			if err != nil && err != http.ErrServerClosed {
				logger.Fatal(err)
			}
		}()

	} else {
		go func() {
			logger.Printf("Listening and serving HTTP on :%s", PORT)
			err := server.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				logger.Fatal(err)
			}
		}()
	}
	
	sig := <-signals
	logger.Printf("%s signal received, initiating graceful shutdown", strings.ToUpper(sig.String()))
	shutCtx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	err = server.Shutdown(shutCtx)
	if err != nil && err != http.ErrServerClosed {
		logger.Fatal(err)
	}
	logger.Println("Server shutdown OK")

}
