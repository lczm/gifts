package main

import (
	"errors"
	"fmt"
	"time"

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

func CheckCanRedeem(db *gorm.DB, teamName string) (bool, error) {
	var redemption RedemptionEntry
	if err := db.First(&redemption, "team_name = ?", teamName).Error; err != nil {
		// if the record is not found, then the team has not redeemed
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func InsertRedemption(db *gorm.DB, teamName string) (RedemptionEntry, error) {
	redemption := RedemptionEntry{
		TeamName:   teamName,
		RedeemedAt: time.Now(),
	}

	// open a transaction to prevent multiple writers to the same row
	err := db.Transaction(func(tx *gorm.DB) error {
		// once a transaction opened, check for existing redemption
		var existing RedemptionEntry
		if err := tx.First(&existing, "team_name = ?", teamName).Error; err == nil {
			return fmt.Errorf("team '%s' has already redeemed their gift", teamName)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// no existing transaction, try to insert and close the transaction
		if err := tx.Create(&redemption).Error; err != nil {
			return err
		}
		return nil
	})

	// error-ed out during the db transaction
	if err != nil {
		return RedemptionEntry{}, err
	}

	// no issues with insertion
	return redemption, nil
}
