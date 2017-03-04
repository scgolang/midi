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
