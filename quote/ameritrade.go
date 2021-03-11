package quote

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const AMERITRADE_URL_BASE = "https://api.tdameritrade.com/v1/marketdata"

type ApiAmeritrade struct {
	KEY string
}

func (SRC Obj) ToQuote() S_Quote {

	STR := func(key string) string { return ToStr(SRC[key]) }
	NUM := func(key string) T_Num { return ToNum(SRC[key]) }
	TIME := func(key string) T_Time { return ToTime(SRC[key]) }

	return S_Quote{
		Exchange:    strings.ToUpper(STR("exchangeName")),
		Symbol:      strings.ToUpper(STR("symbol")),
		Description: STR("description"),
		Time: S_Time{
			Quote: TIME("quoteTimeInLong"),
			Trade: TIME("tradeTimeInLong"),
		},
		Price: S_Price{
			Bid:  NUM("bidPrice"),
			Ask:  NUM("askPrice"),
			Last: NUM("lastPrice"),
			Open: NUM("openPrice"),
		},
		Size: S_Size{
			Bid:         NUM("bidSize"),
			Ask:         NUM("askSize"),
			Last:        NUM("lastSize"),
			TotalVolume: NUM("totalVolume"),
		},
	}
}

type ameritradeCandle struct {
	High, Low, Volume, Datetime float64
}

func (vC *ameritradeCandle) ToCandle() S_Candle {
	return S_Candle{
		Hi:   ToNum(vC.High),
		Lo:   ToNum(vC.Low),
		Vol:  ToNum(vC.Volume),
		Time: ToTime(vC.Datetime),
	}
}

// TODO: authenticated mode
// TODO: price history call

func (API *ApiAmeritrade) GetPriceHistory(pCli *http.Client, sym string, tMin, tMax time.Time) (sC S_CandleArray, E error) {

	sym = strings.ToUpper(strings.TrimSpace(sym))

	sURL := []string{
		"apikey=" + url.QueryEscape(API.KEY),
		"frequencyType=minute",
		"needExtendedHoursData=true",
		"startDate=" + fmt.Sprintf("%d", tMin.Unix()*1000),
		"endDate=" + fmt.Sprintf("%d", tMax.Unix()*1000),
	}

	url := AMERITRADE_URL_BASE + "/" + sym + "/pricehistory?" + strings.Join(sURL, "&")
	RSP, E := pCli.Get(url)
	if E != nil {
		return
	}
	defer RSP.Body.Close()

	if RSP.StatusCode != 200 {
		E = errors.New(RSP.Status)
		return
	}

	var DAT struct {
		Candles []ameritradeCandle `json:"candles"`
	}
	pDec := json.NewDecoder(RSP.Body)
	E = pDec.Decode(&DAT)
	if E != nil {
		return
	}

	sC = make([]S_Candle, len(DAT.Candles))
	for ix := range DAT.Candles {
		sC[ix] = DAT.Candles[ix].ToCandle()
	}

	return
}

func (API *ApiAmeritrade) GetQuotes(pCli *http.Client, syms []string) (sQ []S_Quote, E error) {

	sURL := []string{
		"apikey=" + url.QueryEscape(API.KEY),
		"symbol=" + strings.Join(syms, ","),
	}

	RSP, E := pCli.Get(AMERITRADE_URL_BASE + "/quotes?" + strings.Join(sURL, "&"))
	if E != nil {
		return
	}
	defer RSP.Body.Close()

	if RSP.StatusCode != 200 {
		E = errors.New(RSP.Status)
		return
	}

	pDec := json.NewDecoder(RSP.Body)
	var DAT map[string]Obj
	E = pDec.Decode(&DAT)
	if E != nil {
		return
	}

	sQ = make([]S_Quote, 0, len(DAT))
	for _, vObj := range DAT {
		sQ = append(sQ, vObj.ToQuote())
	}

	return
}
