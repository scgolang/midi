package midi

import (
	"testing"
)

func TestDevices(t *testing.T) {
	if _, err := Devices(); err != nil {
		t.Fatal(err)
	}
}
