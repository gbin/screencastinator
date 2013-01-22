package main

import (
	"os"
	"fmt"
	"bufio"
	"syscall"
	"strings"
	"scriptedit"
	"bytes"
	"flag"
)

type Timing struct {
	time   float32
	length int
}

//var content string

var sessionFilename string
var timingFilename *string
var parsedcontent []scriptedit.AnsiCmd

var timings []Timing = make([]Timing, 0)

var position int
var bytepos int
var size int
var total_time float32

const STATUS_POS = 43
const WIDTH = 132
const POINTER = WIDTH/2

var TOP_NAVBAR = strings.Repeat("─", POINTER) + "┬" + strings.Repeat("─", WIDTH - POINTER - 1)
var BOTTOM_NAVBAR = strings.Repeat("─", POINTER) + "┴" + strings.Repeat("─", WIDTH - POINTER - 1)

const ESC = scriptedit.ESC
const ESC_CHR = scriptedit.ESC_CHR

const RESET string = ESC + "c"
const SMCUP = ESC + "7" + ESC + "[?47h"
const RMCUP = ESC + "[2J" + ESC + "[?47l" + ESC + "8"

const CLEAR_SCREEN string = ESC + "[2J"

const CHANGE_SIZE = ESC + "[8;%d;%dt"
const MOVE_CURSOR = ESC + "[%d;%dH"

const RESET_COLOR = ESC + "[0m"
const READ_CURSOR_POSITION = ESC + "[6n"

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
	ttyfd int = 0 // STDIN_FILENO
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

	file, err := os.Open("demo.session");
	if err != nil {
		fmt.Println(err)
		return
	}

	contentreader := bufio.NewReader(file)
	contentreader.ReadBytes('\n') // Kicks out the preliminary from script (This script has been started BLAHBLAH
	parsedcontent = scriptedit.ParseANSI(contentreader)
	size = len(parsedcontent)

	file.Close()
	timings_file, err := os.Open("demo.timing");
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
	_, _, total_time = deduceTiming(size - 1)

	defer func() {
		if err != nil { fmt.Println(err) }
	}();

	err = scriptedit.GetTermios(&orig_termios, ttyfd)
	if err != nil {
		fmt.Println("GetTermios fluked", err)
		return
	}

	defer func() {
		err = scriptedit.SetTermios(&orig_termios, ttyfd)
	}();
	new_termios = orig_termios
	err = scriptedit.Tty_raw(&new_termios, ttyfd)
	if err != nil {
		fmt.Println("Tty_raw fluked", err)
		return
	}
	err = screenio()

	if err != nil {
		fmt.Println(err)
		return
	}
}

func write(chrs string) {
	_, err := syscall.Write(ttyfd, []byte(chrs))
	if err != nil {
		fmt.Println(err)
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

func readCursorPosition() (int, int) {
	write(READ_CURSOR_POSITION)
	var chr byte
	var nb []byte = make([]byte, 0)
	var x, y int

	chr = readchr()
	if chr != ESC_CHR {
		return 0, 0 // something failed !
	}
	chr = readchr()
	if chr != '[' {
		return 0, 0 // something failed !
	}

	for chr != ';' {
		chr = readchr()
		nb = append(nb, chr)
	}
	fmt.Sscanf(string(nb), "%d", &x)
	nb = make([]byte, 0)
	for chr != 'R' {
		chr = readchr()
		nb = append(nb, chr)
	}
	fmt.Sscanf(string(nb), "%d", &y)
	return x, y

}



func writeTicker() {
	left := position - POINTER
	if left < 0 {
		write(strings.Repeat("-", -left))
		left = 0
	}
	right := position + WIDTH - POINTER
	if right >= size {
		defer write(strings.Repeat("-", right - size))
		right = size - 1
	}


	for index, ansi := range (parsedcontent[left:right]) {
		if ansi.Code != nil {
			write(ansi.Code.Symbol)
		} else {
			write(string(scriptedit.EdulcorateCharacter(ansi.Letter)))
		}
		if index == WIDTH {
			break
		}
	}

}
func navBar() {
	write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS, 1))
	write(TOP_NAVBAR)
	write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 1, 1))
	writeTicker()
	write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 2, 1))
	write(BOTTOM_NAVBAR)
}

func writeStatus() {
	x, y := readCursorPosition()
	write(RESET_COLOR)
	navBar()
	var explanation string
	currentAnsi := parsedcontent[position]
	if currentAnsi.Code != nil {
		explanation = currentAnsi.Code.Explanation
	} else {
		explanation = fmt.Sprintf("Character %c (%x)", scriptedit.EdulcorateCharacter(currentAnsi.Letter), currentAnsi.Letter)
	}

	if currentAnsi.Params != "" {
		explanation += " (" + currentAnsi.Params + ")"
	}
	leftExplanation := POINTER - len(explanation)/2

	write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 3, leftExplanation))
	write("| " + explanation + " |")

	write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 5, 5))
	write(fmt.Sprintf("Offset %d / %d", position, size))
	write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 6, 5))
	_, _, time := deduceTiming(bytepos)
	write(fmt.Sprintf("Time   %f / %f s", time, total_time))

	write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 5, 50))
	write(fmt.Sprintf("[←] : for   [→] : rev   [SPACE] : Play/Pause   [d] : del  %dx%d", x, y))
	write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 6, 50))
	write(fmt.Sprintf("[CTRL] + [←] : ff   [CTRL] + [→] : rw   [q] : quit"))
	write(fmt.Sprintf(MOVE_CURSOR, x, y))
}

func redraw() {
	write(RESET)
	write(CLEAR_SCREEN)
	var buffer bytes.Buffer

	for _, ansi := range (parsedcontent[0:position]) {
		buffer.WriteString(ansi.String())
	}
	bytepos = buffer.Len()
	write(buffer.String())
	writeStatus()
}

func bytepos2position(bytepos int) int {
	var offset int
	for index, ansi := range parsedcontent {
		offset+= len(ansi.String())
		if offset >= bytepos {
			return index
		}
	}
	return -1
}

func readchr() byte {
	var c_in [1]byte
	syscall.Read(ttyfd, c_in[0:])
	return c_in[0]
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
		for _, ansi := range parsedcontent {
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

			write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 3, 20))
			write(fmt.Sprintf("File Saved"))
		}

	}
	return nil

}

func screenio() error {
	write(SMCUP)
	write(CLEAR_SCREEN)
	writeStatus()
out:
	for {
		chr := readchr()
		switch chr {
		case ESC_CHR:
			chr = readchr()
			if chr == '[' {
				chr = readchr()

				switch chr {
				case '1':
					if (readchr() == ';' && readchr() == '5') {
						chr = readchr()
						switch chr {   // this is a CTRL + ARROW
						case BACK:
							timeindex, offset, _ := deduceTiming(bytepos)
							if bytepos > offset {
								position = bytepos2position(offset)
								redraw()
							} else {
								if timeindex > 0 {
									position = bytepos2position(bytepos - timings[timeindex - 1].length)
									redraw()
								}
							}
							break
						case FORWARD:
							timeindex, offset, _ := deduceTiming(bytepos)
							if timeindex < len(timings) - 1 {
								position = bytepos2position(offset + timings[timeindex + 1].length)
								redraw()
							}
							break
						}

					}
				case '3':
					if (readchr() == '~') { // this is DEL
						bytesToRemove := len([]byte(parsedcontent[position].String()))
						timeindex, _, _ := deduceTiming(bytepos)     // FIXME we should look if it is at the end or at the beginning of the block
						if timings[timeindex].length >= bytesToRemove {
							timings[timeindex].length-= bytesToRemove
						} else {
							bytesToRemove-= timings[timeindex].length
							timings[timeindex].length = 0
							timings[timeindex + 1].length -= bytesToRemove  // FIXME, we should recurse here

						}
						copy(parsedcontent[position:], parsedcontent[position + 1:])
						parsedcontent = parsedcontent[:len(parsedcontent) - 1]
						writeStatus() // It should not change the screen
					}
				case BACK:
					if position > 0 {
						position-=1
					}
					redraw()
				case FORWARD:
					if position < size {
						position+=1
					}
					redraw()
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
			write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 3, 50))
			write(fmt.Sprintf("Unknown Key '%c' (%d)", chr, chr))

		}
	}

	write(RESET)
	write(RMCUP)
	write(CLEAR_SCREEN)
	return nil
}


