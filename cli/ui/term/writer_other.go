//go:build !windows

package term

func (w *Writer) clearLine() error {
	_, err := w.out.Write([]byte("\033[1A\033[2K\r"))
	return err
}
