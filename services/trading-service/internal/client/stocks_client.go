package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const finnhubBaseURL = "https://finnhub.io/api/v1"

type Symbol struct {
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	MIC         string `json:"mic"`
}

type Profile struct {
	Name             string  `json:"name"`
	Exchange         string  `json:"exchange"`
	ShareOutstanding float64 `json:"shareOutstanding"`
	Ticker           string  `json:"ticker"`
}

type Quote struct {
	CurrentPrice float64 `json:"c"`
	High         float64 `json:"h"`
	Change       float64 `json:"d"`
	Volume       float64 `json:"v"`
}

type BasicFinancials struct {
	Metric struct {
		DividendYieldIndicatedAnnual float64 `json:"dividendYieldIndicatedAnnual"`
	} `json:"metric"`
}

type StockClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewStockClient(apiKey string) *StockClient {
	return &StockClient{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func (c *StockClient) get(path string, out interface{}) error {
	separator := "?"
	for _, ch := range path {
		if ch == '?' {
			separator = "&"
			break
		}
	}
	url := fmt.Sprintf("%s%s%stoken=%s", finnhubBaseURL, path, separator, c.apiKey)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("finnhub returned status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *StockClient) GetSymbols(exchange string) ([]Symbol, error) {
	var symbols []Symbol
	if err := c.get(fmt.Sprintf("/stock/symbol?exchange=%s", exchange), &symbols); err != nil {
		return nil, err
	}
	return symbols, nil
}

func (c *StockClient) GetProfile(ticker string) (*Profile, error) {
	var p Profile
	if err := c.get(fmt.Sprintf("/stock/profile2?symbol=%s", ticker), &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *StockClient) GetQuote(ticker string) (*Quote, error) {
	var q Quote
	if err := c.get(fmt.Sprintf("/quote?symbol=%s", ticker), &q); err != nil {
		return nil, err
	}
	return &q, nil
}

func (c *StockClient) GetBasicFinancials(ticker string) (*BasicFinancials, error) {
	var f BasicFinancials
	if err := c.get(fmt.Sprintf("/stock/metric?symbol=%s&metric=all", ticker), &f); err != nil {
		return nil, err
	}
	return &f, nil
}

type OptionContract struct {
	ContractName      string  `json:"contractName"`
	ContractSize      string  `json:"contractSize"`
	Currency          string  `json:"currency"`
	Type              string  `json:"type"`
	InTheMoney        bool    `json:"inTheMoney"`
	Strike            float64 `json:"strike"`
	LastPrice         float64 `json:"lastPrice"`
	Bid               float64 `json:"bid"`
	Ask               float64 `json:"ask"`
	Change            float64 `json:"change"`
	ChangePercent     float64 `json:"changePercent"`
	Volume            int     `json:"volume"`
	OpenInterest      int     `json:"openInterest"`
	ImpliedVolatility float64 `json:"impliedVolatility"`
}

type OptionChainExpiration struct {
	ExpirationDate string `json:"expirationDate"`
	Options        struct {
		Call []OptionContract `json:"CALL"`
		Put  []OptionContract `json:"PUT"`
	} `json:"options"`
}

type OptionChainResponse struct {
	Code string                  `json:"code"`
	Data []OptionChainExpiration `json:"data"`
}

func (c *StockClient) GetOptionChain(symbol string) (*OptionChainResponse, error) {
	var resp OptionChainResponse
	if err := c.get(fmt.Sprintf("/stock/option-chain?symbol=%s", symbol), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

const yahooBaseURL = "https://query1.finance.yahoo.com/v7/finance/options"

type YahooOptionContract struct {
	ContractSymbol    string  `json:"contractSymbol"`
	Strike            float64 `json:"strike"`
	LastPrice         float64 `json:"lastPrice"`
	Bid               float64 `json:"bid"`
	Ask               float64 `json:"ask"`
	Change            float64 `json:"change"`
	PercentChange     float64 `json:"percentChange"`
	Volume            int     `json:"volume"`
	OpenInterest      int     `json:"openInterest"`
	ImpliedVolatility float64 `json:"impliedVolatility"`
	Expiration        int64   `json:"expiration"`
	ContractSize      string  `json:"contractSize"`
}

type YahooOptionChain struct {
	OptionChain struct {
		Result []struct {
			UnderlyingSymbol string  `json:"underlyingSymbol"`
			ExpirationDates  []int64 `json:"expirationDates"`
			Options          []struct {
				ExpirationDate int64                 `json:"expirationDate"`
				Calls          []YahooOptionContract `json:"calls"`
				Puts           []YahooOptionContract `json:"puts"`
			} `json:"options"`
		} `json:"result"`
	} `json:"optionChain"`
}

func (c *StockClient) GetOptionChainYahoo(symbol string) (*YahooOptionChain, error) {
	url := fmt.Sprintf("%s/%s", yahooBaseURL, symbol)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yahoo returned status %d", resp.StatusCode)
	}

	var chain YahooOptionChain
	if err := json.NewDecoder(resp.Body).Decode(&chain); err != nil {
		return nil, err
	}
	return &chain, nil
}
