package scriptedit

import (
	"bufio"
	"fmt"
)

type Timing struct {
	Time   float32
	Length int
}

type EditorState struct {
	Content        []AnsiCmd // The parsed graphical content
	Timings        []Timing  // The associated timings to play them
	Position       int       // The current index in the Content
	Bytepos        int       // The raw offset corresponding to the Position
	Time           float32   // The time correspoding to the Position
	Total_time     float32   // The total time of the replay
	In		     int       // The IN marker
	Out            int       // The OUT marker
}

func NewEditorState() *EditorState {
	state := new(EditorState)
	state.In = -1
	state.Out = -1
	return state
}

func (state *EditorState) Bytepos2position(bytepos int) int {
	var offset int
	for index, ansi := range state.Content {
		current_width := len(ansi.String())
		if offset + current_width > bytepos {
			return index
		}
		offset+= current_width
	}
	return -1
}

func (state *EditorState) NextTiming() bool {
	timeindex, offset, _ := state.deduceTiming(state.Bytepos)
	if timeindex < len(state.Timings) {
		state.Bytepos  = offset + state.Timings[timeindex].Length
		state.Position = state.Bytepos2position(state.Bytepos) // FIXME could be optimized
		return true // position changed
	}
	return false
}

func (state *EditorState) PreviousTiming() bool {
	timeindex, offset, _ := state.deduceTiming(state.Bytepos)
	if state.Bytepos > offset {
		state.Bytepos = offset
		state.Position = state.Bytepos2position(offset)
		return true
	}
	if timeindex > 0 {
		state.Bytepos -= state.Timings[timeindex - 1].Length
		state.Position = state.Bytepos2position(state.Bytepos)
		return true
	}
	return false
}

func (state *EditorState) DeleteRegion(from, to int) bool {
	var bytesToRemove int
	for i := from; i < to; i++ {
		bytesToRemove += len([]byte(state.Content[i].String()))
	}

	timeindex, _, _ := state.deduceTiming(state.Bytepos) //  find back from with time bucket it is from

	var bytesRemoved int
	for bytesRemoved < bytesToRemove {
		currentElementLength := state.Timings[timeindex].Length
		if currentElementLength > bytesToRemove {
			state.Timings[timeindex].Length-=bytesToRemove
			break
		}
		bytesRemoved += currentElementLength
		copy(state.Timings[timeindex:], state.Timings[timeindex + 1:])
		state.Timings = state.Timings[:len(state.Timings) - 1]
	}

	copy(state.Content[from:], state.Content[to:])
	state.Content = state.Content[:len(state.Content) - (to - from)]
	_, _, state.Time = state.deduceTiming(state.Bytepos)

	return false

}


// it gets the correct timing for a given absolute byte offset in the stream
func (state *EditorState) deduceTiming(offset int) (int, int, float32) {
	var time float32
	var index int
	for pos , t := range state.Timings {
		time += t.Time
		if index + t.Length > offset {
			return pos, index, time
		}
		index += t.Length
	}
	return len(state.Timings) - 1 , index, time
}

func (state *EditorState) ParseTimings(reader *bufio.Reader) {
	state.Timings = make([]Timing, 0)

	// Put an artificial 0 at the beginning
	state.Timings = append(state.Timings, Timing{0, 0})

	for true {
		var line, err = reader.ReadBytes('\n')
		var entry Timing = Timing{0, 0}
		if err != nil {
			break
		}
		fmt.Sscanf(string(line), "%f %d", &entry.Time, &entry.Length)
		state.Timings = append(state.Timings, entry)
	}
	_, _, state.Total_time = state.deduceTiming(len(state.Content) - 1)

}
