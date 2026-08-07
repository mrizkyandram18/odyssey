package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"odyssey/pkg/observability"
	"odyssey/pkg/server"
)

func main() {
	_ = godotenv.Load("../../.env")

	srvHandler, err := server.BuildHandler()
	if err != nil {
		log.Fatalf("Server build error: %v", err)
	}

	srv := &http.Server{
		Addr:              ":" + getPort(),
		Handler:           srvHandler.Handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	port := getPort()
	log.Printf("Odyssey server starting on :%s", port)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sm := observability.NewShutdownManager(srv, 30*time.Second)
	sm.AddHook("server_cleanup", func(ctx context.Context) error {
		return srvHandler.Cleanup(ctx)
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-sigCh
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := sm.Shutdown(ctx); err != nil {
		log.Fatalf("Graceful shutdown error: %v", err)
	}
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}
