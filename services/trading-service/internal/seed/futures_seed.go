package seed

import (
	"encoding/csv"
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

func SeedFuturesContracts(db *gorm.DB) error {
	_, filename, _, _ := runtime.Caller(0)
	csvPath := filepath.Join(filepath.Dir(filename), "futures_with_dates.csv")

	f, err := os.Open(csvPath)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	for i, row := range records {
		// skip header
		if i == 0 {
			continue
		}

		if len(row) != 5 {
			log.Printf("invalid row length at line %d", i+1)
			continue
		}

		size, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			log.Printf("invalid contract size at line %d: %v", i+1, err)
			continue
		}

		date, err := time.Parse("2006-01-02", row[4])
		if err != nil {
			log.Printf("invalid date at line %d: %v", i+1, err)
			continue
		}

		contract := model.FuturesContract{
			Ticker:         row[0],
			Name:           row[1],
			ContractSize:   size,
			ContractUnit:   row[3],
			SettlementDate: date,
		}

		var existing model.FuturesContract
		err = db.Where("ticker = ?", contract.Ticker).First(&existing).Error
		if err == nil {
			continue // Skip if contract with that ticker already exists
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
