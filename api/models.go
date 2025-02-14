package main

import "time"

// db table entry
type MappingEntry struct {
	StaffPassID string    `json:"staff_pass_id" gorm:"primaryKey"`
	TeamName    string    `json:"team_name"`
	CreatedAt   time.Time `json:"created_at"`
}

// db table entry
type RedemptionEntry struct {
	TeamName   string    `json:"team_name" gorm:"primaryKey"`
	RedeemedAt time.Time `json:"redeemed_at"`
}

// post request payload body
type RedemptionPayload struct {
	StaffPassID string `json:"staff_pass_id"`
}
