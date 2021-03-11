package quote

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var gPRT *message.Printer

func init() {

	// try to detect language from environment, fallback to en_US on fail
	langTag := language.AmericanEnglish

	if envlang, ok := os.LookupEnv("LANG"); ok {
		if P := strings.Split(envlang, "."); len(P) > 0 {
			if tag, e := language.Parse(P[0]); e == nil {
				langTag = tag
			}
		}
	}

	gPRT = message.NewPrinter(langTag)
}

type T_Time struct {
	time.Time
}

type T_Num float32

func (T T_Num) String(wid, dec int) string {

	layout := fmt.Sprintf("%%%d.0%df", wid, dec)
	return gPRT.Sprintf(layout, T)
}

func (T T_Time) String() string {

	if T.IsZero() {
		return "-"
	} else {
		return T.Format("01/02 15:04:05pm MST")
	}
}

type I_Chartable interface {
	Length() int
	Timestamp(int) int64
	Price(int) T_Num
}

type S_Price struct {
	Bid  T_Num
	Ask  T_Num
	Last T_Num
	Open T_Num
}

type S_Size struct {
	Bid         T_Num
	Ask         T_Num
	Last        T_Num
	TotalVolume T_Num
}

type S_Sample struct {
	Time  T_Time
	Price S_Price
	Size  S_Size
}

type S_SampleArray []S_Sample

func (V S_SampleArray) Length() int {
	return len(V)
}

func (V S_SampleArray) Timestamp(ix int) int64 {
	return V[ix].Time.Unix()
}

func (V S_SampleArray) Price(ix int) T_Num {
	return V[ix].Price.Last
}

type S_Time struct {
	Quote T_Time
	Trade T_Time
}

type S_Quote struct {
	Symbol      string
	Description string
	Exchange    string
	Time        S_Time
	Price       S_Price
	Size        S_Size
}

type S_Candle struct {
	Hi   T_Num
	Lo   T_Num
	Vol  T_Num
	Time T_Time
}

type S_CandleArray []S_Candle

func (V S_CandleArray) Length() int {
	return len(V)
}

func (V S_CandleArray) Timestamp(ix int) int64 {
	return V[ix].Time.Unix()
}

func (V S_CandleArray) Price(ix int) T_Num {
	pC := &V[ix]
	return (pC.Hi + pC.Lo) / 2
}

// CAST INTERFACE TO string
func ToStr(i interface{}) string {
	if s, ok := i.(string); ok {
		return s
	}
	return ""
}

// CONVERT INTERFACE TO T_Num
func ToNum(i interface{}) T_Num {

	f := math.NaN()

	switch v := i.(type) {
	case float64:
		f = v
	case int:
		f = float64(v)
	case string:
		fTmp, e := strconv.ParseFloat(v, 64)
		if e == nil {
			f = fTmp
		}
	}

	return T_Num(f)
}

// CONVERT INTERFACE (ms timestamp) TO T_Time
func ToTime(i interface{}) T_Time {

	var msec int64

	switch v := i.(type) {
	case float64:
		msec = int64(v)
	case int64:
		msec = v
	case uint64:
		msec = int64(v)
	}

	if msec > 0 {
		return T_Time{time.Unix(msec/1000, 0)}
	}

	return T_Time{}
}

type Bucket struct {
	Sum     T_Num
	Samples uint
}

func (B *Bucket) AddSample(v T_Num) {
	B.Sum += v
	B.Samples += 1
}

func (B Bucket) Val() T_Num {
	if B.Samples > 0 {
		return B.Sum / T_Num(B.Samples)
	}
	return 0
}

type S_Buckets struct {
	Bkts       []Bucket
	TMin, TMax int64
	PMin, PMax T_Num
}

func ToBuckets(hist I_Chartable, nB int) (RET S_Buckets) {

	nSamp := hist.Length()
	if (nSamp < 2) || (nB < 1) {
		return
	}

	// TRACK TIME RANGE (ASSUMES SAMPLES ARE IN CHRONOLOGICAL ORDER)
	RET.TMin = hist.Timestamp(0)
	RET.TMax = hist.Timestamp(nSamp - 1)
	dT := RET.TMax - RET.TMin
	if dT <= 0 {
		return
	}

	RET.Bkts = make([]Bucket, nB)

	rngInv := 1.0 / float64(dT)
	RET.PMin = hist.Price(0)
	RET.PMax = RET.PMin

	for ix_samp := 0; ix_samp < nSamp; ix_samp++ {

		// CALC BUCKET INDEX
		ix_bucket := int(float64((hist.Timestamp(ix_samp)-RET.TMin)*int64(nB)) * rngInv)
		if ix_bucket >= nB {
			ix_bucket = nB - 1
		}

		// TRACK PRICE RANGE
		nPrice := hist.Price(ix_samp)
		if nPrice > RET.PMax {
			RET.PMax = nPrice
		} else if nPrice < RET.PMin {
			RET.PMin = nPrice
		}

		RET.Bkts[ix_bucket].AddSample(nPrice)
	}

	return
}
