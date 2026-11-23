// Command j3d runs the J3 venue server and manages local demo data.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/restful-otaku/j3/internal/api"
	"github.com/restful-otaku/j3/internal/config"
	"github.com/restful-otaku/j3/internal/db"
	"github.com/restful-otaku/j3/internal/seed"
	"github.com/restful-otaku/j3/internal/simulator"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "seed":
			must(seedCommand(os.Args[2:], false))
			return
		case "reset":
			must(seedCommand(os.Args[2:], true))
			return
		case "simulate":
			must(simulateCommand(os.Args[2:]))
			return
		}
	}
	serve()
}

func serve() {
	cfg := config.Load()
	addr := flag.String("addr", cfg.Addr, "listen address")
	dbPath := flag.String("db", cfg.DBPath, "SQLite database path")
	version := flag.String("version", cfg.Version, "build version")
	seedDemo := flag.Bool("seed", cfg.SeedDemo, "ensure demo data exists on boot")
	flag.Parse()

	store, err := db.Open(*dbPath)
	must(err)
	defer store.Close()

	generatedOwnerPassword, err := seed.EnsureOwner(store, cfg, *seedDemo)
	must(err)
	if generatedOwnerPassword != "" {
		log.Printf("J3 owner admin created with a generated one-time password: %s", generatedOwnerPassword)
		log.Printf("Sign in as %s, then change it at /login", cfg.OwnerEmail)
	}

	if *seedDemo {
		must(seed.Ensure(store))
		log.Printf("demo data ready")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	venue := api.New(store, *version)
	venue.StartClock(ctx)
	server := &http.Server{
		Addr:              *addr,
		Handler:           venue.Handler(),
		ReadHeaderTimeout: cfg.Timeout,
	}
	go func() {
		log.Printf("J3 venue server listening on %s (db: %s)", *addr, *dbPath)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("serve: %v", err)
			server.Close()
			return
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func seedCommand(args []string, reset bool) error {
	cfg := config.Load()
	flags := flag.NewFlagSet("seed", flag.ContinueOnError)
	dbPath := flags.String("db", cfg.DBPath, "SQLite database path")
	full := flags.Bool("full", false, "seed a complete verifiable tournament")
	okami := flags.Bool("okami", true, "seed the Okami Jiu Jitsu Challenger demo")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if reset {
		for _, path := range []string{*dbPath, *dbPath + "-wal", *dbPath + "-shm"} {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove %s: %w", path, err)
			}
		}
	}
	store, err := db.Open(*dbPath)
	if err != nil {
		return err
	}
	defer store.Close()
	if err := seed.Ensure(store); err != nil {
		return err
	}
	if *okami {
		if err := seed.EnsureOkami(store); err != nil {
			return err
		}
		log.Printf("Okami Jiu Jitsu Challengers Tournament 2027 seeded (mid-event)")
	}
	if *full {
		if err := seed.EnsureFull(store); err != nil {
			return err
		}
		log.Printf("full tournament seeded")
	}
	log.Printf("demo data ready in %s", *dbPath)
	return nil
}

// simulateCommand emits the same deterministic snapshot as the public demo
// endpoint, so QA containers can generate a tournament without curl or jq.
func simulateCommand(args []string) error {
	flags := flag.NewFlagSet("simulate", flag.ContinueOnError)
	seed := flags.Int64("seed", 42, "simulation seed")
	fighters := flags.Int("fighters", 50, "number of fighters")
	out := flags.String("out", "", "write JSON snapshot to this file (default: stdout)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	result := simulator.Run(*seed, *fighters)
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if *out == "" {
		_, err = os.Stdout.Write(data)
		return err
	}
	if dir := filepath.Dir(*out); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(*out, data, 0o644)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
