# GOODER

*Graph goes up means world more gooder*

### TODOs

## Release
	- GODOC comments
	- rasterm documentation
	- build/install instructions
	- screenshots

## Terminal
	- improve terminal identification
	- check that mintty supports iterm/wezterm format, get mintty identifier

## Meat & Potatoes
	- append streaming prices to candles
	- chart label placement
	- configurable time range
	- query/sleep selection for different assets
	- user command to specify sixel, kitty, or wez
	- alternate data sources / api selector (ameritrade, yahoo, google?, iex)
		https://github.com/ranaroussi/yfinance
		https://iexcloud.io/docs/api/#quote
		https://developer.tdameritrade.com/quotes/apis/get/marketdata/%7Bsymbol%7D/quotes
		https://live.euronext.com/en/intraday_chart/getDetailedQuoteAjax/FR0000031122-XPAR/full
		https://finnhub.io/

### TESTING
- test sixel with
	- https://github.com/liamg/aminal
	- https://domterm.org/
	- https://www.macterm.net/
- test wez/iterm img with
	- https://www.macterm.net/

### MARKET HOURS

	DEFS:
		BID: highest buyer will pay
		ASK: lowest seller will accept

	MARKET HOURS:
		pre
			NYSE:	6:30a-9:30a
			NASD: 4:00a-9:30a
		reg
			9:30a-4:00p EST
		post
			NYSE: 4:00p-8:00p
			NASD: 4:00p-8:00p

### Supported Image Encodings

- *WezTerm & iTerm2*: https://iterm2.com/documentation-images.html
- *Kitty*: https://sw.kovidgoyal.net/kitty/graphics-protocol.html
- *Sixel*: https://saitoha.github.io/libsixel/

### Terminal requirements

Per https://saitoha.github.io/libsixel/:

To view the charts, use a terminal that supports the SIXEL image format:

Now SIXEL feature is supported by the following terminals:

- DEC VT series, VT240/VT241/VT330/VT340/VT282/VT284/VT286/VT382
- DECterm(dxterm)
- Kermit
- ZSTEM 340
- WRQ Reflection
- RLogin (Japanese terminal emulator) http://nanno.dip.jp/softlib/man/rlogin/
- mlterm http://mlterm.sourceforge.net/
- yaft https://github.com/uobikiemukot/yaft
- Mintty (>= 2.6.0) https://mintty.github.io/
- cancer https://github.com/meh/cancer/
- XTerm (compiled with --enable-sixel-graphics option)

### XTerm Configuration

```
	http://invisible-island.net/xterm/
	You should launch xterm with “-ti vt340” option. The SIXEL palette is limited to a maximum of 16 colors. To avoid this limitation, Try

$ echo "XTerm*decTerminalID: vt340" >> $HOME/.Xresources
$ echo "XTerm*numColorRegisters: 256" >>  $HOME/.Xresources
$ xrdb $HOME/.Xresources
$ xterm
or

$ xterm -xrm "XTerm*decTerminalID: vt340" -xrm "XTerm*numColorRegisters: 256"
```
