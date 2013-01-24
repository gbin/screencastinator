package scriptedit

import (
	"bufio"
	"fmt"
)

const (
	ESC     = "\033"
	ESC_CHR = '\033'

	CSI_CHR = '['
	OSC_CHR = ']'
	BEL     = '\007'
)

type AnsiCode struct {
	Prefix      rune
	Code        rune
	Explanation string
	Symbol      string
}

type AnsiCmd struct {
	Letter rune
	Code   *AnsiCode
	Params string
}

func (a AnsiCmd) String() string {
	if a.Code != nil {
		return fmt.Sprintf("%c%c%s%c", ESC_CHR, a.Code.Prefix, a.Params, a.Code.Code)
	}

	return string(a.Letter)
}

var (
	ICH     = AnsiCode{CSI_CHR, '@', "insert blank characters", "-"}
	CUU     = AnsiCode{CSI_CHR, 'A', "move cursor up", "↑"}
	CUD     = AnsiCode{CSI_CHR, 'B', "move cursor down", "↓"}
	CUF     = AnsiCode{CSI_CHR, 'C', "move cursor right", "→"}
	CUB     = AnsiCode{CSI_CHR, 'D', "move cursor left", "←"}
	CNL     = AnsiCode{CSI_CHR, 'E', "move cursor down and to column 1", "↲"}
	CPL     = AnsiCode{CSI_CHR, 'F', "move cursor up and to column 1", "↰"}
	CHA     = AnsiCode{CSI_CHR, 'G', "move cursor to column in current row", "X"}
	CUP     = AnsiCode{CSI_CHR, 'H', "move cursor to row, column", "X"}
	ED      = AnsiCode{CSI_CHR, 'J', "erase display", "X"}
	EL      = AnsiCode{CSI_CHR, 'K', "erase line", "E"}
	IL      = AnsiCode{CSI_CHR, 'L', "insert blank lines", "⋮"}
	DL      = AnsiCode{CSI_CHR, 'M', "delete lines", "⟠"}
	DCH     = AnsiCode{CSI_CHR, 'P', "delete characters on current line", "⇻"}
	ECH     = AnsiCode{CSI_CHR, 'X', "erase characters on current line", "⇸"}
	HPR     = AnsiCode{CSI_CHR, 'a', "move cursor right", "→"}
	DA      = AnsiCode{CSI_CHR, 'c', "Device attributes", "D"}
	VPA     = AnsiCode{CSI_CHR, 'd', "move to row (current column)", "↕"}
	VPR     = AnsiCode{CSI_CHR, 'e', "move cursor down", "↓"}
	HVP     = AnsiCode{CSI_CHR, 'f', "move cursor to row, column", "⇲"}
	SGR     = AnsiCode{CSI_CHR, 'm', "set graphic rendition", "·"}
	DSR     = AnsiCode{CSI_CHR, 'n', "device status report", "⎙"}
	DECSTBM = AnsiCode{CSI_CHR, 'r', "set scrolling region to (top, bottom) rows", "-"}
	CUPSV   = AnsiCode{CSI_CHR, 's', "save cursor position", "⇅"}
	CUPRS   = AnsiCode{CSI_CHR, 'u', "restore cursor position", "⟲"}
	HPA     = AnsiCode{CSI_CHR, '`', "move cursor to column in current row", "↔"}
	TBC     = AnsiCode{CSI_CHR, 'g', "clear tab stop", "↯"}

	OSC = AnsiCode{OSC_CHR, BEL, "OSC", "☓"}

	// Standalone ESC codes
	RIS   = AnsiCode{'c', 0, "Reset", "☓"}
	IND   = AnsiCode{'D', 0, "Line Feed", "☓"}
	NEL   = AnsiCode{'E', 0, "New Line", "☓"}
	HTS   = AnsiCode{'H', 0, "Set Tab Stop", "☓"}
	RI    = AnsiCode{'M', 0, "Reverse Linefeed", "☓"}
	DECID = AnsiCode{'Z', 0, "DEC private identification", "☓"}
	DECSC = AnsiCode{'7', 0, "Save current state (cursor, attrs, chrs set)", "☓"}
	DECRC = AnsiCode{'8', 0, "Restore state (cursor, attrs, chrs set)", "☓"}

	DECPNM = AnsiCode{'>', 0, "Set numeric keypad mode", "☓"}
	DECPAM = AnsiCode{'=', 0, "Set application keypad mode", "☓"}

	// % ones
	ISO8859  = AnsiCode{'%', '@', "Select default (ISO 646 / ISO 8859-1)", "☓"}
	UTF8     = AnsiCode{'%', 'G', "Select UTF-8", "☓"}
	UTF8_OLD = AnsiCode{'%', '8', "Select UTF-8", "☓"}

	// ( ones
	G0MAP_8859  = AnsiCode{'(', 'B', "G0 Select default (ISO 8859-1 mapping)", "☓"}
	G0MAP_VT100 = AnsiCode{'(', '0', "G0 Select VT100 graphics mapping", "☓"}
	G0MAP_NULL  = AnsiCode{'(', 'U', "G0 Select null mapping", "☓"}
	G0MAP_USER  = AnsiCode{'(', 'K', "G0 Select user mapping", "☓"}

	// ) ones
	G1MAP_8859  = AnsiCode{')', 'B', "G1 Select default (ISO 8859-1 mapping)", "☓"}
	G1MAP_VT100 = AnsiCode{')', '0', "G1 Select VT100 graphics mapping", "☓"}
	G1MAP_NULL  = AnsiCode{')', 'U', "G1 Select null mapping", "☓"}
	G1MAP_USER  = AnsiCode{')', 'K', "G1 Select user mapping", "☓"}

)


var ALL_CSI []AnsiCode = []AnsiCode { ICH, CUU, CUD, CUF, CUB, CNL, CPL, CHA, CUP, ED , EL , IL , DL , DCH, ECH, HPR, DA , VPA, VPR, HVP, SGR, DSR, DECSTBM, CUPSV, CUPRS, HPA, TBC}
var ALL_G0 []AnsiCode = []AnsiCode { G0MAP_8859, G0MAP_VT100, G0MAP_NULL, G0MAP_USER}
var ALL_G1 []AnsiCode = []AnsiCode { G1MAP_8859, G1MAP_VT100, G1MAP_NULL, G1MAP_USER}
var ALL_ENCODING []AnsiCode = []AnsiCode { ISO8859, UTF8, UTF8_OLD}
var ALL_SINGLES []AnsiCode = []AnsiCode {RIS, IND, NEL, HTS, RI, DECID, DECSC, DECRC, DECPNM, DECPAM}


// FIX ME those are all manually encoded sequences, they should be AnsiCmd
const READ_CURSOR_POSITION = ESC + "[6n"
const MOVE_CURSOR = ESC + "[%d;%dH"
const CLEAR_SCREEN string = ESC + "[2J"
const CHANGE_SIZE = ESC + "[8;%d;%dt"
const RESET_COLOR = ESC + "[0m"
const RESET string = ESC + "c"

const SMCUP = ESC + "7" + ESC + "[?47h"
const RMCUP = ESC + "[2J" + ESC + "[?47l" + ESC + "8"


func ParseANSI(reader *bufio.Reader) []AnsiCmd {
	var result []AnsiCmd = make([]AnsiCmd, 0)
	for {
		b, _, err := reader.ReadRune()
		if err != nil {
			break
		}
		if b == ESC_CHR {
			b, _, err = reader.ReadRune()
			switch b {
			case CSI_CHR:
				params := ""
				b, _, err = reader.ReadRune()
				for b < 0x40 || b > 0x7E {
					params = fmt.Sprintf("%s%c", params, b)
					b, _, err = reader.ReadRune()
				}
				for _, code := range ALL_CSI {
					if code.Code == b {
						result = append(result, AnsiCmd{0, &code, string(params)})
						break
					}
				}
			case OSC_CHR:
				params := ""
				b, _, err = reader.ReadRune()
				for b != BEL {
					params = fmt.Sprintf("%s%c", params, b)
					b, _, err = reader.ReadRune()
				}
				result = append(result, AnsiCmd{0, &OSC, string(params)})
			case '(':
				b, _, err = reader.ReadRune()
				for _, code := range ALL_G0 {
					if code.Code == b {
						result = append(result, AnsiCmd{0, &code, ""})
						break
					}
				}
			case ')':
				b, _, err = reader.ReadRune()
				for _, code := range ALL_G1 {
					if code.Code == b {
						result = append(result, AnsiCmd{0, &code, ""})
						break
					}
				}
			default:
				for _, code := range ALL_SINGLES {
					if code.Code == b {
						result = append(result, AnsiCmd{0, &code, ""})
						break
					}
				}
				fmt.Printf("PARSING ERROR on Single %c", rune(b))

			}
		} else {
			result = append(result, AnsiCmd{b, nil, ""})
		}
	}
	return result
}

func EdulcorateCharacter(c rune) rune {
	switch c {
	case '\000':
		return '?'
	case '\001':
		return '?'
	case '\002':
		return '?'
	case '\003':
		return '?'
	case '\004':
		return '?'
	case '\005':
		return '?'
	case '\006':
		return '?'
	case '\007':
		return '?'
	case '\010':
		return '?'
	case '\011':
		return '?'
	case '\012':
		return '?'
	case '\013':
		return '?'
	case '\014':
		return '?'
	case '\015':
		return '?'
	case '\016':
		return '?'
	case '\017':
		return '?'
	case '\020':
		return '?'
	case '\021':
		return '?'
	case '\022':
		return '?'
	case '\023':
		return '?'
	case '\024':
		return '?'
	case '\025':
		return '?'
	case '\026':
		return '?'
	case '\027':
		return '?'
	case '\030':
		return '?'
	case '\031':
		return '?'
	case '\032':
		return '?'
	case '\033':
		return '?'
	case '\034':
		return '?'
	case '\035':
		return '?'
	case '\036':
		return '?'
	case '\037':
		return '?'
	default:
		return c
	}
	return '?'
}

func RuneWidth(r rune) int {
	if r >= 0x1100 &&
			(r <= 0x115f || r == 0x2329 || r == 0x232a ||
					(r >= 0x2e80 && r <= 0xa4cf && r != 0x303f) ||
					(r >= 0xac00 && r <= 0xd7a3) ||
					(r >= 0xf900 && r <= 0xfaff) ||
					(r >= 0xfe30 && r <= 0xfe6f) ||
					(r >= 0xff00 && r <= 0xff60) ||
					(r >= 0xffe0 && r <= 0xffe6) ||
					(r >= 0x20000 && r <= 0x2fffd) ||
					(r >= 0x30000 && r <= 0x3fffd)) {
		return 2
	}
	return 1
}
