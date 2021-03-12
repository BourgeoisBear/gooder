package main

import (
	"embed"
	_ "embed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"strings"

	"github.com/BourgeoisBear/gooder/quote"
	"github.com/BourgeoisBear/gooder/sixel"
	"github.com/BourgeoisBear/gooder/wuline"
)

const N_WID = 350
const N_HGT = 75

var gIMG *image.Paletted
var gIMG_CLEAR_PIX []uint8
var IX_GREEN, IX_RED, N_GREEN, N_RED int

var gFONT image.Image
var CHR_W, CHR_H, CHRS_PER_ROW int

//go:embed fonts/mono.png
var FsAssets embed.FS

func init() {

	P := color.Palette{
		color.NRGBA{0, 0, 0, 255},
		color.NRGBA{255, 255, 255, 255}, // WHITE
	}

	IX_RED = len(P)
	P = append(
		P,
		color.NRGBA{128, 0, 0, 255},
		color.NRGBA{192, 0, 0, 255},
		color.NRGBA{255, 0, 0, 255},
	)
	N_RED = len(P) - IX_RED

	IX_GREEN = len(P)
	P = append(
		P,
		color.NRGBA{5, 102, 68, 255},
		color.NRGBA{5, 172, 114, 255},
		color.NRGBA{10, 212, 139, 255},
		color.NRGBA{59, 254, 184, 255},
	)
	N_GREEN = len(P) - IX_GREEN

	// TEMP IMAGE (for chart)
	gIMG = image.NewPaletted(image.Rect(0, 0, N_WID, N_HGT), P)
	gIMG_CLEAR_PIX = make([]uint8, len(gIMG.Pix))

	// FONT
	fFont, e := FsAssets.Open("fonts/mono.png")
	if e != nil {
		panic(e)
	}

	gFONT, e = png.Decode(fFont)
	if e != nil {
		panic(e)
	}

	bnds := gFONT.Bounds()
	CHRS_PER_ROW = 32
	CHR_W = bnds.Max.X / CHRS_PER_ROW
	CHR_H = bnds.Max.Y / 3
	//fmt.Printf("%#v, %d, %d", bnds, CHR_W, CHR_H)
}

func SixelChart(out io.Writer, B quote.S_Buckets) error {

	pEnc := sixel.NewEncoder(out)

	// CLEAR BITMAP
	copy(gIMG.Pix, gIMG_CLEAR_PIX)

	// DRAW AXES
	for i := 0; i < N_HGT; i++ {
		gIMG.SetColorIndex(0, i, 1)
	}

	for i := 0; i < N_WID; i++ {
		gIMG.SetColorIndex(i, N_HGT-1, 1)
	}

	// IMG DIMS (MINUS AXES WIDTHS)
	dx, dy := gIMG.Rect.Dx()-1, gIMG.Rect.Dy()-1

	// CALC PRICESCALE [Y]
	YPad := int(0.15 * float32(dy))
	dP := B.PMax - B.PMin
	if dP < 0 {
		dP = -dP
	}
	sP := quote.T_Num(dy-YPad) / dP

	priceY := func(ix int) (float32, bool) {

		if bkt := &B.Bkts[ix]; bkt.Samples > 0 {
			return float32(
				quote.T_Num(dy) - ((bkt.Val() - B.PMin) * sP) - (quote.T_Num(YPad) / 2.0),
			), true
		}

		return -1, false
	}

	var sT float32 = float32(dx) / float32(len(B.Bkts))

	price_prev, prev_ok := priceY(0)
	var t_prev float32 = 1
	fGREENS := float32(N_GREEN)

	fnPlot := func(x, y int, bright float32) {

		c_ix := int((bright * fGREENS) + 0.5)
		if c_ix >= N_GREEN {
			c_ix = N_GREEN - 1
		}

		gIMG.SetColorIndex(x, y, uint8(IX_GREEN+c_ix))
	}

	// PLOT BUCKETS
	for ix := range B.Bkts {

		price, price_ok := priceY(ix)
		t_cur := t_prev + sT

		// LEAVE BLANK FOR EMPTY BUCKETS
		if price_ok {

			if !prev_ok {
				price_prev = price
			}

			wuline.Line(t_prev, price_prev, t_cur, price, fnPlot)
		}

		price_prev, prev_ok, t_prev = price, price_ok, t_cur
	}

	/*
		TODO:
			- label placement
			- embed test candles
			- blank time range for buckets (for active trading day)
			- only last 8hrs
	*/

	pad := 9
	x_txt, y_txt := 2, 3
	_, h := Text(gIMG, B.PMax.String(pad, 2), x_txt, y_txt)
	Text(gIMG, B.PMin.String(pad, 2), x_txt, y_txt+h)

	return pEnc.Encode(gIMG)
}

func Text(iImg image.Image, txt string, x, y int) (width, height int) {

	bs := []byte(strings.ToUpper(txt))
	ptDst := image.Point{x, y}

	for _, chr := range bs {

		if (chr >= 32) && (chr <= 126) {

			ccol := int(chr-32) % CHRS_PER_ROW
			crow := int(chr-32) / CHRS_PER_ROW

			draw.Draw(
				gIMG,
				image.Rectangle{
					Min: ptDst,
					Max: image.Point{X: ptDst.X + CHR_W, Y: ptDst.Y + CHR_H},
				},
				gFONT,
				image.Point{X: ccol * CHR_W, Y: crow * CHR_H},
				draw.Over,
			)
		}

		ptDst.X += CHR_W
	}

	return CHR_W * len(bs), CHR_H
}
