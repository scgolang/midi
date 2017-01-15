// Package midi is a self-contained (i.e. doesn't depend on a C library)
// package for talking to midi devices in Go.
package midi

// #include <stddef.h>
// #include <stdlib.h>
// #include "midi.h"
// #cgo linux LDFLAGS: -lasound
import "C"

// Device provides an interface for MIDI devices.
type Device struct {
	conn C.Midi
	buf  []byte
}

// Open opens a MIDI device.
func Open(deviceID string) (*Device, error) {
	conn, err := C.Midi_open(C.CString(deviceID))
	return &Device{conn: conn}, err
}

// Close closes the MIDI connection.
func (d *Device) Close() error {
	_, err := C.Midi_close(d.conn)
	return err
}

// Read reads data from a MIDI device.
func (d *Device) Read(buf []byte) (int, error) {
	cbuf := make([]C.char, len(buf))
	n, err := C.Midi_read(d.conn, &cbuf[0], C.size_t(len(buf)))
	for i := C.ssize_t(0); i < n; i++ {
		buf[i] = byte(cbuf[i])
	}
	return int(n), err
}

// Write writes data to a MIDI device.
func (d *Device) Write(buf []byte) (int, error) {
	n, err := C.Midi_write(d.conn, C.CString(string(buf)), C.size_t(len(buf)))
	return int(n), err
}
