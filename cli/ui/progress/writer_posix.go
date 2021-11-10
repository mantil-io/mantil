//go:build !windows

package progress

func (w *writer) clearLine() error {
	_, err := w.out.Write([]byte("\033[1A\033[2K\r"))
	return err
}
