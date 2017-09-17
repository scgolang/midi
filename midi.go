// Package midi is a package for talking to midi devices in Go.
package midi

// Packet is a MIDI packet.
type Packet struct {
	Data [3]byte
	Err  error
}

// DeviceType is a flag that says if a device is an input, an output, or duplex.
type DeviceType int

// Device types.
const (
	DeviceInput DeviceType = iota
	DeviceOutput
	DeviceDuplex
)

// Note represents a MIDI note.
type Note struct {
	Number   int
	Velocity int
}

// CC represents a MIDI control change message.
type CC struct {
	Number int
	Value  int
}

const (
	MessageTypeUnknown = iota
	MessageTypeCC
	MessageTypeNoteOff
	MessageTypeNoteOn
	MessageTypePolyKeyPressure
)

// GetMessageType returns the message type for the provided packet.
func GetMessageType(p Packet) int {
	switch p.Data[0] & 0xF0 {
	case 0x80:
		return MessageTypeNoteOff
	case 0x90:
		return MessageTypeNoteOn
	default:
		return MessageTypeUnknown
	}
}
