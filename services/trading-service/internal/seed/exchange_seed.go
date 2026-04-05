package seed

import (
	_ "embed"
	"encoding/csv"
	"errors"
	"log"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

//go:embed exchanges_with_offset.csv
var exchangesCSV string

func RunExchangeSeed(db *gorm.DB) error {
	reader := csv.NewReader(strings.NewReader(exchangesCSV))
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	// skip header row
	for _, record := range records[1:] {
		if len(record) < 8 {
			continue
		}

		timeZone, tzErr := strconv.Atoi(strings.TrimSpace(record[5]))
		if tzErr != nil {
			log.Printf("invalid timezone for %s: %v, defaulting to 0", strings.TrimSpace(record[2]), tzErr)
			timeZone = 0
		}

		exchange := model.Exchange{
			Name:           strings.TrimSpace(record[0]),
			Acronym:        strings.TrimSpace(record[1]),
			MicCode:        strings.TrimSpace(record[2]),
			Polity:         strings.TrimSpace(record[3]),
			Currency:       strings.TrimSpace(record[4]),
			TimeZone:       timeZone,
			OpenTime:       strings.TrimSpace(record[6]),
			CloseTime:      strings.TrimSpace(record[7]),
			TradingEnabled: true,
		}

		var existing model.Exchange
		dbErr := db.Where("mic_code = ?", exchange.MicCode).First(&existing).Error
		if errors.Is(dbErr, gorm.ErrRecordNotFound) {
			if err := db.Create(&exchange).Error; err != nil {
				log.Printf("failed to create exchange %s: %v", exchange.MicCode, err)
				return err
			}
			log.Printf("created exchange: %s (%s)", exchange.Name, exchange.MicCode)
		} else if dbErr != nil {
			log.Printf("failed to query exchange %s: %v", exchange.MicCode, dbErr)
			return dbErr
		} else {
			log.Printf("exchange already exists: %s (%s)", exchange.Name, exchange.MicCode)
		}
	}

	return nil
}
