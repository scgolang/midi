// +build cgo
#ifndef MIDI_H
#define MIDI_H

#include <stddef.h>
#include <stdlib.h>
#include <unistd.h>

#include <CoreMIDI/CoreMIDI.h>

// Midi represents a connection to a MIDI device.
typedef struct Midi *Midi;

// Midi_open opens a MIDI connection to the specified device.
Midi Midi_open(const char *inputID, const char *outputID, const char *name);

// Midi_read_proc is the callback that gets invoked when MIDI data comes in.
void Midi_read_proc(const MIDIPacketList *pkts, void *readProcRefCon, void *srcConnRefCon);

// Midi_write writes bytes to the provided MIDI connection.
ssize_t Midi_write(Midi midi, const char *buffer, size_t buffer_size);

// Midi_close closes a MIDI connection.
int Midi_close(Midi midi);

#endif
