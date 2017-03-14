package midi

// #include <alsa/asoundlib.h>
// #include <stddef.h>
// #include <stdlib.h>
// #include "midi_linux.h"
// #cgo linux LDFLAGS: -lasound
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/pkg/errors"
)

// Device provides an interface for MIDI devices.
type Device struct {
	ID        string
	Name      string
	QueueSize int
	Type      DeviceType

	conn C.Midi
}

// Open opens a MIDI device.
func (d *Device) Open() error {
	var (
		id     = C.CString(d.ID)
		result = C.Midi_open(id)
	)
	defer C.free(unsafe.Pointer(id))

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
				ch <- Packet{Err: err}
				return
			}
			ch <- Packet{
				Data: [3]byte{buf[0], buf[1], buf[2]},
			}
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
	cs := C.CString(string(buf))
	n, err := C.Midi_write(d.conn, cs, C.size_t(len(buf)))
	C.free(unsafe.Pointer(cs))
	return int(n), err
}

// Devices returns a list of devices.
func Devices() ([]*Device, error) {
	var card C.int = -1

	if rc := C.snd_card_next(&card); rc != 0 {
		return nil, alsaMidiError(rc)
	}
	if card < 0 {
		return nil, errors.New("no sound card found")
	}
	devices := []*Device{}

	for {
		cardDevices, err := getCardDevices(card)
		if err != nil {
			return nil, err
		}
		devices = append(devices, cardDevices...)

		if rc := C.snd_card_next(&card); rc != 0 {
			return nil, alsaMidiError(rc)
		}
		if card < 0 {
			break
		}
	}
	return devices, nil
}

func getCardDevices(card C.int) ([]*Device, error) {
	var (
		ctl  *C.snd_ctl_t
		name = C.CString(fmt.Sprintf("hw:%d", card))
	)
	defer C.free(unsafe.Pointer(name))

	if rc := C.snd_ctl_open(&ctl, name, 0); rc != 0 {
		return nil, alsaMidiError(rc)
	}
	var (
		cardDevices       = []*Device{}
		device      C.int = -1
	)
	for {
		if rc := C.snd_ctl_rawmidi_next_device(ctl, &device); rc != 0 {
			return nil, alsaMidiError(rc)
		}
		if device < 0 {
			break
		}
		deviceDevices, err := getDeviceDevices(ctl, card, C.uint(device))
		if err != nil {
			return nil, err
		}
		cardDevices = append(cardDevices, deviceDevices...)
	}
	if rc := C.snd_ctl_close(ctl); rc != 0 {
		return nil, alsaMidiError(rc)
	}
	return cardDevices, nil
}

func getDeviceDevices(ctl *C.snd_ctl_t, card C.int, device C.uint) ([]*Device, error) {
	var info *C.snd_rawmidi_info_t
	C.snd_rawmidi_info_malloc(&info)
	C.snd_rawmidi_info_set_device(info, device)

	defer C.snd_rawmidi_info_free(info)

	// Get inputs.
	var subsIn C.uint
	C.snd_rawmidi_info_set_stream(info, C.SND_RAWMIDI_STREAM_INPUT)
	if rc := C.snd_ctl_rawmidi_info(ctl, info); rc != 0 {
		return nil, alsaMidiError(rc)
	}
	subsIn = C.snd_rawmidi_info_get_subdevices_count(info)

	// Get outputs.
	var subsOut C.uint
	C.snd_rawmidi_info_set_stream(info, C.SND_RAWMIDI_STREAM_OUTPUT)
	if rc := C.snd_ctl_rawmidi_info(ctl, info); rc != 0 {
		return nil, alsaMidiError(rc)
	}
	subsOut = C.snd_rawmidi_info_get_subdevices_count(info)

	// List subdevices.
	var subs C.uint
	if subsIn > subsOut {
		subs = subsIn
	} else {
		subs = subsOut
	}
	if subs == C.uint(0) {
		return nil, errors.New("no streams")
	}
	devices := []*Device{}

	for sub := C.uint(0); sub < subs; sub++ {
		subDevice, err := getSubdevice(ctl, info, card, device, sub, subsIn, subsOut)
		if err != nil {
			return nil, err
		}
		devices = append(devices, subDevice)
	}
	return devices, nil
}

func getSubdevice(ctl *C.snd_ctl_t, info *C.snd_rawmidi_info_t, card C.int, device, sub, subsIn, subsOut C.uint) (*Device, error) {
	if sub < subsIn {
		C.snd_rawmidi_info_set_stream(info, C.SND_RAWMIDI_STREAM_INPUT)
	} else {
		C.snd_rawmidi_info_set_stream(info, C.SND_RAWMIDI_STREAM_OUTPUT)
	}
	C.snd_rawmidi_info_set_subdevice(info, sub)
	if rc := C.snd_ctl_rawmidi_info(ctl, info); rc != 0 {
		return nil, alsaMidiError(rc)
	}
	var (
		name    = C.GoString(C.snd_rawmidi_info_get_name(info))
		subName = C.GoString(C.snd_rawmidi_info_get_subdevice_name(info))
	)
	var dt DeviceType
	if sub < subsIn && sub >= subsOut {
		dt = DeviceInput
	} else if sub >= subsIn && sub < subsOut {
		dt = DeviceOutput
	} else {
		dt = DeviceDuplex
	}
	if sub == 0 && len(subName) > 0 && subName[0] == 0 {
		return &Device{
			ID:   fmt.Sprintf("hw:%d,%d", card, device),
			Name: name,
			Type: dt,
		}, nil
	}
	return &Device{
		ID:   fmt.Sprintf("hw:%d,%d,%d", card, device, sub),
		Name: subName,
		Type: dt,
	}, nil
}

func alsaMidiError(code C.int) error {
	if code == C.int(0) {
		return nil
	}
	return errors.New(C.GoString(C.snd_strerror(code)))
}
