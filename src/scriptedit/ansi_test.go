package scriptedit

import (
	"testing"
	"strings"
	"bufio"
)

func TestParsing(t *testing.T) {
	r := bufio.NewReader(strings.NewReader("This is a normal string\033[12;13fNormal Again\033[A"))
	parsedAnsi := ParseANSI(r)
	if len(parsedAnsi) != 37 {
		t.Errorf("Wrong length %d ", len(parsedAnsi))
	}
	if parsedAnsi[0].Letter != 'T' {
		t.Error("Could not parse a normal letter")
	}
	if parsedAnsi[23].Code.Symbol != "⇲" {
		t.Error("Could not a CSI sequence")
	}
	if parsedAnsi[23].Params != "12;13" {
		t.Error("Could not CSI parameters")
	}
	if parsedAnsi[24].Letter != 'N' {
		t.Error("Could not come back to normal letters")
	}

}

func TestUnitRendering(t *testing.T) {
	a := AnsiCmd{0, &CUP, "12;11"}
	if a.String() != "\033[12;11H" {
		t.Errorf("This is not the normal representation (ESC+%s)", a.String()[1:])
	}
}


func TestRendering(t *testing.T) {
	orig := "This is a normal string\033[12;13fNormal Again\033[A"
	r := bufio.NewReader(strings.NewReader(orig))
	parsedAnsi := ParseANSI(r)
	var dest string = ""
	for _, ansi := range parsedAnsi {
		dest += ansi.String()
	}
	if orig != dest {
		t.Errorf("Problem while rerendering %s!=%s", orig, dest)
	}

}

