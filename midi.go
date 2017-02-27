// Package midi is a self-contained (i.e. doesn't depend on a C library)
// package for talking to midi devices in Go.
package midi

// Packet is a MIDI packet.
type Packet [3]byte

// DeviceType is a flag that says if a device is an input, an output, or duplex.
type DeviceType int

const (
	DeviceInput DeviceType = iota
	DeviceOutput
	DeviceDuplex
)
