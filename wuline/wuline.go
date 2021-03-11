package wuline

type PlotFunc func(x, y int, bright float32)

func ipart(x float32) float32 {
	return float32(int(x))
}

func round(x float32) float32 {
	return ipart(x + 0.5)
}

func fpart(x float32) float32 {
	return x - ipart(x)
}

func rfpart(x float32) float32 {
	return 1.0 - fpart(x)
}

func swap(a, b *float32) {
	tmp := *a
	*a = *b
	*b = tmp
}

func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

func Line(x0, y0, x1, y1 float32, usrPlot PlotFunc) {

	fnPlot := func(x, y, bright float32) {
		usrPlot(int(x+0.5), int(y+0.5), bright)
	}

	steep := abs(y1-y0) > abs(x1-x0)

	if steep {
		swap(&x0, &y0)
		swap(&x1, &y1)
	}

	if x0 > x1 {
		swap(&x0, &x1)
		swap(&y0, &y1)
	}

	dx := x1 - x0
	dy := y1 - y0
	slope := dy / dx
	if dx == 0.0 {
		slope = 1.0
	}

	PLT := func(x, y, xgap, yend float32) {

		// FIXED BRIGHTNESS FOR ENDCAPS
		var b1, b2 float32 = 0.4, 0.4

		// b1, b2 := rfpart(yend)*xgap, fpart(yend)*xgap
		_, _ = xgap, yend

		if steep {

			fnPlot(y, x, b1)
			fnPlot(y+1, x, b2)

		} else {

			fnPlot(x, y, b1)
			fnPlot(x, y+1, b2)
		}
	}

	// ENDPOINT 0
	xend := round(x0)
	yend := y0 + slope*(xend-x0)
	xgap := rfpart(x0 + 0.5)
	xpxl1 := xend
	ypxl1 := ipart(yend)
	y_isect := yend + slope
	PLT(xpxl1, ypxl1, xgap, yend)

	// ENDPOINT 1
	xend = round(x1)
	yend = y1 + slope*(xend-x1)
	xgap = fpart(x1 + 0.5)
	xpxl2 := xend
	ypxl2 := ipart(yend)
	PLT(xpxl2, ypxl2, xgap, yend)

	if steep {

		for x := xpxl1 + 1; x <= (xpxl2 - 1); x++ {
			fnPlot(ipart(y_isect), x, rfpart(y_isect))
			fnPlot(ipart(y_isect)+1, x, fpart(y_isect))
			y_isect = y_isect + slope
		}

	} else {

		for x := xpxl1 + 1; x <= (xpxl2 - 1); x++ {
			fnPlot(x, ipart(y_isect), rfpart(y_isect))
			fnPlot(x, ipart(y_isect)+1, fpart(y_isect))
			y_isect = y_isect + slope
		}
	}
}
