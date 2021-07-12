package util

import (
	"os"
	"testing"
)

func TestSetAndCheckPassword(t *testing.T) {
	err := SetPassword(os.TempDir(), "password", "salt")

	if err != nil {
		t.Fatalf(err.Error())
	}

	if !CheckPassword(os.TempDir(), "password", "salt") {
		t.Fatalf("password set wrong")
	}
}

func TestGetPassword(t *testing.T) {}
