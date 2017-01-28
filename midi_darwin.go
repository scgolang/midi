// Package midi is a self-contained (i.e. doesn't depend on a C library)
// package for talking to midi devices in Go.
package midi

// #include <stddef.h>
// #include <stdlib.h>
// #include "midi_darwin.h"
// #cgo darwin LDFLAGS: -framework CoreFoundation -framework CoreMIDI
import "C"

import (
	"errors"
	"sync"
)

var ErrDeviceNotFound = errors.New("device not found")

// Packet is a MIDI packet.
type Packet [3]byte

var (
	packetChans      = map[*Device]chan Packet{}
	packetChansMutex sync.RWMutex
)

// Device provides an interface for MIDI devices.
type Device struct {
	Name string

	buf  []byte
	conn C.Midi
}

// Open opens a MIDI device.
func Open(inputID, outputID, name string) (*Device, error) {
	conn, err := C.Midi_open(C.CString(inputID), C.CString(outputID), C.CString(name))
	if err != nil {
		return nil, err
	}
	if conn == nil {
		return nil, errors.New("could not connect to device")
	}
	device := &Device{Name: name, conn: conn}
	packetChansMutex.Lock()
	packetChans[device] = make(chan Packet)
	packetChansMutex.Unlock()
	return device, nil
}

// Close closes the connection to the MIDI device.
func (d *Device) Close() error {
	_, err := C.Midi_close(d.conn)
	return err
}

// Packets emits MIDI packets.
func (d *Device) Packets() (<-chan Packet, error) {
	packetChansMutex.RLock()
	for device, packetChan := range packetChans {
		if d.conn == device.conn {
			packetChansMutex.RUnlock()
			return packetChan, nil
		}
	}
	packetChansMutex.RUnlock()
	return nil, ErrDeviceNotFound
}

// Write writes data to a MIDI device.
func (d *Device) Write(buf []byte) (int, error) {
	n, err := C.Midi_write(d.conn, C.CString(string(buf)), C.size_t(len(buf)))
	return int(n), err
}

//export SendPacket
func SendPacket(conn C.Midi, c1 C.uchar, c2 C.uchar, c3 C.uchar) {
	packetChansMutex.RLock()
	for device, packetChan := range packetChans {
		if device.conn == conn {
			packetChan <- Packet{byte(c1), byte(c2), byte(c3)}
		}
	}
	packetChansMutex.RUnlock()
}
