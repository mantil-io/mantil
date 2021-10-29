package event

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

//Gzip - compress input
func Gzip(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// IsGziped provjerava da li je buffer gzipan
func IsGziped(buf []byte) bool {
	if len(buf) > 2 {
		return buf[0] == 0x1f && buf[1] == 0x8b
	}
	return false
}

//Gunzip - decompress data
func gunzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	buf.Write(data)
	r, err := gzip.NewReader(&buf)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	out, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GunzipIf gunzipa ako je data gzipan
func Gunzip(data []byte) ([]byte, error) {
	if IsGziped(data) {
		return gunzip(data)
	}
	return data, nil
}
