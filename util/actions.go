package util

import (
	"errors"
	"io"
)

func ActionAdd(name, data string, writer io.Writer) error {
	if name == "" {
		return errors.New("data name must be not empty string")
	}

	if data == "" {
		return errors.New("data name must be not empty string")
	}

	_, err := writer.Write([]byte(data))

	return err
}

func ActionGet(name string, reader io.Reader) (string, error) {
	var p []byte
	_, err := reader.Read(p)
	return string(p), err
}

func ActionEdit(name, data string) error {
	return nil
}

func ActionDelete(name string) error {
	return nil
}
