package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexeylcp/lucx-core/internal/config"
)

func main() {
	cfg := config.Parse()
	log.Printf("LucX Core starting on %s", cfg.ListenAddr)
	// TODO: init store, init SSH pool, init API router, start server

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("LucX Core shutting down")
}
