package seed

import (
	_ "embed"
	"encoding/csv"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

//go:embed futures_with_dates.csv
var futuresCSV string

func SeedFuturesContracts(db *gorm.DB) error {
	reader := csv.NewReader(strings.NewReader(futuresCSV))
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	for i, row := range records {
		// skip header
		if i == 0 {
			continue
		}

		if len(row) != 7 {
			log.Printf("invalid row length at line %d", i+1)
			continue
		}

		size, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			log.Printf("invalid contract size at line %d: %v", i+1, err)
			continue
		}

		price, err := strconv.ParseFloat(row[5], 64)
		if err != nil {
			log.Printf("invalid price at line %d: %v", i+1, err)
			continue
		}

		date, err := time.Parse("2006-01-02", row[4])
		if err != nil {
			log.Printf("invalid date at line %d: %v", i+1, err)
			continue
		}

		contract := model.FuturesContract{
			ContractSize:   size,
			ContractUnit:   row[3],
			SettlementDate: date,
			Listing: model.Listing{
				Ticker:      row[0],
				Name:        row[1],
				ExchangeMIC: resolveExistingExchangeMIC(db, row[6]),
				LastRefresh: time.Now(),
				Price:       price * size,
				Ask:         price * size,
				ListingType: model.ListingTypeFuture,
			},
		}

		var existingListing model.Listing
		err = db.Where("ticker = ?", contract.Listing.Ticker).First(&existingListing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err := db.Create(&contract).Error; err != nil {
			return err
		}
	}

	return nil
}
