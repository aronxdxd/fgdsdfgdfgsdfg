package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/cors"

	"project/funcs"
	"project/funcs/modifies"
	"project/funcs/recovery"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error running server: %v", err)
	}
}

func run() error {
	log.Println("Connected to MongoDB")

	mux := http.NewServeMux()
	mux.HandleFunc("/get_info", funcs.HandleInfo)
	mux.HandleFunc("/check_subscription", funcs.CheckSubscription)
	mux.HandleFunc("/clicker", funcs.HandleClicker)
	mux.HandleFunc("/buy_toques", modifies.HandleBuyToques)
	mux.HandleFunc("/top", funcs.HandleTop)

	// Create a new CORS handler
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Allow all origins
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		MaxAge:         300, // Maximum age for the preflight request
	})

	// Wrap our existing mux with the CORS handler
	handler := c.Handler(mux)

	server := &http.Server{
		Addr:    ":80",
		Handler: handler,
	}

	go recovery.StartEnergyRecovery()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Server starting on :80")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	<-done
	log.Println("Server stopping...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return err
	}

	log.Println("Server stopped gracefully")
	return nil
}
