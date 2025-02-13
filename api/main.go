package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type GiftRedemptionSystem struct {
	db *gorm.DB
}

func (g *GiftRedemptionSystem) handleLookup(w http.ResponseWriter, r *http.Request) {
}

func (g *GiftRedemptionSystem) handleRedemption(w http.ResponseWriter, r *http.Request) {

}

func main() {
	// all possible
	dbPath := flag.String("db", "", "Path to existing database. If it does not exist, then it will be created")
	flag.Parse()

	if *dbPath == "" {
		*dbPath = "gifts.db"
	}

	// if there is no existing db, then create new one
	if _, err := os.Stat(*dbPath); os.IsNotExist(err) {
		log.Printf("Creating a new database %s", *dbPath)
	}

	// load sqlite db into a gorm object and pass it around the methods
	gormDB, err := gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	system := &GiftRedemptionSystem{db: gormDB}

	// router services
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/lookup", system.handleLookup)
	r.Post("/redemption", system.handleRedemption)
	http.ListenAndServe(":3000", r)
}
