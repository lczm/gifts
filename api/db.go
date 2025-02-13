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
