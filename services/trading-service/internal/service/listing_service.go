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

func mapHistory(infos []model.ListingDailyPriceInfo) []dto.DailyPriceResponse {
	history := make([]dto.DailyPriceResponse, len(infos))
	for i, h := range infos {
		history[i] = dto.DailyPriceResponse{
			Date: h.Date, Price: h.Price, Ask: h.Ask, Bid: h.Bid, Change: h.Change, Volume: h.Volume,
		}
	}
	return history
}

// --- Stocks ---

func (s *ListingService) GetStocks(ctx context.Context, q dto.ListingQuery) (*dto.PaginatedStockResponse, error) {
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

	return &dto.PaginatedStockResponse{
		Data:     data,
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
	}, nil
}
func (s *ListingService) GetStockDetails(ctx context.Context, listingID uint) (*dto.StockDetailedResponse, error) {
	l, err := s.listingRepo.FindByID(ctx, listingID)
	if err != nil {
		return nil, commonErrors.InternalErr(err)
	}
	if l == nil || l.ListingType != model.ListingTypeStock || l.Stock == nil {
		return nil, commonErrors.NotFoundErr("stock not found")
	}

	daily := latestDaily(l.DailyPriceInfos)
	stockResp := dto.StockResponse{
		BaseListingResponse: baseResponse(*l, daily),
		OutstandingShares:   l.Stock.OutstandingShares,
		DividendYield:       l.Stock.DividendYield,
	}

	history := make([]dto.DailyPriceResponse, len(l.DailyPriceInfos))
	for i, h := range l.DailyPriceInfos {
		history[i] = dto.DailyPriceResponse{
			Date:   h.Date,
			Price:  h.Price,
			Ask:    h.Ask,
			Bid:    h.Bid,
			Change: h.Change,
			Volume: h.Volume,
		}
	}

	options, err := s.optionRepo.FindByStockID(ctx, l.Stock.StockID)
	if err != nil {
		return nil, commonErrors.InternalErr(err)
	}

	optionResponses := make([]dto.OptionResponse, len(options))
	for i, o := range options {

		optionResponses[i] = dto.OptionResponse{
			BaseListingResponse: baseResponse(o.Listing, nil),
			Strike:              o.StrikePrice,
			OptionType:          string(o.OptionType),
			SettlementDate:      o.SettlementDate,
			ImpliedVolatility:   o.ImpliedVolatility,
			OpenInterest:        o.OpenInterest,
		}
	}

	return &dto.StockDetailedResponse{
		StockResponse: stockResp,
		History:       history,
		Options:       optionResponses,
	}, nil
}

// --- Futures ---

func (s *ListingService) GetFutures(ctx context.Context, q dto.ListingQuery) (*dto.PaginatedFuturesResponse, error) {
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

	return &dto.PaginatedFuturesResponse{
		Data:     data,
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
	}, nil
}

func (s *ListingService) GetFutureDetails(ctx context.Context, listingID uint) (*dto.FutureDetailedResponse, error) {
	l, err := s.listingRepo.FindByID(ctx, listingID)
	if err != nil || l == nil || l.ListingType != model.ListingTypeFuture {
		return nil, commonErrors.NotFoundErr("future not found")
	}

	fc, err := s.futuresRepo.FindByListingIDs(ctx, []uint{listingID})
	if err != nil || len(fc) == 0 {
		return nil, commonErrors.NotFoundErr("contract details not found")
	}

	return &dto.FutureDetailedResponse{
		FuturesResponse: dto.FuturesResponse{
			BaseListingResponse: baseResponse(*l, latestDaily(l.DailyPriceInfos)),
			SettlementDate:      fc[0].SettlementDate,
			ContractSize:        fc[0].ContractSize,
			ContractUnit:        fc[0].ContractUnit,
		},
		History: mapHistory(l.DailyPriceInfos),
	}, nil
}

// --- Forex ---

func (s *ListingService) GetForex(ctx context.Context, q dto.ListingQuery) (*dto.PaginatedForexResponse, error) {
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
		daily := latestDaily(p.Listing.DailyPriceInfos)
		base := baseResponse(p.Listing, daily)
		base.Ticker = p.Base + "/" + p.Quote
		base.Name = p.Base + "/" + p.Quote
		base.Exchange = p.Listing.ExchangeMIC
		base.Price = p.Rate

		data[i] = dto.ForexResponse{
			BaseListingResponse:  base,
			Base:                 p.Base,
			Quote:                p.Quote,
			ProviderUpdatedAt:    p.ProviderUpdatedAt,
			ProviderNextUpdateAt: p.ProviderNextUpdateAt,
		}
	}

	return &dto.PaginatedForexResponse{
		Data:     data,
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
	}, nil
}

func (s *ListingService) GetForexDetails(ctx context.Context, listingID uint) (*dto.ForexDetailedResponse, error) {
	l, err := s.listingRepo.FindByID(ctx, listingID)
	if err != nil || l == nil || l.ListingType != model.ListingTypeForexPair {
		return nil, commonErrors.NotFoundErr("forex not found")
	}

	pairs, err := s.forexRepo.FindByListingIDs(ctx, []uint{listingID})
	if err != nil || len(pairs) == 0 {
		return nil, commonErrors.NotFoundErr("forex pair details not found")
	}
	p := pairs[0]

	return &dto.ForexDetailedResponse{
		ForexResponse: dto.ForexResponse{
			BaseListingResponse:  baseResponse(*l, latestDaily(l.DailyPriceInfos)),
			Base:                 p.Base,
			Quote:                p.Quote,
			ProviderUpdatedAt:    p.ProviderUpdatedAt,
			ProviderNextUpdateAt: p.ProviderNextUpdateAt,
		},
		History: mapHistory(l.DailyPriceInfos),
	}, nil
}

func (s *ListingService) GetOptions(ctx context.Context, q dto.ListingQuery) (*dto.PaginatedOptionResponse, error) {
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

	result := &dto.PaginatedOptionResponse{
		Data:     data,
		Total:    total,
		Page:     q.Page,
		PageSize: q.PageSize,
	}
	return result, nil
}

func (s *ListingService) GetOptionDetails(ctx context.Context, listingID uint) (*dto.OptionDetailedResponse, error) {
	l, err := s.listingRepo.FindByID(ctx, listingID)
	if err != nil || l == nil || l.ListingType != model.ListingTypeOption {
		return nil, commonErrors.NotFoundErr("option not found")
	}

	opt, err := s.optionRepo.FindByListingIDs(ctx, []uint{listingID})
	if err != nil || len(opt) == 0 {
		return nil, commonErrors.NotFoundErr("option details not found")
	}

	return &dto.OptionDetailedResponse{
		OptionResponse: dto.OptionResponse{
			BaseListingResponse: baseResponse(*l, latestDaily(l.DailyPriceInfos)),
			Strike:              opt[0].StrikePrice,
			OptionType:          string(opt[0].OptionType),
			SettlementDate:      opt[0].SettlementDate,
			ImpliedVolatility:   opt[0].ImpliedVolatility,
			OpenInterest:        opt[0].OpenInterest,
		},
		History: mapHistory(l.DailyPriceInfos),
	}, nil
}
