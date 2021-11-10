//go:build !windows

package progress

func (w *writer) clearLine() error {
	_, err := w.buffer.Write([]byte("\r\u001b[2K\r"))
	return err
}
