package midi

import (
	"testing"
)

func TestGetMessageType(t *testing.T) {
	for _, tc := range []struct {
		Expect int
		Input  Packet
		Name   string
	}{
		{
			Expect: MessageTypeNoteOn,
			Input:  Packet{Data: [3]uint8{0x90, 0x4f, 0x16}},
			Name:   "Note On message type",
		},
		{
			Expect: MessageTypeNoteOff,
			Input:  Packet{Data: [3]uint8{0x80, 0x4f, 0x0}},
			Name:   "Note Off message type",
		},
	} {
		if expect, got := tc.Expect, GetMessageType(tc.Input); expect != got {
			t.Fatalf("%s: expected %d, got %d", tc.Name, expect, got)
		}
	}
}
