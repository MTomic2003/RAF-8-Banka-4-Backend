package job

import (
	"context"
	"log"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
)

type DailyPriceJob struct {
	listingRepo repository.ListingRepository
}

func NewDailyPriceJob(listingRepo repository.ListingRepository) *DailyPriceJob {
	return &DailyPriceJob{listingRepo: listingRepo}
}

func (j *DailyPriceJob) Run(ctx context.Context) error {
	today := time.Now().Truncate(24 * time.Hour)
	if err := j.createListingDailyInfos(ctx, today); err != nil {
		log.Printf("Failed to create listing daily infos: %v", err)
		return err
	}
	log.Println("Daily price info job completed successfully")
	return nil
}

func (j *DailyPriceJob) createListingDailyInfos(ctx context.Context, date time.Time) error {
	listings, err := j.listingRepo.FindAll(ctx)
	if err != nil {
		return err
	}
	for _, listing := range listings {
		prevInfo, _ := j.listingRepo.FindLastDailyPriceInfo(ctx, listing.ListingID, date)
		change := 0.0
		if prevInfo != nil && prevInfo.Price != 0 {
			change = (listing.Price - prevInfo.Price) / prevInfo.Price * 100
		}
		dailyInfo := &model.ListingDailyPriceInfo{
			ListingID: listing.ListingID,
			Date:      date,
			Price:     listing.Price,
			Ask:       listing.Ask,
			Bid:       0.0,
			Change:    change,
			Volume:    0,
		}
		if err := j.listingRepo.CreateDailyPriceInfo(ctx, dailyInfo); err != nil {
			log.Printf("Failed to save daily price for listing %d: %v", listing.ListingID, err)

		}
	}
	return nil
}