#ifndef MIDI_DARWIN_H
#define MIDI_DARWIN_H

#include <stddef.h>
#include <stdlib.h>
#include <unistd.h>

#include <CoreMIDI/CoreMIDI.h>

// InsufficientSpaceInPacket occurs when you attempt to add a packet
// to a MIDIPacketList which doesn't have enough space to hold the packet.
// Defining this const here causes link errors.
// See the coreMidiError helper in midi_darwin.go [briansorahan]
/* const OSStatus InsufficientSpaceInPacket = -10900; */

// Midi represents a connection to a MIDI device.
typedef struct Midi *Midi;

// Midi_open_result enables us to return both the Midi instance and an error from Midi_open.
typedef struct Midi_open_result {
	Midi     midi;
	OSStatus error;
} Midi_open_result;

typedef struct Midi_device_endpoints {
	MIDIDeviceRef   device;
	MIDIEndpointRef input;
	MIDIEndpointRef output;
	OSStatus        error;
} Midi_device_endpoints;

// Midi_open opens a MIDI connection to the specified device.
Midi_open_result Midi_open(MIDIEndpointRef input, MIDIEndpointRef output);

// Midi_read_proc is the callback that gets invoked when MIDI data comes in.
void Midi_read_proc(const MIDIPacketList *pkts, void *readProcRefCon, void *srcConnRefCon);

// Midi_write_result enables us to return both a ByteCount and an error from Midi_write.
typedef struct Midi_write_result {
	ByteCount n;
	OSStatus  error;
} Midi_write_result;

// Midi_write writes bytes to the provided MIDI connection.
Midi_write_result Midi_write(Midi midi, const char *buffer, size_t buffer_size);

// Midi_close closes a MIDI connection.
int Midi_close(Midi midi);

// CFStringToUTF8 converts a CFStringRef to a UTF8-encoded C string.
char *CFStringToUTF8(CFStringRef aString);

#endif // MIDI_DARWIN_H defined
