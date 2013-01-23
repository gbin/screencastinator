package main

import (
	"os"
	"fmt"
	"bufio"
	"scriptedit"
	"flag"
)

var editorState scriptedit.EditorState

var sessionFilename string
var timingFilename *string

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
	if timingFilename == nil || *timingFilename == "" || sessionFilename == "" {
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
	timingsreader := bufio.NewReader(timings_file)
	editorState.ParseTimings(timingsreader)
	timings_file.Close()


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
							if editorState.PreviousTiming() {
								ttyfd.Redraw(&editorState)
							}
						case FORWARD:
							if editorState.NextTiming() {
								ttyfd.Redraw(&editorState)
							}
						}

					}
				case '3':
					if (ttyfd.Readchr() == '~') { // this is DEL
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


