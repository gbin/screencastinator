package scriptedit

import (
	"testing"
	"bufio"
	"strings"
)

func getPopulatedEditorState(t *testing.T) *EditorState {
	editorState := NewEditorState()
	s := "01234567890\033[12;13f2345678901234\033[A"
	r := bufio.NewReader(strings.NewReader(s))
	editorState.Content = ParseANSI(r)
	editorState.Timings = []Timing {Timing{1.1, 6}, Timing{2.3, 16}, Timing{12.1, 5}, Timing{12, 1}, Timing{1, 7}}
	var l int
	for _, t := range (editorState.Timings) {
		l += t.Length
	}
	if l != len(s) {
		t.Errorf("Wrong initial state %i != %i", l, len(s))
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
		for index, chr := range (state.Content) {
			if chr.Code != nil {
				t.Errorf("%d = %s\n", index, chr.Code.Symbol)
			} else {
				t.Errorf("%d = %s\n", index, chr)
			}
		}
	}
	position= state.Bytepos2position(22)
	if position != 15 {
		t.Errorf("Wrong index %d\n", position)
		t.Errorf("Targeted CHR %s\n", "01234567890\033[12;13f2345678901234\033[A"[22:23])
		for index, chr := range (state.Content) {
			if chr.Code != nil {
				t.Errorf("%d = %s\n", index, chr.Code.Symbol)
			} else {
				t.Errorf("%d = %s\n", index, chr)
			}
		}
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
	state.Position = 7 // it should snap to the next timing
	state.NextTiming()
	if state.Bytepos != 22 {
		t.Errorf("Wrong Bytepos %d", state.Bytepos)
	}
	if state.Position != 15 {
		t.Errorf("Wrong position %d", state.Position)
	}



}

//func (state *EditorState) PreviousTiming() bool {
//	timeindex, offset, _ := state.deduceTiming(state.Bytepos)
//	if state.Bytepos > offset {
//		state.Position = state.Bytepos2position(offset)
//		return true
//	}
//	if timeindex > 0 {
//		state.Position = state.Bytepos2position(state.Bytepos - state.Timings[timeindex - 1].Length)
//		return true
//	}
//	return false
//}
//
//func (state *EditorState) DeleteRegion(from, to int) bool {
//	var bytesToRemove int
//	for i := from; i < to; i++ {
//		bytesToRemove += len([]byte(state.Content[i].String()))
//	}
//
//	timeindex, _, _ := state.deduceTiming(state.Bytepos) //  find back from with time bucket it is from
//
//	var bytesRemoved int
//	for bytesRemoved < bytesToRemove {
//		currentElementLength := state.Timings[timeindex].Length
//		if currentElementLength > bytesToRemove {
//			state.Timings[timeindex].Length-=bytesToRemove
//			break
//		}
//		bytesRemoved += currentElementLength
//		copy(state.Timings[timeindex:], state.Timings[timeindex + 1:])
//		state.Timings = state.Timings[:len(state.Timings) - 1]
//	}
//
//	copy(state.Content[from:], state.Content[to:])
//	state.Content = state.Content[:len(state.Content) - (to - from)]
//	_, _, state.Time = state.deduceTiming(state.Bytepos)
//
//	return false
//
//}
//
//
//func (state *EditorState) deduceTiming(position int) (int, int, float32) {
//	var time float32
//	var index int
//	for pos , t := range state.Timings {
//		if index >= position {
//			return pos, index, time
//		}
//		time += t.Time
//		index += t.Length
//	}
//	return len(state.Timings) - 1 , index, time
//}
//
//func (state *EditorState) ParseTimings(reader *bufio.Reader) {
//	state.Timings = make([]Timing, 0)
//
//	// Put an artificial 0 at the beginning
//	state.Timings = append(state.Timings, Timing{0, 0})
//
//	for true {
//		var line, err = reader.ReadBytes('\n')
//		var entry Timing = Timing{0, 0}
//		if err != nil {
//			break
//		}
//		fmt.Sscanf(string(line), "%f %d", &entry.Time, &entry.Length)
//		state.Timings = append(state.Timings, entry)
//	}
//	_, _, state.Total_time = state.deduceTiming(len(state.Content) - 1)
//
//}
