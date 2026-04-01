package service

import (
	"context"

	commonErrors "github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
)

type ListingService struct {
	listingRepo repository.ListingRepository
	futuresRepo repository.FuturesContractRepository
	forexRepo   repository.ForexRepository
	optionRepo  repository.OptionRepository
}

func NewListingService(
	listingRepo repository.ListingRepository,
	futuresRepo repository.FuturesContractRepository,
	forexRepo repository.ForexRepository,
	optionRepo repository.OptionRepository,
) *ListingService {
	return &ListingService{
		listingRepo: listingRepo,
		futuresRepo: futuresRepo,
		forexRepo:   forexRepo,
		optionRepo:  optionRepo,
	}
}

// --- Helpers ---

func latestDaily(infos []model.ListingDailyPriceInfo) *model.ListingDailyPriceInfo {
	if len(infos) == 0 {
		return nil
	}

	latest := infos[0]
	for _, info := range infos[1:] {
		if info.Date.After(latest.Date) {
			latest = info
		}
	}
	return &latest
}

func baseResponse(l model.Listing, daily *model.ListingDailyPriceInfo) dto.BaseListingResponse {
	r := dto.BaseListingResponse{
		ListingID:         l.ListingID,
		Ticker:            l.Ticker,
		Name:              l.Name,
		Exchange:          l.ExchangeMIC,
		Price:             l.Price,
		Ask:               l.Ask,
		MaintenanceMargin: l.MaintenanceMargin,
		InitialMarginCost: l.MaintenanceMargin * 1.1,
	}
	if daily != nil {
		r.Bid = daily.Bid
		r.Change = daily.Change
		r.Volume = daily.Volume
	}
	return r
}

func toFilter(q dto.ListingQuery) (repository.ListingFilter, error) {
	f := repository.ListingFilter{
		Search:    q.Search,
		Exchange:  q.Exchange,
		PriceMin:  q.PriceMin,
		PriceMax:  q.PriceMax,
		AskMin:    q.AskMin,
		AskMax:    q.AskMax,
		BidMin:    q.BidMin,
		BidMax:    q.BidMax,
		VolumeMin: q.VolumeMin,
		VolumeMax: q.VolumeMax,
		SortBy:    q.SortBy,
		SortDir:   q.SortDir,
		Page:      q.Page,
		PageSize:  q.PageSize,
	}
	sd, err := q.ParseSettlementDate()
	if err != nil {
		return f, err
	}
	f.SettlementDate = sd
	return f, nil
}

// --- Stocks ---

func (s *ListingService) GetStocks(ctx context.Context, q dto.ListingQuery) (*dto.PaginatedResponse[dto.StockResponse], error) {
	filter, err := toFilter(q)
	if err != nil {
		return nil, commonErrors.BadRequestErr("invalid settlement_date format")
	}

	listings, total, err := s.listingRepo.FindStocks(ctx, filter)
	if err != nil {
		return nil, commonErrors.InternalErr(err)
	}

	data := make([]dto.StockResponse, len(listings))
	for i, l := range listings {
		daily := latestDaily(l.DailyPriceInfos)
		var outstandingShares, dividendYield float64
		if l.Stock != nil {
			outstandingShares = l.Stock.OutstandingShares
			dividendYield = l.Stock.DividendYield
		}
		data[i] = dto.StockResponse{
			BaseListingResponse: baseResponse(l, daily),
			OutstandingShares:   outstandingShares,
			DividendYield:       dividendYield,
		}
	}

	return &dto.PaginatedResponse[dto.StockResponse]{
		Data:     data,
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
	}, nil
}

// --- Futures ---

func (s *ListingService) GetFutures(ctx context.Context, q dto.ListingQuery) (*dto.PaginatedResponse[dto.FuturesResponse], error) {
	filter, err := toFilter(q)
	if err != nil {
		return nil, commonErrors.BadRequestErr("invalid settlement_date format")
	}

	listings, total, err := s.listingRepo.FindFutures(ctx, filter)
	if err != nil {
		return nil, commonErrors.InternalErr(err)
	}

	// IZMENA: koristimo ListingIDs umesto tickera
	ids := make([]uint, len(listings))
	for i, l := range listings {
		ids[i] = l.ListingID
	}

	contracts, err := s.futuresRepo.FindByListingIDs(ctx, ids)
	if err != nil {
		return nil, commonErrors.InternalErr(err)
	}

	contractMap := make(map[uint]model.FuturesContract)
	for _, fc := range contracts {
		contractMap[fc.ListingID] = fc
	}

	data := make([]dto.FuturesResponse, len(listings))
	for i, l := range listings {
		daily := latestDaily(l.DailyPriceInfos)
		fc := contractMap[l.ListingID] // IZMENA
		data[i] = dto.FuturesResponse{
			BaseListingResponse: baseResponse(l, daily),
			SettlementDate:      fc.SettlementDate,
			ContractSize:        fc.ContractSize,
			ContractUnit:        fc.ContractUnit,
		}
	}

	return &dto.PaginatedResponse[dto.FuturesResponse]{
		Data:     data,
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
	}, nil
}

// --- Forex ---

func (s *ListingService) GetForex(ctx context.Context, q dto.ListingQuery) (*dto.PaginatedResponse[dto.ForexResponse], error) {
	filter, err := toFilter(q)
	if err != nil {
		return nil, commonErrors.BadRequestErr("invalid settlement_date format")
	}

	pairs, total, err := s.forexRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, commonErrors.InternalErr(err)
	}

	data := make([]dto.ForexResponse, len(pairs))
	for i, p := range pairs {
		data[i] = dto.ForexResponse{
			ForexPairID: p.ForexPairID,
			Ticker:      p.Base + "/" + p.Quote,
			Base:        p.Base,
			Quote:       p.Quote,
			Price:       p.Rate,
		}
	}

	return &dto.PaginatedResponse[dto.ForexResponse]{
		Data:     data,
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
	}, nil
}

func (s *ListingService) GetOptions(ctx context.Context, q dto.ListingQuery) (*dto.PaginatedResponse[dto.OptionResponse], error) {
	filter, err := toFilter(q)
	if err != nil {
		return nil, commonErrors.BadRequestErr("invalid settlement_date format")
	}

	listings, total, err := s.listingRepo.FindOptions(ctx, filter)
	if err != nil {
		return nil, commonErrors.InternalErr(err)
	}

	// batch fetch options po listing ID-evima
	ids := make([]uint, len(listings))
	for i, l := range listings {
		ids[i] = l.ListingID
	}

	options, err := s.optionRepo.FindByListingIDs(ctx, ids)
	if err != nil {
		return nil, commonErrors.InternalErr(err)
	}

	optionMap := make(map[uint]model.Option)
	for _, o := range options {
		optionMap[o.ListingID] = o
	}

	data := make([]dto.OptionResponse, len(listings))
	for i, l := range listings {
		daily := latestDaily(l.DailyPriceInfos)
		o := optionMap[l.ListingID]
		data[i] = dto.OptionResponse{
			BaseListingResponse: baseResponse(l, daily),
			Strike:              o.StrikePrice,
			OptionType:          string(o.OptionType),
			SettlementDate:      o.SettlementDate,
			ImpliedVolatility:   o.ImpliedVolatility,
			OpenInterest:        o.OpenInterest,
		}
	}

	result := &dto.PaginatedResponse[dto.OptionResponse]{
		Data:     data,
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
	}
	return result, nil
}
