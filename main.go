package main

import (
	"os"
	"fmt"
	"bufio"
	"screencastinator/scriptedit"
	"flag"
)

var editorState scriptedit.EditorState

var sessionFilename string
var timingFilename string

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
		fmt.Fprintf(os.Stderr, "basefilename\n\n")
		fmt.Fprintf(os.Stderr, "It will load the [basefilename].session and [basefilename].timing\n\n")
		fmt.Fprintf(os.Stderr, "The timing and session files can be created with the standard tool \"script\" that comes with the linux-util package.\n\nNote: You need to name your session and timing file with the .session and .timing extensions like this:\n%% script --timing=test.timing test.session\n\nYou can then edit it with:\n%% screencastinator test\n\nSee http://www.linuxinsight.com/replaying-terminal-sessions-with-scriptreplay.html for more information\n\n")
	}

	flag.Parse()


	if flag.Arg(0) == "" {
		flag.Usage()
		return
	}
	sessionFilename = flag.Arg(0) + ".session"
	timingFilename = flag.Arg(0) + ".timing"


	file, err := os.Open(sessionFilename);
	if err != nil {
		fmt.Println(err)
		return
	}

	contentreader := bufio.NewReader(file)
	contentreader.ReadBytes('\n') // Kicks out the preliminary from script (This script has been started BLAHBLAH
	editorState.Content = scriptedit.ParseANSI(contentreader)

	file.Close()
	timings_file, err := os.Open(timingFilename);
	if err != nil {
		fmt.Println(err)
		return
	}
	timingsreader := bufio.NewReader(timings_file)
	editorState.ParseTimings(timingsreader)
	timings_file.Close()

	editorState.In = -1
	editorState.Out = -1


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
			for _, entry := range editorState.Timings {
				file.WriteString(fmt.Sprintf("%f %d\n", entry.Time, entry.Length))
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

	playing := false
out:
	for {
		chr, _, err := ttyfd.Readchr()

		if playing && err != nil {
			ttyfd.PlayingPoll(&editorState)
			continue out
		}

		switch chr {
		case ESC_CHR:
			chr, _, _ = ttyfd.Readchr()
			if chr == '[' {
				chr, _, _ = ttyfd.Readchr()

				switch chr {
				case '1':
					chr, _, _ = ttyfd.Readchr()
					if (chr == ';') {
						chr, _, _ = ttyfd.Readchr()
						if (chr == '5') {
							chr, _, _ = ttyfd.Readchr()
							switch chr {   // this is a CTRL + ARROW
							case BACK:
								if editorState.PreviousTiming() {
									ttyfd.Redraw(&editorState)
								}
							case FORWARD:
								if editorState.NextTiming() {
									ttyfd.Redraw(&editorState)
								}
							}  }

					}
				case '3':
					chr, _, _ = ttyfd.Readchr()
					if (chr == '~') { // this is DEL
						if editorState.In == -1 {
							editorState.DeleteRegion(editorState.Position, editorState.Position + 1)
							ttyfd.WriteStatus(&editorState) // It should not change the screen
						} else {
							editorState.DeleteRegion(editorState.In, editorState.Out)
							if editorState.Position != editorState.In {
								editorState.Position = editorState.In
								editorState.In = -1
								editorState.Out = -1
								ttyfd.Redraw(&editorState)
							} else {
								editorState.In = -1
								editorState.Out = -1
								ttyfd.WriteStatus(&editorState) // It should not change the screen
							}
						}

					}
				case BACK:
					if editorState.Previous() {
						ttyfd.Redraw(&editorState)
					}
				case FORWARD:
					if editorState.Next() {
						ttyfd.Redraw(&editorState)
					}

				}
			} else if chr == ESC_CHR {
				editorState.Out = -1
				editorState.In = -1
				ttyfd.Redraw(&editorState)
			}

		case 'i':
			editorState.In = editorState.Position
			if editorState.Out < editorState.In {
				editorState.Out = editorState.In + 1
			}
			ttyfd.Redraw(&editorState)

		case 'o':
			editorState.Out = editorState.Position
			if editorState.Out < editorState.In {
				editorState.Out = editorState.In + 1
			}

			ttyfd.Redraw(&editorState)

		case 'n':
			if (editorState.In == -1) {
				editorState.In = editorState.Position
			}
			result := ttyfd.JumpToNextSameCursorPosition(&editorState)
			if result {
				editorState.Out = editorState.Position
				ttyfd.WriteStatus(&editorState)
			} else {
				ttyfd.Redraw(&editorState)
			}

		case 'q':
			break out
		case ' ':
			if !playing {
				ttyfd.StartPlaying(&editorState)
			}
			playing = !playing
			ttyfd.SetNonBlocking(playing)
		case 's':
			err := save(sessionFilename, timingFilename)
			if err != nil {
				fmt.Println(err)
			}
		default :
			ttyfd.Notify(fmt.Sprintf("Unknown Key '%c' (%d)", chr, chr))
		}
	}
	return nil
}


