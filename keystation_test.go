package midi

import (
	"fmt"
	"strings"
	"testing"
)

func TestKeystation(t *testing.T) {
	devices, err := Devices()
	if err != nil {
		t.Fatal(err)
	}
	var keystation *Device
	for _, d := range devices {
		if strings.Contains(strings.ToLower(d.Name), "keystation") {
			keystation = d
			break
		}
	}
	if keystation == nil {
		t.Log("no keystation detected")
		t.SkipNow()
	}
	fmt.Println("keystation detected")

	if err := keystation.Open(); err != nil {
		t.Fatal(err)
	}
	packets, err := keystation.Packets()
	if err != nil {
		t.Fatal(err)
	}
	i := 0
	for pkt := range packets {
		if i == 4 {
			break
		}
		fmt.Printf("%#v\n", pkt)
	}
}
