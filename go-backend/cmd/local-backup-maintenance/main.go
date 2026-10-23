// local-backup-maintenance can operate with the business database and login
// unavailable. It never starts application dependencies, migrations or jobs.
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"commerce-platform/internal/localbackup"
)

func main() {
	if err := runIndependentBackupMaintenanceCommand(); err != nil {
		log.Fatal(err)
	}
}

func runIndependentBackupMaintenanceCommand() error {
	configurationPath := flag.String("config", "", "absolute independent recovery configuration JSON file")
	listen := flag.String("listen", "127.0.0.1:19090", "maintenance HTTP listen address; use SSH tunnel or the scoped admin reverse proxy")
	checkConfigurationOnly := flag.Bool("check-configuration-only", false, "validate local configuration and emit a credential-free JSON report without starting the executor, database access or HTTP server")
	flag.Parse()
	if *configurationPath == "" || flag.NArg() != 0 {
		return errors.New("-config is required; no positional arguments accepted")
	}
	if *checkConfigurationOnly {
		return checkIndependentBackupConfigurationAndWriteCredentialFreeReport(*configurationPath, os.Stdout)
	}
	cfg, err := localbackup.LoadConfiguration(*configurationPath)
	if err != nil {
		return err
	}
	manager, err := localbackup.NewManager(cfg)
	if err != nil {
		return err
	}
	defer func(closeResource func() error) { _ = closeResource() }(manager.Close)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	server := &http.Server{Addr: *listen, Handler: manager.HTTPHandler(), ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 3 * time.Minute, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10}
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			cleanup, cancel := context.WithTimeout(ctx, 5*time.Minute)
			err := manager.ExpireWorkCopies(cleanup)
			cancel()
			if err != nil {
				log.Print("expired work copies retained because cleanup could not complete")
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			log.Print("independent maintenance HTTP shutdown did not complete within its deadline")
		}
	}()
	log.Print("independent local backup maintenance console listening; no business jobs started")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
