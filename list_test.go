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
	fmt.Printf("devices %#v\n", devices)
}
