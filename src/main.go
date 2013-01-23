package main

import (
	"os"
	"fmt"
	"bufio"
	"scriptedit"
	"flag"
)

type Timing struct {
	time   float32
	length int
}

var editorState scriptedit.EditorState

var sessionFilename string
var timingFilename *string

var timings []Timing = make([]Timing, 0)

const ESC = scriptedit.ESC
const ESC_CHR = scriptedit.ESC_CHR


// const RESTORE = ESC + "[20h" + ESC + "[8m"


// keys
const UP byte = 'A'
const DOWN byte = 'B'
const FORWARD byte = 'C'
const BACK byte = 'D'

const CTRL_PREFIX = "1;5"

var (
	orig_termios scriptedit.Termios
	new_termios scriptedit.Termios
	ttyfd scriptedit.TTY = 0 // STDIN_FILENO
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s ", os.Args[0])
		fmt.Fprintf(os.Stderr, "--t timingFile sessionFile\n\n")
		fmt.Fprintf(os.Stderr, "The timing and session files can be created with the standard tool \"script\" that comes with the linux-util package.\nSee for example http://www.linuxinsight.com/replaying-terminal-sessions-with-scriptreplay.html\n\n")
	}

	timingFilename = flag.String("t", "", "The timings filename of your script capture")
	flag.Parse()

	sessionFilename = flag.Arg(0)
	if timingFilename == nil ||  *timingFilename == "" || sessionFilename == "" {
		flag.Usage()
		return
	}

	file, err := os.Open(sessionFilename);
	if err != nil {
		fmt.Println(err)
		return
	}

	contentreader := bufio.NewReader(file)
	contentreader.ReadBytes('\n') // Kicks out the preliminary from script (This script has been started BLAHBLAH
	editorState.Content = scriptedit.ParseANSI(contentreader)

	file.Close()
	timings_file, err := os.Open(*timingFilename);
	if err != nil {
		fmt.Println(err)
		return
	}
	defer timings_file.Close()
	timingsreader := bufio.NewReader(timings_file)

	// Put an artificial 0 at the beginning
	timings = append(timings, Timing{0, 0})

	for true {
		var line, err = timingsreader.ReadBytes('\n')
		var entry Timing = Timing{0, 0}
		if err != nil {
			break
		}
		fmt.Sscanf(string(line), "%f %d", &entry.time, &entry.length)
		timings = append(timings, entry)
	}
	_, _, editorState.Total_time = deduceTiming(len(editorState.Content) - 1)

	defer func() {
		if err != nil { fmt.Println(err) }
	}();

	err = ttyfd.GetTermios(&orig_termios)
	if err != nil {
		fmt.Println("GetTermios fluked", err)
		return
	}

	defer func() {
		err = ttyfd.SetTermios(&orig_termios)
	}();
	new_termios = orig_termios
	err = ttyfd.Tty_raw(&new_termios)
	if err != nil {
		fmt.Println("Tty_raw fluked", err)
		return
	}
	err = mainLoop()

	if err != nil {
		fmt.Println(err)
		return
	}
}


func deduceTiming(position int) (int, int, float32) {
	var time float32
	var index int
	for pos , t := range timings {
		if index >= position {
			return pos, index, time
		}
		time += t.time
		index += t.length
	}
	return len(timings) - 1 , index, time
}


func bytepos2position(bytepos int) int {
	var offset int
	for index, ansi := range editorState.Content {
		offset+= len(ansi.String())
		if offset >= editorState.Bytepos {
			return index
		}
	}
	return -1
}


func save(sessionFilename string, timingFilename string) error {
	var err error
	var file *os.File

	err = os.Rename(sessionFilename, sessionFilename + ".bak")
	if err != nil {
		return err
	}

	err = os.Rename(timingFilename, timingFilename + ".bak")
	if err != nil {
		return err
	}

	file, err = os.Create(sessionFilename);
	if err != nil {
		fmt.Println(err)
		return err
	} else {
		file.WriteString("This file has been edited by scriptcastinator\n")
		for _, ansi := range editorState.Content {
			file.WriteString(ansi.String())
		}
		file.Close()
		file, err = os.Create(timingFilename);
		if err != nil {
			return err
		} else {
			for _, entry := range timings {
				file.WriteString(fmt.Sprintf("%f %d\n", entry.time, entry.length))
			}
			file.Close()
			ttyfd.Notify("File Saved")
		}

	}
	return nil

}

func mainLoop() error {
	ttyfd.Init()
	ttyfd.WriteStatus(&editorState)
out:
	for {
		chr := ttyfd.Readchr()
		switch chr {
		case ESC_CHR:
			chr = ttyfd.Readchr()
			if chr == '[' {
				chr = ttyfd.Readchr()

				switch chr {
				case '1':
					if (ttyfd.Readchr() == ';' && ttyfd.Readchr() == '5') {
						chr = ttyfd.Readchr()
						switch chr {   // this is a CTRL + ARROW
						case BACK:
							timeindex, offset, _ := deduceTiming(editorState.Bytepos)
							if editorState.Bytepos > offset {
								editorState.Position = bytepos2position(offset)
								ttyfd.Redraw(&editorState)
							} else {
								if timeindex > 0 {
									editorState.Position = bytepos2position(editorState.Bytepos - timings[timeindex - 1].length)
									ttyfd.Redraw(&editorState)
								}
							}
							break
						case FORWARD:
							timeindex, offset, _ := deduceTiming(editorState.Bytepos)
							if timeindex < len(timings) - 1 {
								editorState.Position = bytepos2position(offset + timings[timeindex + 1].length)
								ttyfd.Redraw(&editorState)
							}
							break
						}

					}
				case '3':
					if (ttyfd.Readchr() == '~') { // this is DEL
						bytesToRemove := len([]byte(editorState.Content[editorState.Position].String()))
						timeindex, _, _ := deduceTiming(editorState.Bytepos)     // FIXME we should look if it is at the end or at the beginning of the block
						if timings[timeindex].length >= bytesToRemove {
							timings[timeindex].length-= bytesToRemove
						} else {
							bytesToRemove-= timings[timeindex].length
							timings[timeindex].length = 0
							timings[timeindex + 1].length -= bytesToRemove  // FIXME, we should recurse here

						}
						copy(editorState.Content[editorState.Position:], editorState.Content[editorState.Position + 1:])
						editorState.Content = editorState.Content[:len(editorState.Content) - 1]
						_, _, editorState.Time = deduceTiming(editorState.Bytepos)
						ttyfd.WriteStatus(&editorState) // It should not change the screen
					}
				case BACK:
					if editorState.Position > 0 {
						editorState.Position-=1
					}
					ttyfd.Redraw(&editorState)
				case FORWARD:
					if editorState.Position < len(editorState.Content) {
						editorState.Position+=1
					}
					ttyfd.Redraw(&editorState)
				}
			}
		case 'q':
			break out
			//case ' ':
		case 's':
			err := save(sessionFilename, *timingFilename)
			if err != nil {
				fmt.Println(err)
			}
		default :
			ttyfd.Notify(fmt.Sprintf("Unknown Key '%c' (%d)", chr, chr))
		}
	}
	return nil
}


