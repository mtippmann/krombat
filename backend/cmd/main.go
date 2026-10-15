package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pnz1990/krombat/backend/internal/handlers"
	"github.com/pnz1990/krombat/backend/internal/k8s"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := run(); err != nil {
		slog.Error("backend stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	secret := os.Getenv("SESSION_SECRET")
	if err := handlers.ValidateSessionSecret(secret); err != nil {
		return err
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return errors.New("PORT must be an integer between 1 and 65535")
	}
	client, err := k8s.NewClient()
	if err != nil {
		return fmt.Errorf("create Kubernetes client: %w", err)
	}
	api, err := handlers.New(client.Dynamic, secret)
	if err != nil {
		return fmt.Errorf("initialize API: %w", err)
	}
	server := &http.Server{
		Addr: ":" + port, Handler: api.Routes(),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
		WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second,
	}
	slog.Info("local demo backend starting", "addr", server.Addr, "namespace", k8s.GameNamespace)
	return server.ListenAndServe()
}
