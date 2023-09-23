package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"libdb.so/hserve"
)

var addr = ":8080"

func main() {
	flag.StringVar(&addr, "addr", addr, "address to listen on")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		log.Fatalln("DB_PATH is not set")
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger) // log all requests
	r.Use(allowAllCORS())    // allow all CORS requests

	r.Mount("/api", newExerciseHandler(dbPath))
	r.Mount("/", http.FileServer(http.Dir("static")))

	// Start the server.
	log.Println("listening on", addr)
	hserve.MustListenAndServe(ctx, addr, r)
}
