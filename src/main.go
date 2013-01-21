package main

import (
	"os"
	"fmt"
	"bufio"
	"syscall"
	"strings"
	"scriptedit"
)

type Timing struct {
	time   float32
	length int
}

//var content string
var parsedcontent []scriptedit.AnsiCmd

var timings []Timing = make([]Timing, 0)

var position int
var size int
var total_time float32

const STATUS_POS = 43

var HORIZONTAL_LINE = strings.Repeat("─", 132)
const ESC = scriptedit.ESC
const ESC_CHR = scriptedit.ESC_CHR

const RESET string = ESC + "c"
const CLEAR_SCREEN string = ESC + "[2J"
const CHANGE_SIZE = ESC + "[8;%d;%dt"
const MOVE_CURSOR = ESC + "[%d;%dH"
const RESET_COLOR = ESC + "[0m"
const READ_CURSOR_POSITION = ESC + "[6n"

// keys
const UP byte = 'A'
const DOWN byte = 'B'
const FORWARD byte = 'C'
const BACK byte = 'D'

const CTRL_PREFIX = "1;5"

var (
	orig_termios scriptedit.Termios;
	ttyfd int = 0 // STDIN_FILENO
)

func main() {
	//rawcontent := make([]byte, 1000000)
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

	err = scriptedit.Tty_raw(&orig_termios, ttyfd)
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

func writeStatus() {
	x, y := readCursorPosition()
	write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS, 1))
	write(RESET_COLOR)
	write(HORIZONTAL_LINE)
	write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 1, 1))
	for index, ansi := range(parsedcontent[position:]) {
		if ansi.Code != nil {
			write(ansi.Code.Symbol)
		} else {
			switch ansi.Letter {
			case '\000':
				write("?")
			case '\001':
				write("?")
			case '\002':
				write("?")
			case '\003':
				write("?")
			case '\004':
				write("?")
			case '\005':
				write("?")
			case '\006':
				write("?")
			case '\007':
				write("?")
			case '\010':
				write("?")
			case '\011':
				write("?")
			case '\012':
				write("?")
			case '\013':
				write("?")
			case '\014':
				write("?")
			case '\015':
				write("?")
			case '\016':
				write("?")
			case '\017':
				write("?")
			case '\020':
				write("?")
			case '\021':
				write("?")
			case '\022':
				write("?")
			case '\023':
				write("?")
			case '\024':
				write("?")
			case '\025':
				write("?")
			case '\026':
				write("?")
			case '\027':
				write("?")
			case '\030':
				write("?")
			case '\031':
				write("?")
			case '\032':
				write("?")
			case '\033':
				write("?")
			case '\034':
				write("?")
			case '\035':
				write("?")
			case '\036':
				write("?")
			case '\037':
				write("?")
			default:
				write(string(ansi.Letter))
			}


		}
		if index == 132 {
			break
		}
	}
	write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 2, 1))
	write(HORIZONTAL_LINE)
	write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 3, 5))
	write(fmt.Sprintf("Offset %d / %d", position, size))
	write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 4, 5))
	_, _, time := deduceTiming(position)
	write(fmt.Sprintf("Time   %f / %f s", time, total_time))

	write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 3, 50))
	write(fmt.Sprintf("[←] : for   [→] : rev   [SPACE] : Play/Pause   [d] : del  %dx%d",x,y))
	write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 4, 50))
	write(fmt.Sprintf("[CTRL] + [←] : ff   [CTRL] + [→] : rw   [q] : quit"))
	write(fmt.Sprintf(MOVE_CURSOR, x, y))
}

func redraw() {
	write(RESET)
	write(CLEAR_SCREEN)
	for _, ansi := range(parsedcontent[0:position]) {
		write(ansi.String())
	}

	writeStatus()
}

func readchr() byte {
	var c_in [1]byte
	syscall.Read(ttyfd, c_in[0:])
	return c_in[0]
}

func screenio() error {

	//write(fmt.Sprintf(CHANGE_SIZE, 46, 132))
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
				if (chr == '1' && readchr() == ';' && readchr() == '5') {
					chr = readchr()
					switch chr {   // this is a CTRL + ARROW
					case BACK:
						pos, index, _ := deduceTiming(position)
						if position > index {
							position = index
							redraw()
						} else {
							if pos > 0 {
								position -= timings[pos - 1].length
								redraw()
							}
						}
						break
					case FORWARD:
						pos, index, _ := deduceTiming(position)
						if pos < len(timings) - 1 {
							position = index + timings[pos + 1].length
							redraw()
						}
						break
					}

				} else {
					switch chr {
					case BACK:
						if position > 0 {
							position-=1
						}
						redraw()
						break
					case FORWARD:
						if position < size {
							position+=1
						}

						redraw()
						break
					}
				}
			}
		case 'q':
			break out
		case ' ':
			break
		default :
			write(fmt.Sprintf(MOVE_CURSOR, STATUS_POS + 3, 50))
			write(fmt.Sprintf("Unknown Key '%c' (%d)", chr, chr))

		}
	}

	write(RESET)
	write(CLEAR_SCREEN)
	return nil
}


