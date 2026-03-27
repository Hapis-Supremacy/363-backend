package main

import (
	"363project/controller"
	"363project/initializer"
	"363project/middleware"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

func init() {
	initializer.LoadEnvVar()   // Baca .env
	initializer.ConnecttoDB()  // Konek MySQL
	initializer.SyncDatabase() // Bikin tabel otomatis
}

func addRoutes(mux *http.ServeMux) {
	mux.Handle("/api/v1/ussd", controller.USSDHandler())
}

func newServer() http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux)

	var handler http.Handler = mux
	handler = middleware.AuthMiddleware(handler)
	return handler
}

func run(ctx context.Context, w io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	srv := newServer()

	httpServer := &http.Server{
		Addr:         net.JoinHostPort(host, port),
		Handler:      srv,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(w, "error listening and serving: %s\n", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(w, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()
	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
