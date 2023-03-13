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
	"github.com/yaq-cc/tmppki"
)

var (
	PORT = os.Getenv("PROXY_PORT")
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

	tpki, err := tmppki.NewTemporaryPKI(tmppki.RSA, tmppki.S128, nil)
	if err != nil {
		logger.Fatal(err)
	}
	
	go func() {
		logger.Printf("Listening and serving HTTP(S) on :%s", PORT)
		err := tpki.ListenAndServeTLS(server)
		if err != nil && err != http.ErrServerClosed {
			logger.Fatal(err)
		}
	}()
	sig := <-signals
	logger.Printf("%s signal received, initiating graceful server shutdown", strings.ToUpper(sig.String()))
	shutCtx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	err = server.Shutdown(shutCtx)
	if err != nil && err != http.ErrServerClosed {
		logger.Fatal(err)
	}
	logger.Println("Server shutdown OK")

}
