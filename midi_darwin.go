// Package midi is a self-contained (i.e. doesn't depend on a C library)
// package for talking to midi devices in Go.
package midi

// #include <stddef.h>
// #include <stdlib.h>
// #include "midi_darwin.h"
// #cgo darwin LDFLAGS: -framework CoreFoundation -framework CoreMIDI
import "C"

// Packet is a MIDI packet.
type Packet [3]byte

var packetChan = make(chan Packet)

//export SendPacket
func SendPacket(c1 C.uchar, c2 C.uchar, c3 C.uchar) {
	packetChan <- Packet{byte(c1), byte(c2), byte(c3)}
}

// Device provides an interface for MIDI devices.
type Device struct {
	Name string

	conn C.Midi
	buf  []byte
}

// Open opens a MIDI device.
func Open(deviceID, name string) (*Device, error) {
	conn, err := C.Midi_open(C.CString(deviceID), C.CString(name))
	if err != nil {
		return nil, err
	}
	return &Device{Name: name, conn: conn}, nil
}

// Close closes the connection to the MIDI device.
func (d *Device) Close() error {
	_, err := C.Midi_close(d.conn)
	return err
}

// Packets emits MIDI packets.
func (d *Device) Packets() <-chan Packet {
	return packetChan
}

// Write writes data to a MIDI device.
func (d *Device) Write(buf []byte) (int, error) {
	n, err := C.Midi_write(d.conn, C.CString(string(buf)), C.size_t(len(buf)))
	return int(n), err
}
