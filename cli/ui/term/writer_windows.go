//go:build windows

package term

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procFillConsoleOutputCharacter = kernel32.NewProc("FillConsoleOutputCharacterW")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	procSetConsoleCursorPosition   = kernel32.NewProc("SetConsoleCursorPosition")
)

type short int16
type dword uint32
type word uint16

type coord struct {
	x short
	y short
}

type smallRect struct {
	bottom short
	left   short
	right  short
	top    short
}

type consoleScreenBufferInfo struct {
	size              coord
	cursorPosition    coord
	attributes        word
	window            smallRect
	maximumWindowSize coord
}

type consoleCursorInfo struct {
	size    dword
	visible int32
}

func (wr *Writer) clearLine() error {
	fd := int(os.Stdout.Fd())
	h := syscall.Handle(fd)
	var csbi consoleScreenBufferInfo
	_, _, _ = procGetConsoleScreenBufferInfo.Call(uintptr(h), uintptr(unsafe.Pointer(&csbi)))
	clearLine(h, csbi)
	moveCursor(h, csbi, 0, -1)
	return nil
}

func clearLine(handle syscall.Handle, csbi consoleScreenBufferInfo) {
	var w uint32
	var x short
	cursor := csbi.cursorPosition
	x = csbi.size.x
	_, _, _ = procFillConsoleOutputCharacter.Call(uintptr(handle), uintptr(' '), uintptr(x), uintptr(*(*int32)(unsafe.Pointer(&cursor))), uintptr(unsafe.Pointer(&w)))
}

func moveCursor(handle syscall.Handle, csbi consoleScreenBufferInfo, x, y int) {
	var cursor coord
	cursor.x = csbi.cursorPosition.x + short(x)
	cursor.y = csbi.cursorPosition.y + short(y)
	_, _, _ = procSetConsoleCursorPosition.Call(uintptr(handle), uintptr(*(*int32)(unsafe.Pointer(&cursor))))
}
