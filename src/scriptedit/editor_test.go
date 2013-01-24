package scriptedit

import (
	"testing"
	"bufio"
	"strings"
)

var DOC string = "01234567890\033[12;13f2345678901234\033[A"

func (state *EditorState) fullStateDump(t *testing.T) {
	t.Errorf("------------\n")
	for index, chr := range (state.Content) {
		if chr.Code != nil {
			t.Errorf("%d = %s\n", index, chr.Code.Symbol)
		} else {
			t.Errorf("%d = %s\n", index, chr)
		}
	}
	t.Errorf("\n\n\n")
	for i, timing := range (state.Timings) {
		t.Errorf("%d = %f %d\n", i, timing.Time, timing.Length)
	}
}

func getPopulatedEditorState(t *testing.T) *EditorState {
	editorState := NewEditorState()
	r := bufio.NewReader(strings.NewReader(DOC))
	editorState.Content = ParseANSI(r)
	editorState.Timings = []Timing {Timing{1.1, 6}, Timing{2.3, 16}, Timing{12.1, 5}, Timing{12, 1}, Timing{1, 7}}
	var l int
	for _, t := range (editorState.Timings) {
		l += t.Length
	}
	if l != len(DOC) {
		t.Errorf("Wrong initial state %i != %i", l, len(DOC))
	}

	return editorState
}

func TestBytepos2position(t *testing.T) {
	state := getPopulatedEditorState(t)
	position := state.Bytepos2position(10)
	if position != 10 {
		t.Errorf("Wrong index %d", position)
	}
	position = state.Bytepos2position(11)
	if position != 11 {
		t.Errorf("Wrong index %d", position)
	}
	p1, p2, p3, p4 := state.Bytepos2position(12), state.Bytepos2position(13), state.Bytepos2position(14), state.Bytepos2position(15)
	if p1 != 11 || p2 != 11 || p3 != 11 || p4 != 11 {
		t.Errorf("Wrong index %d,%d,%d,%d\n", p1, p2, p3, p4)
		state.fullStateDump(t)
	}
	position = state.Bytepos2position(22)
	if position != 15 {
		t.Errorf("Wrong index %d\n", position)
		t.Errorf("Targeted CHR %s\n", DOC[22:23])
		state.fullStateDump(t)
	}

}

func TestDeduceTiming(t *testing.T) {
	state := getPopulatedEditorState(t)

	timeindex, offset, time := state.deduceTiming(0)
	if timeindex != 0 {
		t.Errorf("Wrong timing index %d", timeindex)
	}
	if offset != 0 {
		t.Errorf("Wrong offset %d", offset)
	}
	if time != 1.1 {
		t.Errorf("Wrong timing %f", time)
	}


	timeindex, offset, time = state.deduceTiming(5)
	if timeindex != 0 {
		t.Errorf("Wrong timing index %d", timeindex)
	}
	if offset != 0 {
		t.Errorf("Wrong offset %d", offset)
	}
	if time != 1.1 {
		t.Errorf("Wrong timing %f", time)
	}

	timeindex, offset, time = state.deduceTiming(6)
	if timeindex != 1 {
		t.Errorf("Wrong timing index %d", timeindex)
	}
	if offset != 6 {
		t.Errorf("Wrong offset %d", offset)
	}
	if time != 3.4 {
		t.Errorf("Wrong timing %f", time)
	}



	timeindex, offset, time = state.deduceTiming(41)
	if timeindex != 4 {
		t.Errorf("Wrong timing index %d", timeindex)
	}
	if offset != 35 {
		t.Errorf("Wrong offset %d", offset)
	}
	if time != 28.5 {
		t.Errorf("Wrong timing %f", time)
	}
}

func TestNextPrevTiming(t *testing.T) {
	state := getPopulatedEditorState(t)
	state.NextTiming()
	if state.Bytepos != 6 {
		t.Errorf("Wrong Bytepos %d", state.Bytepos)
	}
	if state.Position != 6 {
		t.Errorf("Wrong position %d", state.Position)
	}
	state.NextTiming()
	if state.Bytepos != 22 {
		t.Errorf("Wrong Bytepos %d", state.Bytepos)
	}
	if state.Position != 15 {
		t.Errorf("Wrong position %d", state.Position)
	}
	state.PreviousTiming()
	if state.Bytepos != 6 {
		t.Errorf("Wrong Bytepos %d", state.Bytepos)
	}
	if state.Position != 6 {
		t.Errorf("Wrong position %d", state.Position)
	}

	state.Bytepos = 7
	state.Position = 7 // it should snap to the next timing
	state.PreviousTiming()
	if state.Bytepos != 6 {
		t.Errorf("Wrong Bytepos %d", state.Bytepos)
	}
	if state.Position != 6 {
		t.Errorf("Wrong position %d", state.Position)
	}

	state.Bytepos = 7
	state.Position = 7 // it should snap to the next timing         "01234567890\033[12;13f2345678901234\033[A"
	state.NextTiming()
	if state.Bytepos != 22 {
		t.Errorf("Wrong Bytepos %d", state.Bytepos)
	}
	if state.Position != 15 {
		t.Errorf("Wrong position %d", state.Position)
	}
}

func TestDeleteRegionBegin(t *testing.T) {
	state := getPopulatedEditorState(t)
	l := len(state.Content)
	timings := len(state.Timings)
	state.DeleteRegion(0, 1)

	if state.Timings[0].Length != 5 {
		t.Errorf("Wrong change in the timings %d", state.Timings[0].Length)
	}

	if len(state.Content) != l - 1 {
		t.Errorf("Size error %d", len(state.Content))
	}

	state.DeleteRegion(0, 7)
	if len(state.Content) != l - 8 {
		t.Errorf("Size error %d", len(state.Content))
	}

	if timings - 1 != len(state.Timings) {
		t.Errorf("Wrong change in the timings deletion %d", len(state.Timings))
	}

	if state.Timings[0].Length != 14 {
		t.Errorf("Wrong change in the timings %d", state.Timings[0].Length)
	}

}

func TestDeleteRegionMiddle(t *testing.T) {
	state := getPopulatedEditorState(t)
	state.DeleteRegion(2, 12)

	if state.Content[0].Letter == '2' {
		t.Errorf("Wrong region removed %s", state.Content[0])
		//state.fullStateDump(t)
	}

	if state.Timings[0].Length != 2 {
		t.Errorf("Wrong change in the timings %d", state.Timings[0].Length)
		//state.fullStateDump(t)
	}


}
