package scriptedit

import (
	"fmt"
	"syscall"
	"unsafe"
)

// termios types
type cc_t byte
type speed_t uint
type tcflag_t uint

// termios constants
const (
	BRKINT = tcflag_t(0000002)
	ICRNL  = tcflag_t(0000400)
	INPCK  = tcflag_t(0000020)
	ISTRIP = tcflag_t(0000040)
	IXON   = tcflag_t(0002000)
	OPOST  = tcflag_t(0000001)
	CS8    = tcflag_t(0000060)
	ECHO   = tcflag_t(0000010)
	ICANON = tcflag_t(0000002)
	IEXTEN = tcflag_t(0100000)
	ISIG   = tcflag_t(0000001)
	VTIME  = tcflag_t(5)
	VMIN   = tcflag_t(6)
)

const NCCS = 32

type Termios struct {
	c_iflag,  c_oflag,  c_cflag,  c_lflag tcflag_t
	c_line                                cc_t
	c_cc[NCCS] cc_t
	c_ispeed,  c_ospeed                   speed_t
}

// ioctl constants
const (
	TCGETS = 0x5401
	TCSETS = 0x5402
)

func GetTermios(dst *Termios, ttyfd int) error {
	r1, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(ttyfd), uintptr(TCGETS), uintptr(unsafe.Pointer(dst)))
	if errno != 0 {
		return errno
	}

	if r1 != 0 {
		fmt.Println("Error r1", r1)
	}

	return nil
}

func SetTermios(src *Termios, ttyfd int) error {
	r1, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(ttyfd), uintptr(TCSETS), uintptr(unsafe.Pointer(src)))
	if errno != 0 {
		return errno
	}

	if r1 != 0 {
		fmt.Println("Error r1", r1)
	}

	return nil
}

func Tty_raw(raw *Termios, ttyfd int) error {

	raw.c_iflag &= ^(BRKINT | ICRNL | INPCK | ISTRIP | IXON)
	raw.c_oflag &= ^(OPOST)
	raw.c_cflag |= (CS8)
	raw.c_lflag &= ^(ECHO | ICANON | IEXTEN | ISIG)

	raw.c_cc[VMIN] = 1
	raw.c_cc[VTIME] = 0

	err := SetTermios(raw, ttyfd)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}


