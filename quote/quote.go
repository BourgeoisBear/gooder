package quote

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

const COL_WIDTH = 15
const ESC_SGR_RESET string = "\x1b[0m"

type Quoter interface {
	GetQuotes(*http.Client, []string) ([]S_Quote, error)
}

type Obj map[string]interface{}

func HeaderRow() []string {

	sL := fmt.Sprintf("%%-%ds", COL_WIDTH)

	return []string{
		fmt.Sprintf(sL, "Volume"),
		fmt.Sprintf(sL, "Bid"),
		fmt.Sprintf(sL, "Ask"),
		fmt.Sprintf(sL, "Last"),
	}
}

func (Q *S_Quote) Display(out io.Writer, hist S_SampleArray) error {

	var delta S_Price
	var volumeDelta T_Num

	if histLen := len(hist); histLen > 1 {

		cur := hist[histLen-1]
		prev := hist[histLen-2]
		delta = S_Price{
			Bid:  cur.Price.Bid - prev.Price.Bid,
			Ask:  cur.Price.Ask - prev.Price.Ask,
			Last: cur.Price.Last - prev.Price.Last,
		}

		volumeDelta = cur.Size.TotalVolume - prev.Size.TotalVolume
	}

	P := Q.Price

	colorDelta := func(cur, delta T_Num) string {

		var pfx, suf string
		nWid := COL_WIDTH

		if delta > 0 {
			pfx = "\x1b[38;2;0;255;0m ▲"
			suf = ESC_SGR_RESET
			nWid -= 2
		} else if delta < 0 {
			pfx = "\x1b[38;2;255;0;0m ▼"
			suf = ESC_SGR_RESET
			nWid -= 2
		}

		return pfx + cur.String(nWid, 2) + suf
	}

	// `locale` program
	// LC_MONETARY, LC_NUMERIC, LC_TIME
	sPrice := []string{
		Q.Size.TotalVolume.String(COL_WIDTH, 0),
		colorDelta(P.Bid, delta.Bid),
		colorDelta(P.Ask, delta.Ask),
		colorDelta(P.Last, delta.Last),
	}

	sDelta := []string{
		volumeDelta.String(COL_WIDTH, 0),
		colorDelta(delta.Bid, delta.Bid),
		colorDelta(delta.Ask, delta.Ask),
		colorDelta(delta.Last, delta.Last),
	}

	V := []string{
		Q.Exchange + ":" + Q.Symbol + " - " + Q.Description,
		"\x1b[36m" + "TRD: " + Q.Time.Trade.String() + ", " + "QTD: " + Q.Time.Quote.String() + ESC_SGR_RESET,
		strings.Join(HeaderRow(), " "),
		strings.Join(sPrice, " "),
		strings.Join(sDelta, " "),
	}

	_, E := out.Write([]byte(strings.Join(V, "\n")))
	return E
}
