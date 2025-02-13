package main

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func handleLookup(w http.ResponseWriter, r *http.Request) {
}

func handleRedemption(w http.ResponseWriter, r *http.Request) {

}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/lookup", handleLookup)
	r.Post("/redemption", handleRedemption)

	http.ListenAndServe(":3000", r)
}
