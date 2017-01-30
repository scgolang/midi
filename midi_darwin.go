// Package midi is a self-contained (i.e. doesn't depend on a C library)
// package for talking to midi devices in Go.
package midi

// #include <stddef.h>
// #include <stdlib.h>
// #include <CoreMIDI/CoreMIDI.h>
// #include "midi_darwin.h"
// #cgo darwin LDFLAGS: -framework CoreFoundation -framework CoreMIDI
import "C"

import (
	"sync"

	"github.com/pkg/errors"
)

// Common errors.
var (
	ErrDeviceNotFound = errors.New("device not found, did you open the device?")
)

// Packet is a MIDI packet.
type Packet [3]byte

var (
	packetChans      = map[*Device]chan Packet{}
	packetChansMutex sync.RWMutex
)

// Device provides an interface for MIDI devices.
type Device struct {
	Name string

	// QueueSize controls the buffer size of the read channel. Use 0 for blocking reads.
	QueueSize int

	conn C.Midi
}

// Open opens a MIDI device.
// queueSize is the number of packets to buffer in the channel associated with the device.
func (d *Device) Open() error {
	result := C.Midi_open(C.CString(d.Name))
	if result.error != 0 {
		return coreMidiError(result.error)
	}
	d.conn = result.midi
	packetChansMutex.Lock()
	packetChans[d] = make(chan Packet, d.QueueSize)
	packetChansMutex.Unlock()
	return nil
}

// Close closes the connection to the MIDI device.
func (d *Device) Close() error {
	return coreMidiError(C.OSStatus(C.Midi_close(d.conn)))
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
	result := C.Midi_write(d.conn, C.CString(string(buf)), C.size_t(len(buf)))
	if result.error != 0 {
		return 0, coreMidiError(result.error)
	}
	return len(buf), nil
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

// coreMidiError maps a CoreMIDI error code to a Go error.
func coreMidiError(code C.OSStatus) error {
	switch code {
	case 0:
		return nil
	case C.kMIDIInvalidClient:
		return errors.New("an invalid MIDIClientRef was passed")
	case C.kMIDIInvalidPort:
		return errors.New("an invalid MIDIPortRef was passed")
	case C.kMIDIWrongEndpointType:
		return errors.New("a source endpoint was passed to a function expecting a destination, or vice versa")
	case C.kMIDINoConnection:
		return errors.New("attempt to close a non-existant connection")
	case C.kMIDIUnknownEndpoint:
		return errors.New("an invalid MIDIEndpointRef was passed")
	case C.kMIDIUnknownProperty:
		return errors.New("attempt to query a property not set on the object")
	case C.kMIDIWrongPropertyType:
		return errors.New("attempt to set a property with a value not of the correct type")
	case C.kMIDINoCurrentSetup:
		return errors.New("there is no current MIDI setup object")
	case C.kMIDIMessageSendErr:
		return errors.New("communication with MIDIServer failed")
	case C.kMIDIServerStartErr:
		return errors.New("unable to start MIDIServer")
	case C.kMIDISetupFormatErr:
		return errors.New("unable to read the saved state")
	case C.kMIDIWrongThread:
		return errors.New("a driver is calling a non-I/O function in the server from a thread other than the server's main thread")
	case C.kMIDIObjectNotFound:
		return errors.New("the requested object does not exist")
	case C.kMIDIIDNotUnique:
		return errors.New("attempt to set a non-unique kMIDIPropertyUniqueID on an object")
	case C.kMIDINotPermitted:
		return errors.New("attempt to perform an operation that is not permitted")
	case -10900:
		// See Midi_write in midi_darwin.c if you're curious where the number comes from.
		// [briansorahan] I tried to add a const to midi_darwin.h for this number, but it
		// resulted in link errors:
		// duplicate symbol _kInsufficientSpaceInPacket in:
		//     $WORK/github.com/scgolang/midi/_test/_obj_test/_cgo_export.o
		//     $WORK/github.com/scgolang/midi/_test/_obj_test/midi_darwin.cgo2.o
		// duplicate symbol _kInsufficientSpaceInPacket in:
		//     $WORK/github.com/scgolang/midi/_test/_obj_test/_cgo_export.o
		//     $WORK/github.com/scgolang/midi/_test/_obj_test/midi_darwin.o
		// ld: 2 duplicate symbols for architecture x86_64
		return errors.New("insufficient space in packet")
	default:
		return errors.Errorf("unknown CoreMIDI error: %d", code)
	}
}
