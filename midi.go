// Package midi is a self-contained (i.e. doesn't depend on a C library)
// package for talking to midi devices in Go.
package midi

// #include <stddef.h>
// #include <stdlib.h>
// #include "midi.h"
// #cgo linux LDFLAGS: -lasound
import "C"

// Midi provides an interface for raw MIDI devices.
type Midi struct {
	conn C.Midi
	buf  []byte
}

// Open opens a MIDI device.
func Open(device string) (*Midi, error) {
	conn := C.Midi_open(C.CString(device))
	return &Midi{conn: conn}, nil
}

// Close closes the MIDI connection.
func (midi *Midi) Close() error {
	return nil
}

// Read reads data from a MIDI device.
func (midi *Midi) Read(buf []byte) (int, error) {
	n, err := C.Midi_read(midi.conn, C.CString(string(buf)), C.size_t(len(buf)))
	return int(n), err
}

// Write writes data to a MIDI device.
func (midi *Midi) Write(buf []byte) (int, error) {
	n, err := C.Midi_write(midi.conn, C.CString(string(buf)), C.size_t(len(buf)))
	return int(n), err
}
