package utils

import (
	"bytes"
	"compress/gzip"
	"io"
)

func CompressString(input string) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte(input))
	if err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecompressString(compressed []byte) (string, error) {
	buf := bytes.NewReader(compressed)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return "", err
	}
	decompressed, err := io.ReadAll(gz)
	if err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}
	return string(decompressed), nil
}
