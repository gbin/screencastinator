package scriptedit

import (
	"strings"
	"fmt"
	"bytes"
	"time"
)

const STATUS_POS = 43
const WIDTH = 132
const POINTER = WIDTH/2

func (ttyfd TTY) readCursorPosition() (int, int) {
	ttyfd.write(READ_CURSOR_POSITION)
	var chr byte
	var nb []byte = make([]byte, 0)
	var x, y int

	chr, _, _ = ttyfd.Readchr()
	if chr != ESC_CHR {
		return 0, 0 // something failed !
	}
	chr, _, _ = ttyfd.Readchr()
	if chr != '[' {
		return 0, 0 // something failed !
	}

	for chr != ';' {
		chr, _, _ = ttyfd.Readchr()
		nb = append(nb, chr)
	}
	fmt.Sscanf(string(nb), "%d", &x)
	nb = make([]byte, 0)
	for chr != 'R' {
		chr, _, _ = ttyfd.Readchr()
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
		inSelect := index + left >= state.In && index + left < state.Out
		if inSelect {
			ttyfd.write(ESC + "[43m")
		}

		if ansi.Code != nil {
			if *ansi.Code == SGR && !inSelect {
				ttyfd.write(ansi.String())
			}
			ttyfd.write(ansi.Code.Symbol)
		} else {
			ttyfd.write(string(EdulcorateCharacter(ansi.Letter)))
		}
		if index == WIDTH {
			break
		}
		ttyfd.write(ESC + "[40m")
	}
	ttyfd.write(RESET_COLOR) // In case the last one was zorglubbed

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

func (ttyfd TTY) JumpToNextSameCursorPosition(state *EditorState) bool {
	initx, inity := ttyfd.readCursorPosition()
	position := state.Position
	for position < len(state.Content) {
		ttyfd.write(state.Content[position].String())
		position++
		x, y := ttyfd.readCursorPosition()
		if x == initx && y == inity {
			state.Position = position
			return true
		}
	}
	return false
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

	ttyfd.write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 3, 2))
	ttyfd.write(fmt.Sprintf("Offset %d / %d", state.Position, len(state.Content)))

	ttyfd.write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 3, 23))
	ttyfd.write(fmt.Sprintf("Time   %.2f / %.2f s", state.Time, state.Total_time))

	ttyfd.write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 3, 123))
	ttyfd.write(fmt.Sprintf("Cur %dx%d", x, y))

	ttyfd.write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 5, 0))
	ttyfd.write(fmt.Sprintf("         [←] : reverse       [→] : forward         [SPACE] : Play/Pause       [i] : IN mark         [o] : OUT mark        [d] : del"))
	ttyfd.write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 6, 0))
	ttyfd.write(fmt.Sprintf("[CTRL] + [←] : rw   [CTRL] + [→] : ff                                         [n] : smart extend    [s] : SAVE            [q] : quit"))
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

var playTime int64
var refTime int64
var playTimeIndex int

func (ttyfd TTY) PlayingPoll(state *EditorState) {
	var nbSecDone int64 = time.Now().UnixNano() - refTime
	bytesToPlay := 0
	for playTime < nbSecDone {
		bytesToPlay += state.Timings[playTimeIndex].Length
		playTime += int64(float64(state.Timings[playTimeIndex].Time)*1000000000)
		playTimeIndex++
	}
	chunkToDisplay := ""
	for len(chunkToDisplay) < bytesToPlay {
		chunkToDisplay += state.Content[state.Position].String()
		state.Position++
	}

	ttyfd.write(chunkToDisplay)
	if state.Timings[playTimeIndex].Time > .250 {
		time.Sleep(250)
	} else {
		time.Sleep(time.Duration(state.Timings[playTimeIndex].Time)*time.Millisecond)
	}

}

func (ttyfd TTY) StartPlaying(state *EditorState) {
	playTime = 0
	refTime = time.Now().UnixNano()
	playTimeIndex, _, _ = state.deduceTiming(state.Bytepos)
}
