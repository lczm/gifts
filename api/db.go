package main

import (
	"fmt"

	"gorm.io/gorm"
)

func initDB(db *gorm.DB) error {
	if err := db.AutoMigrate(&MappingEntry{}, &RedemptionEntry{}); err != nil {
		return fmt.Errorf("error initializing database tables: %w", err)
	}
	return nil
}

func GetStaffPass(db *gorm.DB, staffPassID string) (MappingEntry, error) {
	var mapping MappingEntry
	if err := db.First(&mapping, "staff_pass_id = ?", staffPassID).Error; err != nil {
		return MappingEntry{}, fmt.Errorf("error looking up staff pass: %w", err)
	}
	return mapping, nil
}
