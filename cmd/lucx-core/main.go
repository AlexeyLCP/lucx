package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexeylcp/lucx-core/internal/api"
	"github.com/alexeylcp/lucx-core/internal/chain"
	"github.com/alexeylcp/lucx-core/internal/config"
	"github.com/alexeylcp/lucx-core/internal/ssh"
	"github.com/alexeylcp/lucx-core/internal/store"
)

func main() {
	cfg := config.Parse()

	s, err := store.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer s.Close()

	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "lucx-dev-secret-change-me"
		log.Println("WARNING: using default JWT secret")
	}

	sshFactory := ssh.NewFactory(s)
	engine := chain.NewEngine(s, func(serverID string) (*ssh.Client, error) {
		srv, err := s.GetServer(serverID)
		if err != nil {
			return nil, err
		}
		return sshFactory.Dial(srv)
	})

	handlers := &api.Handlers{
		Store:      s,
		Engine:     engine,
		SSHFactory: func(srv *store.Server) (*ssh.Client, error) { return sshFactory.Dial(srv) },
		JWTSecret:  cfg.JWTSecret,
	}

	router := api.NewRouter(handlers)
	log.Printf("LucX Core listening on %s", cfg.ListenAddr)
	go func() {
		if err := http.ListenAndServe(cfg.ListenAddr, router); err != nil {
			log.Fatalf("http: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("LucX Core shutting down")
}
