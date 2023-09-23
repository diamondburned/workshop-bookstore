package main

import (
	"encoding/base64"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"libdb.so/workshop-bookstore/bookstore/api"
	"libdb.so/workshop-bookstore/bookstore/db"
)

const missingACMHeader = "" +
	`Hey! Looks like you're missing the X-ACM-Name header. Please add it to ` +
	`your request so we know who you are!`

func newExerciseHandler(dbPath string) http.Handler {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acmName := r.Header.Get("X-ACM-Name")
		if acmName == "" {
			http.Error(w, missingACMHeader, http.StatusBadRequest)
			return
		}

		// Truncate the name if it's too long.
		if len(acmName) > 50 {
			acmName = acmName[:50]
		}

		// Turn the string into base64 so that it's safe to use as a filename.
		pathPart := base64.URLEncoding.EncodeToString([]byte(acmName))

		// Create a new database for the user.
		dbPath := filepath.Join(dbPath, pathPart+".db")

		db, err := db.NewBookstore(dbPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		h := api.NewBookstoreHandler(db)
		h.ServeHTTP(w, r)
	})

	r := chi.NewRouter()
	r.Use(middleware.Throttle(100))
	r.Mount("/", h)

	return r
}

func allowAllCORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "*")
			w.Header().Set("Access-Control-Allow-Headers", "*")

			if r.Method == http.MethodOptions {
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
