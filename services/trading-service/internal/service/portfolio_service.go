package service

import (
	"context"

	pkgerrors "github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
)

type PortfolioService struct {
	ownershipRepo repository.AssetOwnershipRepository
	stockRepo     repository.StockRepository
	optionRepo    repository.OptionRepository
	futuresRepo   repository.FuturesContractRepository
	forexRepo     repository.ForexRepository
}

func NewPortfolioService(
	ownershipRepo repository.AssetOwnershipRepository,
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

	// Filter to positive positions and collect asset IDs
	var active []model.AssetOwnership
	var assetIDs []uint
	for _, o := range ownerships {
		if o.Amount > 0 {
			active = append(active, o)
			assetIDs = append(assetIDs, o.AssetID)
		}
	}

	if len(active) == 0 {
		return []dto.PortfolioAssetResponse{}, nil
	}

	// Determine asset types; listing is preloaded on each asset type
	type assetMeta struct {
		assetType         dto.AssetType
		outstandingShares *float64
		listing           *model.Listing
	}
	meta := make(map[uint]assetMeta)

	stocks, err := s.stockRepo.FindByAssetIDs(ctx, assetIDs)
	if err != nil {
		return nil, pkgerrors.InternalErr(err)
	}
	for _, st := range stocks {
		shares := st.OutstandingShares
		meta[st.AssetID] = assetMeta{
			assetType:         dto.AssetTypeStock,
			outstandingShares: &shares,
			listing:           st.Listing,
		}
	}

	options, err := s.optionRepo.FindByAssetIDs(ctx, assetIDs)
	if err != nil {
		return nil, pkgerrors.InternalErr(err)
	}
	for _, op := range options {
		meta[op.AssetID] = assetMeta{assetType: dto.AssetTypeOption, listing: op.Listing}
	}

	futures, err := s.futuresRepo.FindByAssetIDs(ctx, assetIDs)
	if err != nil {
		return nil, pkgerrors.InternalErr(err)
	}
	for _, fc := range futures {
		meta[fc.AssetID] = assetMeta{assetType: dto.AssetTypeFutures, listing: fc.Listing}
	}

	forexPairs, err := s.forexRepo.FindByAssetIDs(ctx, assetIDs)
	if err != nil {
		return nil, pkgerrors.InternalErr(err)
	}
	for _, fp := range forexPairs {
		meta[fp.AssetID] = assetMeta{assetType: dto.AssetTypeForex, listing: fp.Listing}
	}

	var result []dto.PortfolioAssetResponse
	const taxRate = 0.15
	// TODO: calculate TaxAmount

	for _, o := range active {
		m, known := meta[o.AssetID]
		if !known {
			continue
		}

		currentPrice := 0.0
		if m.listing != nil {
			currentPrice = m.listing.Price
		}

		profit := (currentPrice - o.AvgBuyPrice) * o.Amount

		var ticker string
		if o.Asset.Ticker != "" {
			ticker = o.Asset.Ticker
		}

		result = append(result, dto.PortfolioAssetResponse{
			Type:              m.assetType,
			Ticker:            ticker,
			Amount:            o.Amount,
			PricePerUnit:      currentPrice,
			LastModified:      o.UpdatedAt,
			Profit:            profit,
			OutstandingShares: m.outstandingShares,
		})
	}

	return result, nil
}
