package quote

import (
	"fmt"
	"strings"

	c "github.com/achannarasappa/ticker/internal/common"
	"github.com/achannarasappa/ticker/internal/currency"
	"github.com/go-resty/resty/v2"
)

type ResponseQuote struct {
	ShortName                  string  `json:"shortName"`
	Symbol                     string  `json:"symbol"`
	MarketState                string  `json:"marketState"`
	Currency                   string  `json:"currency"`
	ExchangeName               string  `json:"fullExchangeName"`
	ExchangeDelay              float64 `json:"exchangeDataDelayedBy"`
	RegularMarketChange        float64 `json:"regularMarketChange"`
	RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
	RegularMarketPrice         float64 `json:"regularMarketPrice"`
	RegularMarketPreviousClose float64 `json:"regularMarketPreviousClose"`
	RegularMarketOpen          float64 `json:"regularMarketOpen"`
	RegularMarketDayRange      string  `json:"regularMarketDayRange"`
	PostMarketChange           float64 `json:"postMarketChange"`
	PostMarketChangePercent    float64 `json:"postMarketChangePercent"`
	PostMarketPrice            float64 `json:"postMarketPrice"`
	PreMarketChange            float64 `json:"preMarketChange"`
	PreMarketChangePercent     float64 `json:"preMarketChangePercent"`
	PreMarketPrice             float64 `json:"preMarketPrice"`
}

type Quote struct {
	ResponseQuote
	Price                   float64
	Change                  float64
	ChangePercent           float64
	IsActive                bool
	IsRegularTradingSession bool
	CurrencyConverted       string
}

type Response struct {
	QuoteResponse struct {
		Quotes []ResponseQuote `json:"result"`
		Error  interface{}     `json:"error"`
	} `json:"quoteResponse"`
}

func transformResponseQuote(ctx c.Context, responseQuote ResponseQuote) Quote {

	currencyRate, currencyCode := currency.GetCurrencyRateFromContext(ctx, responseQuote.Currency)

	if responseQuote.MarketState == "REGULAR" {
		return Quote{
			ResponseQuote:           responseQuote,
			Price:                   responseQuote.RegularMarketPrice * currencyRate,
			Change:                  (responseQuote.RegularMarketChange) * currencyRate,
			ChangePercent:           responseQuote.RegularMarketChangePercent,
			IsActive:                true,
			IsRegularTradingSession: true,
			CurrencyConverted:       currencyCode,
		}
	}

	if responseQuote.MarketState == "POST" && responseQuote.PostMarketPrice == 0.0 {
		return Quote{
			ResponseQuote:           responseQuote,
			Price:                   responseQuote.RegularMarketPrice * currencyRate,
			Change:                  (responseQuote.RegularMarketChange) * currencyRate,
			ChangePercent:           responseQuote.RegularMarketChangePercent,
			IsActive:                true,
			IsRegularTradingSession: false,
			CurrencyConverted:       currencyCode,
		}
	}

	if responseQuote.MarketState == "PRE" && responseQuote.PreMarketPrice == 0.0 {
		return Quote{
			ResponseQuote:           responseQuote,
			Price:                   responseQuote.RegularMarketPrice * currencyRate,
			Change:                  (responseQuote.RegularMarketChange) * currencyRate,
			ChangePercent:           responseQuote.RegularMarketChangePercent,
			IsActive:                false,
			IsRegularTradingSession: false,
			CurrencyConverted:       currencyCode,
		}
	}

	if responseQuote.MarketState == "POST" {
		return Quote{
			ResponseQuote:           responseQuote,
			Price:                   responseQuote.PostMarketPrice * currencyRate,
			Change:                  (responseQuote.PostMarketChange + responseQuote.RegularMarketChange) * currencyRate,
			ChangePercent:           responseQuote.PostMarketChangePercent + responseQuote.RegularMarketChangePercent,
			IsActive:                true,
			IsRegularTradingSession: false,
			CurrencyConverted:       currencyCode,
		}
	}

	if responseQuote.MarketState == "PRE" {
		return Quote{
			ResponseQuote:           responseQuote,
			Price:                   responseQuote.PreMarketPrice * currencyRate,
			Change:                  (responseQuote.PreMarketChange) * currencyRate,
			ChangePercent:           responseQuote.PreMarketChangePercent,
			IsActive:                true,
			IsRegularTradingSession: false,
			CurrencyConverted:       currencyCode,
		}
	}

	if responseQuote.PostMarketPrice != 0.0 {
		return Quote{
			ResponseQuote:           responseQuote,
			Price:                   responseQuote.PostMarketPrice * currencyRate,
			Change:                  (responseQuote.PostMarketChange + responseQuote.RegularMarketChange) * currencyRate,
			ChangePercent:           responseQuote.PostMarketChangePercent + responseQuote.RegularMarketChangePercent,
			IsActive:                false,
			IsRegularTradingSession: false,
			CurrencyConverted:       currencyCode,
		}
	}

	return Quote{
		ResponseQuote:           responseQuote,
		Price:                   responseQuote.RegularMarketPrice * currencyRate,
		Change:                  (responseQuote.RegularMarketChange) * currencyRate,
		ChangePercent:           responseQuote.RegularMarketChangePercent,
		IsActive:                false,
		IsRegularTradingSession: false,
		CurrencyConverted:       currencyCode,
	}

}

func transformResponseQuotes(ctx c.Context, responseQuotes []ResponseQuote) []Quote {

	quotes := make([]Quote, 0)
	for _, responseQuote := range responseQuotes {
		quotes = append(quotes, transformResponseQuote(ctx, responseQuote))
	}
	return quotes

}

func GetQuotes(ctx c.Context, client resty.Client, symbols []string) func() []Quote {
	return func() []Quote {
		symbolsString := strings.Join(symbols, ",")
		url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&symbols=%s", symbolsString)
		res, _ := client.R().
			SetResult(Response{}).
			Get(url)

		return transformResponseQuotes(ctx, (res.Result().(*Response)).QuoteResponse.Quotes)
	}
}
