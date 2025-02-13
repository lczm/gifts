package main

import "time"

type MappingEntry struct {
	StaffPassID string    `json:"staff_pass_id" gorm:"primaryKey"`
	TeamName    string    `json:"team_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type RedemptionEntry struct {
	TeamName   string    `json:"team_name" gorm:"primaryKey"`
	RedeemedAt time.Time `json:"redeemed_at"`
}
