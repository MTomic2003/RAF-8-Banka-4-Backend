package job

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
	"github.com/stretchr/testify/require"
)

type mockListingRepo struct {
	listings                 []model.Listing
	findAllErr               error
	findLastDailyInfo        *model.ListingDailyPriceInfo
	findLastDailyInfoErr     error
	createDailyInfoCalled    []model.ListingDailyPriceInfo
	createDailyInfoFailFirst bool
	createDailyInfoCall      int
}

func (m *mockListingRepo) FindAll(ctx context.Context) ([]model.Listing, error) {
	return m.listings, m.findAllErr
}

func (m *mockListingRepo) FindLastDailyPriceInfo(ctx context.Context, listingID uint, beforeDate time.Time) (*model.ListingDailyPriceInfo, error) {
	return m.findLastDailyInfo, m.findLastDailyInfoErr
}

func (m *mockListingRepo) CreateDailyPriceInfo(ctx context.Context, info *model.ListingDailyPriceInfo) error {
	if m.createDailyInfoFailFirst && m.createDailyInfoCall == 0 {
		m.createDailyInfoCall++
		return errors.New("db error")
	}
	m.createDailyInfoCall++
	m.createDailyInfoCalled = append(m.createDailyInfoCalled, *info)
	return nil
}

func (m *mockListingRepo) FindStocks(ctx context.Context, filter repository.ListingFilter) ([]model.Listing, int64, error) {
	return nil, 0, nil
}
func (m *mockListingRepo) FindFutures(ctx context.Context, filter repository.ListingFilter) ([]model.Listing, int64, error) {
	return nil, 0, nil
}
func (m *mockListingRepo) FindOptions(ctx context.Context, filter repository.ListingFilter) ([]model.Listing, int64, error) {
	return nil, 0, nil
}
func (m *mockListingRepo) FindByID(ctx context.Context, id uint, daysBack int) (*model.Listing, error) {
	return nil, nil
}
func (m *mockListingRepo) Upsert(ctx context.Context, listing *model.Listing) error { return nil }
func (m *mockListingRepo) UpdatePriceAndAsk(ctx context.Context, listing *model.Listing, price, ask float64) error {
	return nil
}
func (m *mockListingRepo) Count(ctx context.Context) (int64, error) { return 0, nil }
func (m *mockListingRepo) FindByAssetType(ctx context.Context, typ model.AssetType) ([]model.Listing, error) {
	return nil, nil
}

func (m *mockListingRepo) FindByAssetIDs(ctx context.Context, assetIDs []uint) ([]model.Listing, error) {
	return nil, nil
}

func (m *mockListingRepo) FindLatestDailyPriceInfo(ctx context.Context, listingID uint) (*model.ListingDailyPriceInfo, error) {
	return nil, nil
}

func TestDailyPriceJob_Run(t *testing.T) {
	tests := []struct {
		name                     string
		listings                 []model.Listing
		findAllErr               error
		prevListingInfo          *model.ListingDailyPriceInfo
		createDailyInfoFailFirst bool
		expectedListingInfos     int
		expectError              bool
	}{
		{
			name: "happy path - stocks and forex",
			listings: []model.Listing{
				{ListingID: 1, Price: 150.0, Ask: 151.0},
				{ListingID: 2, Price: 1.20, Ask: 1.21},
			},
			expectedListingInfos: 2,
		},
		{
			name: "with previous day data - change calculated",
			listings: []model.Listing{
				{ListingID: 1, Price: 150.0},
			},
			prevListingInfo:      &model.ListingDailyPriceInfo{Price: 100.0},
			expectedListingInfos: 1,
		},
		{
			name: "error on listing create - continues with next",
			listings: []model.Listing{
				{ListingID: 1, Price: 150.0},
				{ListingID: 2, Price: 200.0},
			},
			createDailyInfoFailFirst: true,
			expectedListingInfos:     1, // second succeeds
		},
		{
			name:        "find all error",
			findAllErr:  errors.New("db error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockListing := &mockListingRepo{
				listings:                 tt.listings,
				findAllErr:               tt.findAllErr,
				findLastDailyInfo:        tt.prevListingInfo,
				createDailyInfoFailFirst: tt.createDailyInfoFailFirst,
				createDailyInfoCalled:    []model.ListingDailyPriceInfo{},
			}

			job := NewDailyPriceJob(mockListing)
			err := job.Run(context.Background())

			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Len(t, mockListing.createDailyInfoCalled, tt.expectedListingInfos)

			if tt.prevListingInfo != nil && len(mockListing.createDailyInfoCalled) > 0 {
				info := mockListing.createDailyInfoCalled[0]
				expectedChange := (150.0 - 100.0) / 100.0 * 100
				require.InDelta(t, expectedChange, info.Change, 0.001)
			}
		})
	}
}
