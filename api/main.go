package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type GiftRedemptionSystem struct {
	db *gorm.DB
}

func (g *GiftRedemptionSystem) handleLookup(w http.ResponseWriter, r *http.Request) {
	staffPassID := r.URL.Query().Get("staff_pass_id")
	// if query param does not include staff_pass_id, then nothing to look up
	if staffPassID == "" {
		http.Error(w, "staff_pass_id is needed to lookup redemption", http.StatusBadRequest)
		return
	}

	mapping, err := GetStaffPass(g.db, staffPassID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapping)
}

func (g *GiftRedemptionSystem) handleRedemption(w http.ResponseWriter, r *http.Request) {

}

func loadCsvMapping(db *gorm.DB, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("reading CSV header: %w", err)
	}
	// check that there is only 3 headers
	if len(headers) != 3 {
		return errors.New("invalid CSV format, should only have 3 columns in the header")
	}

	// read row by row, if somehow the data is too large
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading record: %v", err)
		}

		staffPassId := record[0]
		teamName := record[1]

		// convert epoch time to int
		createdAtEpochInt, err := strconv.ParseInt(record[2], 10, 64)
		if err != nil {
			return fmt.Errorf("error converting epoch time to integer : %s", record[2])
		}
		createdAt := time.UnixMilli(createdAtEpochInt)

		mappingEntry := MappingEntry{
			StaffPassID: staffPassId,
			TeamName:    teamName,
			CreatedAt:   createdAt,
		}

		fmt.Println(mappingEntry)

		if err := db.Save(&mappingEntry).Error; err != nil {
			return fmt.Errorf("error inserting mapping entry row: %w", err)
		}
	}

	return nil
}

func main() {
	// all possible flags
	dbPath := flag.String("db", "", "Path to existing database. If it does not exist, then it will be created")
	csvPath := flag.String("csv", "", "Path to CSV mapping file (Optional)")
	flag.Parse()

	if *dbPath == "" {
		*dbPath = "gifts.db"
	}

	// if there is no existing db, then create new one
	if _, err := os.Stat(*dbPath); os.IsNotExist(err) {
		log.Printf("creating a new database %s", *dbPath)
	}

	// load sqlite db into a gorm object and pass it around the methods
	gormDB, err := gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// init/update the db if needed
	if err := initDB(gormDB); err != nil {
		log.Fatalf("failed to auto migrate tables: %v", err)
	}

	// load the csv mapping into the database
	if *csvPath != "" {
		loadCsvMapping(gormDB, *csvPath)
	}

	system := &GiftRedemptionSystem{db: gormDB}

	// router services
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/lookup", system.handleLookup)
	r.Post("/redemption", system.handleRedemption)
	http.ListenAndServe(":3000", r)
}
