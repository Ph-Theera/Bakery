package wire

import (
	"bakery/cmd/bakery/wire/dep"
	"bakery/infrastructure/config"
	rest "bakery/internal/adapters/inbound/rest/v1"
	notificationsin "bakery/internal/ports/port_in/notifications"
	"context"
	"log"
	"net/http"
	"time"
)

func NewServer(ctx context.Context, c config.Config) (http.Handler, func(), error) {
	deps, e := dep.NewDependencies(ctx, c)
	if e != nil {
		return nil, nil, e
	}
	app := NewApplication(deps, c)
	handler := rest.NewRestServer(app.NewRestUsecases(), rest.Auth{Secret: c.AuthSecret, AdminToken: c.AdminToken, AdminID: c.AdminID}, c.LineSecret, c.AllowedOrigin).Handler()
	mux := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", handler))
	mux.Handle("/", handler)
	workerCtx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		if c.LineToken == "" || c.LineRecipient == "" {
			return
		}
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			jobCtx, stop := context.WithTimeout(workerCtx, 15*time.Second)
			e := app.dispatchNotificationsUc.Handle(jobCtx, notificationsin.DispatchNotificationsCommand{})
			stop()
			if e != nil && workerCtx.Err() == nil {
				log.Print("LINE outbox dispatch failed; messages remain queued")
			}
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return mux, func() { cancel(); <-done; deps.Close() }, nil
}
