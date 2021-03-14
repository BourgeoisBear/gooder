package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/BourgeoisBear/gooder/quote"
)

var HTTPTransport *http.Transport = &http.Transport{
	MaxIdleConns:          10,
	IdleConnTimeout:       90 * time.Second,
	ResponseHeaderTimeout: 90 * time.Second,
	TLSHandshakeTimeout:   90 * time.Second,
	TLSClientConfig: &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS12,
	},
}

var HTTPClient *http.Client = &http.Client{
	Transport: HTTPTransport,
	Timeout:   15 * time.Second,
}

var API_KEY = "867EZZQOI92UGHPW4RHHBP9F64ANOPGA"
var ESC_ERASE_DISPLAY = "\x1b[2J\x1b[0;0H"

const TEST_CANDLES_FNAME = "test_candles.json"
const N_BUCKETS = 96

func main() {

	var SYMBOLS string
	var POLL_SECONDS int
	var QUICKCHART bool

	flag.StringVar(&SYMBOLS, "s", "gme", "Ticker symbols, comma separated")
	flag.IntVar(&POLL_SECONDS, "p", 5, "Poll interval, in seconds")
	flag.BoolVar(&QUICKCHART, "c", false, "Quick Chart Mode")
	flag.Parse()

	API := quote.ApiAmeritrade{KEY: API_KEY}

	sSYMS := strings.Split(SYMBOLS, ",")
	for ix := range sSYMS {
		sSYMS[ix] = strings.ToUpper(strings.TrimSpace(sSYMS[ix]))
	}

	HIST := make(map[string]quote.S_SampleArray)

	bufWri := bufio.NewWriterSize(os.Stdout, 128*1024)
	logE := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)
	fnLog := func(cls string, e error) {
		if e != nil {
			s := fmt.Sprintf("[E:%s] %s", cls, e.Error())
			logE.Output(2, s)
		}
	}

	if QUICKCHART {

		var eQ error

		defer func() {
			if eQ != nil {
				fnLog("QUICKCHART", eQ)
			}
		}()

		var sC quote.S_CandleArray
		var bs []byte
		const bDebug = true
		const bSaveCandles = false

		if _, eQ = bufWri.WriteString(ESC_ERASE_DISPLAY); eQ != nil {
			return
		}

		if bDebug {

			fTestData, e := FsAssets.Open("assets/" + TEST_CANDLES_FNAME)
			if e != nil {
				eQ = e
				return
			}
			defer fTestData.Close()

			if bs, eQ = ioutil.ReadAll(fTestData); eQ != nil {
				return
			}

			if eQ = json.Unmarshal(bs, &sC); eQ != nil {
				return
			}

			// NOTE: for testing empty buckets
			// sC = append(sC[:200], sC[450:]...)

		} else {

			// 4a - 8p
			tNow := time.Now()
			var newYork *time.Location
			if newYork, eQ = time.LoadLocation("America/New_York"); eQ != nil {
				return
			}

			tMin := time.Date(tNow.Year(), tNow.Month(), tNow.Day(), 4, 0, 0, 0, newYork)
			tMax := time.Date(tNow.Year(), tNow.Month(), tNow.Day(), 20, 0, 0, 0, newYork)

			if sC, eQ = API.GetPriceHistory(HTTPClient, "GME", tMin, tMax); eQ != nil {
				return
			}

			if bs, eQ = json.Marshal(sC); eQ != nil {
				return
			}

			if bSaveCandles {
				if eQ = ioutil.WriteFile("./"+TEST_CANDLES_FNAME, bs, 0644); eQ != nil {
					return
				}
			}
		}

		BKTS := quote.ToBuckets(sC, N_BUCKETS)
		if eQ = WriteChart(bufWri, BKTS); eQ != nil {
			return
		}

		if eQ = bufWri.Flush(); eQ != nil {
			return
		}

		return
	}

	for {

		sQ, e1 := API.GetQuotes(HTTPClient, sSYMS)
		if e1 != nil {
			fnLog("PULL", e1)
			goto DO_SLEEP
		}

		// TODO: get terminal height for [{n}A

		if _, e1 := bufWri.WriteString(ESC_ERASE_DISPLAY); e1 != nil {
			fnLog("IO", e1)
			goto DO_SLEEP
		}

		for _, sym := range sSYMS {

			// LOOKUP SYMBOL IN QUOTE LIST (TO PRESERVE USER-SPECIFIED SYMBOL ORDER)
			ix := -1
			for j := range sQ {
				if sQ[j].Symbol == sym {
					ix = j
					break
				}
			}

			if ix < 0 {
				continue
			}

			pQuote := &(sQ[ix])
			symH := HIST[pQuote.Symbol]

			// ONLY APPEND TO HISTORY WHEN QUOTE TIME ADVANCES
			if nLen := len(symH); (nLen == 0) || pQuote.Time.Quote.After(symH[nLen-1].Time.Time) {

				symH = append(symH, quote.S_Sample{
					Time:  pQuote.Time.Quote,
					Price: pQuote.Price,
					Size:  pQuote.Size,
				})

				HIST[pQuote.Symbol] = symH
			}

			if e2 := pQuote.Display(bufWri, symH); e2 != nil {
				fnLog("TXT", e2)
				goto DO_SLEEP
			}

			if _, e2 := bufWri.WriteString("\n\n"); e2 != nil {
				fnLog("IO", e2)
				goto DO_SLEEP
			}

			BKTS := quote.ToBuckets(symH, N_BUCKETS)
			if e2 := WriteChart(bufWri, BKTS); e2 != nil {
				fnLog("SIXEL", e2)
				goto DO_SLEEP
			}

			if _, e2 := bufWri.WriteString("\n"); e2 != nil {
				fnLog("IO", e2)
				goto DO_SLEEP
			}
		}

		if e1 := bufWri.Flush(); e1 != nil {
			fnLog("IO", e1)
			goto DO_SLEEP
		}

	DO_SLEEP:

		time.Sleep(time.Duration(POLL_SECONDS) * time.Second)
	}
}
