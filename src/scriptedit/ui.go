package scriptedit

import (
	"strings"
	"fmt"
	"bytes"
)

const STATUS_POS = 43
const WIDTH = 132
const POINTER = WIDTH/2

type EditorState struct {
	Position   int
	Bytepos    int
	Time       float32
	Total_time float32
	Content    []AnsiCmd
}

func (ttyfd TTY) readCursorPosition() (int, int) {
	ttyfd.write(READ_CURSOR_POSITION)
	var chr byte
	var nb []byte = make([]byte, 0)
	var x, y int

	chr = ttyfd.Readchr()
	if chr != ESC_CHR {
		return 0, 0 // something failed !
	}
	chr = ttyfd.Readchr()
	if chr != '[' {
		return 0, 0 // something failed !
	}

	for chr != ';' {
		chr = ttyfd.Readchr()
		nb = append(nb, chr)
	}
	fmt.Sscanf(string(nb), "%d", &x)
	nb = make([]byte, 0)
	for chr != 'R' {
		chr = ttyfd.Readchr()
		nb = append(nb, chr)
	}
	fmt.Sscanf(string(nb), "%d", &y)
	return x, y

}


func (ttyfd TTY) writeTicker(state *EditorState) {
	left := state.Position - POINTER
	if left < 0 {
		ttyfd.write(strings.Repeat("-", -left))
		left = 0
	}
	right := state.Position + WIDTH - POINTER
	size := len(state.Content)
	if right >= size {
		defer ttyfd.write(strings.Repeat("-", right - size))
		right = size - 1
	}


	for index, ansi := range (state.Content[left:right]) {
		if ansi.Code != nil {
			ttyfd.write(ansi.Code.Symbol)
		} else {
			ttyfd.write(string(EdulcorateCharacter(ansi.Letter)))
		}
		if index == WIDTH {
			break
		}
	}

}

var TOP_NAVBAR = strings.Repeat("─", POINTER) + "┬" + strings.Repeat("─", WIDTH - POINTER - 1)
var BOTTOM_NAVBAR = strings.Repeat("─", POINTER) + "┴" + strings.Repeat("─", WIDTH - POINTER - 1)

func (ttyfd TTY) navBar(state *EditorState) {
	ttyfd.write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS, 1))
	ttyfd.write(TOP_NAVBAR)
	ttyfd.write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 1, 1))
	ttyfd.writeTicker(state)
	ttyfd.write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 2, 1))
	ttyfd.write(BOTTOM_NAVBAR)
}

func (ttyfd TTY) WriteStatus(state *EditorState) {
	x, y := ttyfd.readCursorPosition()
	ttyfd.write(RESET_COLOR)
	ttyfd.navBar(state)
	var explanation string
	currentAnsi := state.Content[state.Position]
	if currentAnsi.Code != nil {
		explanation = currentAnsi.Code.Explanation
	} else {
		explanation = fmt.Sprintf("Character %c (%x)", EdulcorateCharacter(currentAnsi.Letter), currentAnsi.Letter)
	}

	if currentAnsi.Params != "" {
		explanation += " (" + currentAnsi.Params + ")"
	}
	leftExplanation := POINTER - len(explanation)/2

	ttyfd.write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 3, leftExplanation))
	ttyfd.write("| " + explanation + " |")

	ttyfd.write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 5, 5))
	ttyfd.write(fmt.Sprintf("Offset %d / %d", state.Position, len(state.Content)))
	ttyfd.write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 6, 5))

	ttyfd.write(fmt.Sprintf("Time   %f / %f s", state.Time, state.Total_time))

	ttyfd.write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 5, 50))
	ttyfd.write(fmt.Sprintf("[←] : for   [→] : rev   [SPACE] : Play/Pause   [d] : del  %dx%d", x, y))
	ttyfd.write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 6, 50))
	ttyfd.write(fmt.Sprintf("[CTRL] + [←] : ff   [CTRL] + [→] : rw   [q] : quit"))
	ttyfd.write(fmt.Sprintf(MOVE_CURSOR, x, y))
}

func (ttyfd TTY) Notify(message string) {
	ttyfd.write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 3, 20))
	ttyfd.write(fmt.Sprintf(message))

}

func (ttyfd TTY) Redraw(state *EditorState) {
	ttyfd.write(RESET)
	ttyfd.write(CLEAR_SCREEN)
	var buffer bytes.Buffer

	for _, ansi := range (state.Content[0:state.Position]) {
		buffer.WriteString(ansi.String())
	}
	state.Bytepos = buffer.Len()
	ttyfd.write(buffer.String())
	ttyfd.WriteStatus(state)
}

func (ttyfd TTY) Init() {
	ttyfd.write(SMCUP)
	ttyfd.write(CLEAR_SCREEN)
}

func (ttyfd TTY) Restore() {
	ttyfd.write(RESET)
	ttyfd.write(RMCUP)
	ttyfd.write(CLEAR_SCREEN)
}
