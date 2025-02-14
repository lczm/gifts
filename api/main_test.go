package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func getTestSystem(t *testing.T) *GiftRedemptionSystem {
	// open in memory sqlite for tests so theres no need to clean up
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
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
	r.Post("/redemption", system.handleRedemption)
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

func TestSequentialCheckCanRedeem(t *testing.T) {
	system := getTestSystem(t)

	// successful case, this will pass
	canRedeem, err := CheckCanRedeem(system.db, "BASS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !canRedeem {
		t.Fatalf("expected BASS to be able to redeem, got false")
	}

	// unsuccessful case, this wont pass, the team has already redeemed
	_, err = InsertRedemption(system.db, "BASS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	canRedeem, err = CheckCanRedeem(system.db, "BASS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if canRedeem {
		t.Fatalf("expected BASS to not be able to redeem, got true")
	}
}

func TestConcurrentCheckCanRedeem(t *testing.T) {
	system := getTestSystem(t)
	count := 0
	numAttempts := 100
	var wg sync.WaitGroup
	var mu sync.Mutex

	// spawn 100 go routines to check if RUST can redeem
	for i := 0; i < numAttempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// db transactions will handle the concurrency
			_, err := InsertRedemption(system.db, "RUST")
			// safely increment count
			if err == nil {
				mu.Lock()
				count++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if count != 1 {
		t.Fatalf("should only have 1 successful insertion, but got %v", count)
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

func TestLookupEndpointEmptyStaffQuery(t *testing.T) {
	router := getTestRouter(t)
	// create a request (that will fail) because there was no staff pass id
	req := httptest.NewRequest(http.MethodGet, "/lookup", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for empty staff_pass_id, got %d", rec.Code)
	}
}

func TestRedemptionEndpoint(t *testing.T) {
	router := getTestRouter(t)

	redeemReq := RedemptionPayload{
		StaffPassID: "MANAGER_T999888420B",
	}
	body, err := json.Marshal(redeemReq)
	if err != nil {
		t.Fatalf("error marshalling JSON: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/redemption", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var redemption RedemptionEntry
	if err := json.Unmarshal(rec.Body.Bytes(), &redemption); err != nil {
		t.Fatalf("error unmarshalling response: %v", err)
	}
	if redemption.TeamName != "RUST" {
		t.Fatalf("expected team RUST, got %s", redemption.TeamName)
	}

	// try to redeem again, this should fail
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req)
	if rec2.Code == http.StatusOK {
		t.Fatalf("expected second redemption to fail, got status %d", rec2.Code)
	}
}
