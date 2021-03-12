# GOODER

*Graph goes up means world more gooder.*

### TODOs

- GODOC comments
- build/install instructions
- screenshots

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
