package midi

import (
	"fmt"
	"testing"
)

func TestDevices(t *testing.T) {
	devices, err := Devices()
	if err != nil {
		t.Fatal(err)
	}
	for i, d := range devices {
		if d == nil {
			continue
		}
		fmt.Printf("device %d: %#v\n", i, *d)
	}
}
