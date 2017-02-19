// Package midi is a self-contained (i.e. doesn't depend on a C library)
// package for talking to midi devices in Go.
package midi

// #include <alsa/asoundlib.h>
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

type Stream struct {
	Name string
}

type DeviceInfo struct {
	ID      string
	Name    string
	Inputs  []Stream
	Outputs []Stream
}

func Devices() ([]DeviceInfo, error) {
	var (
		seqp  *C.snd_seq_t
		cinfo *C.snd_seq_client_info_t
		pinfo *C.snd_seq_port_info_t
	)
	if rc := C.snd_seq_open(&seqp, C.CString("default"), C.SND_SEQ_OPEN_DUPLEX, 0); rc != 0 {
		return nil, alsaMidiError(rc)
	}
	if rc := C.snd_seq_client_info_malloc(&cinfo); rc != 0 {
		return nil, alsaMidiError(rc)
	}

	if rc := C.snd_seq_port_info_malloc(&pinfo); rc != 0 {
		return nil, alsaMidiError(rc)
	}
	C.snd_seq_client_info_set_client(cinfo, -1)

	devices := []DeviceInfo{}

	for C.snd_seq_query_next_client(seqp, cinfo) == 0 {
		C.snd_seq_port_info_set_client(pinfo, C.snd_seq_client_info_get_client(cinfo))
		C.snd_seq_port_info_set_port(pinfo, -1)

		if clientID := C.snd_seq_port_info_get_client(pinfo); isSystemClient(clientID) {
			continue // ignore Timer and Announce ports on client 0
		}
		device := DeviceInfo{
			Name: C.GoString(C.snd_seq_client_info_get_name(cinfo)),
		}
		if card := C.snd_seq_client_info_get_card(cinfo); card >= 0 {
			device.ID = fmt.Sprintf("hw:%d", card)
		}
		for C.snd_seq_query_next_port(seqp, pinfo) == 0 {
			if portType := C.snd_seq_port_info_get_type(pinfo); !isMidiPort(portType) {
				fmt.Printf("portType is not a midi port %d\n", portType)
				continue // Not a MIDI port.
			}
			var (
				caps     = C.snd_seq_port_info_get_capability(pinfo)
				portName = C.GoString(C.snd_seq_port_info_get_name(pinfo))
				stream   = Stream{Name: portName}
			)
			if isDuplexPort(caps) {
				// Add duplex stream.
				device.Inputs = append(device.Inputs, stream)
				device.Outputs = append(device.Outputs, stream)
			} else if isWritePort(caps) {
				// Add write stream.
				device.Outputs = append(device.Outputs, stream)
			} else if isReadPort(caps) {
				// Add read stream.
				device.Inputs = append(device.Inputs, stream)
			}
		}
		devices = append(devices, device)
	}
	return devices, nil
}

func isSystemClient(clientID C.int) bool {
	return clientID == C.SND_SEQ_CLIENT_SYSTEM
}

func isSystemPort(portID C.int) bool {
	return portID == C.SND_SEQ_PORT_SYSTEM_ANNOUNCE || portID == C.SND_SEQ_PORT_SYSTEM_TIMER
}

func isMidiPort(portType C.uint) bool {
	return (portType&C.SND_SEQ_PORT_TYPE_MIDI_GENERIC != 0) ||
		(portType&C.SND_SEQ_PORT_TYPE_MIDI_GM != 0) ||
		(portType&C.SND_SEQ_PORT_TYPE_MIDI_GS != 0) ||
		(portType&C.SND_SEQ_PORT_TYPE_MIDI_XG != 0) ||
		(portType&C.SND_SEQ_PORT_TYPE_MIDI_MT32 != 0) ||
		(portType&C.SND_SEQ_PORT_TYPE_MIDI_GM2 != 0)
}

func isDuplexPort(caps C.uint) bool {
	return (caps & C.SND_SEQ_PORT_CAP_DUPLEX) != 0
}

func isWritePort(caps C.uint) bool {
	return (caps&C.SND_SEQ_PORT_CAP_WRITE != 0) ||
		(caps&C.SND_SEQ_PORT_CAP_SYNC_WRITE != 0) ||
		(caps&C.SND_SEQ_PORT_CAP_SUBS_WRITE != 0)
}

func isReadPort(caps C.uint) bool {
	return (caps&C.SND_SEQ_PORT_CAP_READ != 0) ||
		(caps&C.SND_SEQ_PORT_CAP_SYNC_READ != 0) ||
		(caps&C.SND_SEQ_PORT_CAP_SUBS_READ != 0)
}

func alsaMidiError(code C.int) error {
	return errors.New(C.GoString(C.snd_strerror(code)))
}
