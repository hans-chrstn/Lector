package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/user/lector/internal/extensions"
	"github.com/user/lector/internal/httpapi"
	"github.com/user/lector/internal/storage"
)

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func run() error {
	if len(os.Args) > 1 && os.Args[1] == "plugin-worker" {
		return extensions.Worker(os.Stdin, os.Stdout)
	}
	if len(os.Args) > 1 {
		return errors.New("usage: lector")
	}
	address := env("LISTEN_ADDR", "127.0.0.1:3000")
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	user, password := os.Getenv("AUTH_USER"), os.Getenv("AUTH_PASSWORD")
	if (user == "") != (password == "") {
		return errors.New("set both AUTH_USER and AUTH_PASSWORD")
	}
	ip := net.ParseIP(host)
	if host != "localhost" && (ip == nil || !ip.IsLoopback()) && user == "" {
		return errors.New("authentication is required when listening beyond localhost")
	}
	size, err := strconv.ParseInt(env("MAX_UPLOAD_SIZE", "100"), 10, 64)
	if err != nil || size < 1 || size > 1024 {
		return errors.New("MAX_UPLOAD_SIZE must be 1–1024 MB")
	}
	data := env("DATA_DIR", "data")
	store, err := storage.Open(env("DATABASE_PATH", filepath.Join(data, "lector-v2.db")))
	if err != nil {
		return err
	}
	defer store.Close()
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	plugins, err := extensions.New(env("PLUGIN_DIR", "plugins"), filepath.Join(data, "plugin-grants.json"), executable)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go plugins.Watch(ctx)
	handler := httpapi.New(store, plugins, httpapi.Config{Files: filepath.Join(data, "files"), Public: env("PUBLIC_DIR", "frontend/dist"), User: user, Password: password, MaxUpload: size << 20, AllowedHosts: strings.Split(env("ALLOWED_HOSTS", "localhost,127.0.0.1"), ",")})
	server := &http.Server{Addr: address, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 2 * time.Minute, WriteTimeout: 2 * time.Minute, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32 << 10}
	done := make(chan error, 1)
	go func() { slog.Info("Lector is ready", "address", address); done <- server.ListenAndServe() }()
	select {
	case err := <-done:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdown)
	}
}
func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
