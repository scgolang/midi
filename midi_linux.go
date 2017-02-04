// Package midi is a self-contained (i.e. doesn't depend on a C library)
// package for talking to midi devices in Go.
package midi

// #include <stddef.h>
// #include <stdlib.h>
// #include "midi_linux.h"
// #cgo linux LDFLAGS: -lasound
import "C"

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// Packet is a MIDI packet.
type Packet [3]byte

// Device provides an interface for MIDI devices.
type Device struct {
	Name      string
	QueueSize int

	conn C.Midi
	buf  []byte
}

// Open opens a MIDI device.
func (d *Device) Open() error {
	result := C.Midi_open(C.CString(d.Name))
	if result.error != 0 {
		return errors.Errorf("error opening device %d", result.error)
	}
	d.conn = result.midi
	return nil
}

// Close closes the MIDI connection.
func (d *Device) Close() error {
	_, err := C.Midi_close(d.conn)
	return err
}

// Packets returns a read-only channel that emits packets.
func (d *Device) Packets() (<-chan Packet, error) {
	var (
		buf = make([]byte, 3)
		ch  = make(chan Packet, d.QueueSize)
	)
	go func() {
		for {
			if _, err := d.Read(buf); err != nil {
				fmt.Fprintf(os.Stderr, "could not read from device: %s", err)
				close(ch)
				return
			}
			ch <- Packet{buf[0], buf[1], buf[2]}
		}
	}()
	return ch, nil
}

// Read reads data from a MIDI device.
// Note that this method  is only available on Linux.
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
