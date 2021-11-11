//go:build windows

package term

import (
	"os"
	"syscall"
	"unsafe"
)

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
