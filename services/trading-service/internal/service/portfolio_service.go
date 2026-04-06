package service

import (
	"context"
	"time"

	pkgerrors "github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
)

type PortfolioService struct {
	ownershipRepo repository.OrderOwnershipRepository
	stockRepo     repository.StockRepository
	optionRepo    repository.OptionRepository
	futuresRepo   repository.FuturesContractRepository
	forexRepo     repository.ForexRepository
}

func NewPortfolioService(
	ownershipRepo repository.OrderOwnershipRepository,
	stockRepo repository.StockRepository,
	optionRepo repository.OptionRepository,
	futuresRepo repository.FuturesContractRepository,
	forexRepo repository.ForexRepository,
) *PortfolioService {
	return &PortfolioService{
		ownershipRepo: ownershipRepo,
		stockRepo:     stockRepo,
		optionRepo:    optionRepo,
		futuresRepo:   futuresRepo,
		forexRepo:     forexRepo,
	}
}

func (s *PortfolioService) GetPortfolio(ctx context.Context, identityID uint, ownerType model.OwnerType) ([]dto.PortfolioAssetResponse, error) {
	ownerships, err := s.ownershipRepo.FindByIdentity(ctx, identityID, ownerType)
	if err != nil {
		return nil, pkgerrors.InternalErr(err)
	}

	type aggregated struct {
		ticker        string
		netAmount     float64
		totalBuyQty   float64
		totalBuyValue float64
		lastModified  time.Time
		currentPrice  float64
	}

	byListing := make(map[uint]*aggregated)

	for _, own := range ownerships {
		ord := own.Order
		if ord.Status != model.OrderStatusApproved {
			continue
		}
		if ord.FilledQty == 0 || ord.PricePerUnit == nil {
			continue
		}

		contractSize := ord.ContractSize
		if contractSize <= 0 {
			contractSize = 1
		}
		filledF := float64(ord.FilledQty) * contractSize
		listingID := ord.ListingID

		agg, exists := byListing[listingID]
		if !exists {
			agg = &aggregated{
				ticker:       ord.Listing.Ticker,
				currentPrice: ord.Listing.Price,
			}
			byListing[listingID] = agg
		}

		if ord.UpdatedAt.After(agg.lastModified) {
			agg.lastModified = ord.UpdatedAt
		}

		switch ord.Direction {
		case model.OrderDirectionBuy:
			agg.netAmount += filledF
			agg.totalBuyQty += filledF
			agg.totalBuyValue += (*ord.PricePerUnit) * filledF
		case model.OrderDirectionSell:
			agg.netAmount -= filledF
		}
	}

	var listingIDs []uint
	for id, agg := range byListing {
		if agg.netAmount > 0 {
			listingIDs = append(listingIDs, id)
		}
	}

	if len(listingIDs) == 0 {
		return []dto.PortfolioAssetResponse{}, nil
	}

	type assetMeta struct {
		assetType         dto.AssetType
		outstandingShares *float64
	}
	meta := make(map[uint]assetMeta)

	// Updated repo calls to include ctx
	stocks, err := s.stockRepo.FindByListingIDs(ctx, listingIDs)
	if err != nil {
		return nil, pkgerrors.InternalErr(err)
	}
	for _, st := range stocks {
		shares := st.OutstandingShares
		meta[st.ListingID] = assetMeta{
			assetType:         dto.AssetTypeStock,
			outstandingShares: &shares,
		}
	}

	options, err := s.optionRepo.FindByListingIDs(ctx, listingIDs)
	if err != nil {
		return nil, pkgerrors.InternalErr(err)
	}
	for _, op := range options {
		meta[op.ListingID] = assetMeta{assetType: dto.AssetTypeOption}
	}

	futures, err := s.futuresRepo.FindByListingIDs(ctx, listingIDs)
	if err != nil {
		return nil, pkgerrors.InternalErr(err)
	}
	for _, fc := range futures {
		meta[fc.ListingID] = assetMeta{assetType: dto.AssetTypeFutures}
	}

	forexPairs, err := s.forexRepo.FindByListingIDs(ctx, listingIDs)
	if err != nil {
		return nil, pkgerrors.InternalErr(err)
	}
	for _, fp := range forexPairs {
		meta[fp.ListingID] = assetMeta{assetType: dto.AssetTypeForex}
	}

	var result []dto.PortfolioAssetResponse

	for id, agg := range byListing {
		if agg.netAmount <= 0 {
			continue
		}
		m, known := meta[id]
		if !known {
			continue
		}

		var avgBuyPrice float64
		if agg.totalBuyQty > 0 {
			avgBuyPrice = agg.totalBuyValue / agg.totalBuyQty
		}

		profit := (agg.currentPrice - avgBuyPrice) * agg.netAmount

		result = append(result, dto.PortfolioAssetResponse{
			Type:              m.assetType,
			Ticker:            agg.ticker,
			Amount:            agg.netAmount,
			PricePerUnit:      agg.currentPrice,
			LastModified:      agg.lastModified,
			Profit:            profit,
			OutstandingShares: m.outstandingShares,
		})
	}

	return result, nil
}
