/*
	Forked from https://github.com/mattn/go-sixel/

	altered for own purposes:
		- dropped decoder
		- dropped caching (only using a cached writer anyway)
		- dropped dithering (only using palleted images)
		- updated Encode() to return writer errors
*/

package sixel

import (
	"fmt"
	"image"
	"image/draw"
	"io"
)

// Encoder encode image to sixel format
type Encoder struct {
	w      io.Writer
	Width  int
	Height int
}

// NewEncoder return new instance of Encoder
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

const (
	specialChNr = byte(0x6d)
	specialChCr = byte(0x64)
)

// Encode do encoding
func (e *Encoder) Encode(pI *image.Paletted) (E error) {

	const nc = 255 // (>= 2, 8bit, index 0 is reserved for transparent key color)

	width, height := pI.Bounds().Dx(), pI.Bounds().Dy()
	if width == 0 || height == 0 {
		return
	}
	if e.Width > 0 {
		width = e.Width
	}
	if e.Height > 0 {
		height = e.Height
	}

	// make adaptive palette using median cut alogrithm
	draw.Draw(pI, pI.Bounds(), pI, image.ZP, draw.Over)

	// capture/report errors
	fnWri := func(v []byte) error {
		_, E = e.w.Write(v)
		return E
	}

	// DECSIXEL Introducer(\033P0;0;8q) + DECGRA ("1;1): Set Raster Attributes
	if fnWri([]byte{0x1b, 0x50, 0x30, 0x3b, 0x30, 0x3b, 0x38, 0x71, 0x22, 0x31, 0x3b, 0x31}) != nil {
		return
	}

	for n, v := range pI.Palette {

		r, g, b, _ := v.RGBA()
		r = r * 100 / 0xFFFF
		g = g * 100 / 0xFFFF
		b = b * 100 / 0xFFFF

		// DECGCI (#): Graphics Color Introducer
		if _, E = fmt.Fprintf(e.w, "#%d;2;%d;%d;%d", n+1, r, g, b); E != nil {
			return
		}
	}

	buf := make([]byte, width*nc)
	cset := make([]bool, nc)
	ch0 := specialChNr
	for z := 0; z < (height+5)/6; z++ {

		// DECGNL (-): Graphics Next Line
		if z > 0 {
			if fnWri([]byte{0x2d}) != nil {
				return
			}
		}

		for p := 0; p < 6; p++ {
			y := z*6 + p
			for x := 0; x < width; x++ {
				_, _, _, alpha := pI.At(x, y).RGBA()
				if alpha != 0 {
					idx := pI.ColorIndexAt(x, y) + 1
					cset[idx] = false // mark as used
					buf[width*int(idx)+x] |= 1 << uint(p)
				}
			}
		}

		for n := 1; n < nc; n++ {

			if cset[n] {
				continue
			}

			cset[n] = true

			// DECGCR ($): Graphics Carriage Return
			if ch0 == specialChCr {
				if fnWri([]byte{0x24}) != nil {
					return
				}
			}

			// select color (#%d)
			var tmp []byte
			if n >= 100 {
				digit1 := n / 100
				digit2 := (n - digit1*100) / 10
				digit3 := n % 10
				c1 := byte(0x30 + digit1)
				c2 := byte(0x30 + digit2)
				c3 := byte(0x30 + digit3)
				tmp = []byte{0x23, c1, c2, c3}
			} else if n >= 10 {
				c1 := byte(0x30 + n/10)
				c2 := byte(0x30 + n%10)
				tmp = []byte{0x23, c1, c2}
			} else {
				tmp = []byte{0x23, byte(0x30 + n)}
			}

			if (tmp != nil) && (fnWri(tmp) != nil) {
				return
			}

			cnt := 0
			for x := 0; x < width; x++ {

				// make sixel character from 6 pixels
				ch := buf[width*n+x]
				buf[width*n+x] = 0
				if ch0 < 0x40 && ch != ch0 {

					// output sixel character
					s := 63 + ch0
					for ; cnt > 255; cnt -= 255 {
						if fnWri([]byte{0x21, 0x32, 0x35, 0x35, s}) != nil {
							return
						}
					}

					var tmp []byte
					if cnt == 1 {
						tmp = []byte{s}
					} else if cnt == 2 {
						tmp = []byte{s, s}
					} else if cnt == 3 {
						tmp = []byte{s, s, s}
					} else if cnt >= 100 {

						// DECGRI (!): - Graphics Repeat Introducer
						digit1 := cnt / 100
						digit2 := (cnt - digit1*100) / 10
						digit3 := cnt % 10
						c1 := byte(0x30 + digit1)
						c2 := byte(0x30 + digit2)
						c3 := byte(0x30 + digit3)
						tmp = []byte{0x21, c1, c2, c3, s}

					} else if cnt >= 10 {

						// DECGRI (!): - Graphics Repeat Introducer
						c1 := byte(0x30 + cnt/10)
						c2 := byte(0x30 + cnt%10)
						tmp = []byte{0x21, c1, c2, s}

					} else if cnt > 0 {

						// DECGRI (!): - Graphics Repeat Introducer
						tmp = []byte{0x21, byte(0x30 + cnt), s}
					}

					if (tmp != nil) && (fnWri(tmp) != nil) {
						return
					}

					cnt = 0
				}
				ch0 = ch
				cnt++
			}

			// output sixel character
			if ch0 != 0 {
				s := 63 + ch0
				for ; cnt > 255; cnt -= 255 {
					if fnWri([]byte{0x21, 0x32, 0x35, 0x35, s}) != nil {
						return
					}
				}

				var tmp []byte
				if cnt == 1 {
					tmp = []byte{s}
				} else if cnt == 2 {
					tmp = []byte{s, s}
				} else if cnt == 3 {
					tmp = []byte{s, s, s}
				} else if cnt >= 100 {

					// DECGRI (!): - Graphics Repeat Introducer
					digit1 := cnt / 100
					digit2 := (cnt - digit1*100) / 10
					digit3 := cnt % 10
					c1 := byte(0x30 + digit1)
					c2 := byte(0x30 + digit2)
					c3 := byte(0x30 + digit3)
					tmp = []byte{0x21, c1, c2, c3, s}

				} else if cnt >= 10 {

					// DECGRI (!): - Graphics Repeat Introducer
					c1 := byte(0x30 + cnt/10)
					c2 := byte(0x30 + cnt%10)
					tmp = []byte{0x21, c1, c2, s}

				} else if cnt > 0 {

					// DECGRI (!): - Graphics Repeat Introducer
					tmp = []byte{0x21, byte(0x30 + cnt), s}
				}

				if (tmp != nil) && (fnWri(tmp) != nil) {
					return
				}
			}
			ch0 = specialChCr
		}
	}

	// string terminator(ST)
	fnWri([]byte{0x1b, 0x5c})
	return
}
