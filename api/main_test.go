package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func getTestSystem(t *testing.T) *GiftRedemptionSystem {
	// open in memory sqlite for tests so theres no need to clean up
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to create in-memory sqlite db: %v", err)
	}

	if err := initDB(db); err != nil {
		log.Fatalf("failed to auto migrate tables: %v", err)
	}

	// taken from staff-id-to-team-mapping
	testMappings := []MappingEntry{
		{StaffPassID: "STAFF_H123804820G", TeamName: "BASS", CreatedAt: time.UnixMilli(1623772799000)},
		{StaffPassID: "MANAGER_T999888420B", TeamName: "RUST", CreatedAt: time.UnixMilli(1623772799000)},
		{StaffPassID: "BOSS_T000000001P", TeamName: "RUST", CreatedAt: time.UnixMilli(1623872111000)},
	}

	for _, m := range testMappings {
		if err := db.Create(&m).Error; err != nil {
			t.Fatalf("error inserting mapping entry row: %v", err)
		}
	}

	return &GiftRedemptionSystem{db: db}
}

func getTestRouter(t *testing.T) *chi.Mux {
	system := getTestSystem(t)
	r := chi.NewRouter()
	r.Get("/lookup", system.handleLookup)
	r.Post("/redeem", system.handleRedemption)
	return r
}

func TestLookupStaffPass(t *testing.T) {
	system := getTestSystem(t)

	// successful case, this will pass
	mapping, err := GetStaffPass(system.db, "STAFF_H123804820G")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mapping.TeamName != "BASS" {
		t.Fatalf("expected team BASS, got %s", mapping.TeamName)
	}

	// unsuccessful case, this wont pass, the staff pass id does not exist
	_, err = GetStaffPass(system.db, "NON_EXISTENT")
	if err == nil {
		t.Fatalf("expected error for non-existent staff pass, got nil")
	}

	// unsuccessful case, the staff pass id is case sensitive
	_, err = GetStaffPass(system.db, "STAFF_H123804820g")
	if err == nil {
		t.Fatalf("expected error for non-existent staff pass, got nil")
	}
}

func TestLookupEndpoint(t *testing.T) {
	router := getTestRouter(t)
	// create request with an existing staff pass id
	req := httptest.NewRequest(http.MethodGet, "/lookup?staff_pass_id=STAFF_H123804820G", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var mapping MappingEntry
	if err := json.Unmarshal(rec.Body.Bytes(), &mapping); err != nil {
		t.Fatalf("error unmarshalling response: %v", err)
	}
	if mapping.TeamName != "BASS" {
		t.Fatalf("expected team BASS, got %s", mapping.TeamName)
	}
}

func TestLookupEndpoint_EmptyStaffQuery(t *testing.T) {
	router := getTestRouter(t)
	// create a request (that will fail) because there was no staff pass id
	req := httptest.NewRequest(http.MethodGet, "/lookup", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for empty staff_pass_id, got %d", rec.Code)
	}
}
