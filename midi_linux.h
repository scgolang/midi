// +build cgo
#ifndef MIDI_H
#define MIDI_H

#include <stddef.h>
#include <stdlib.h>
#include <unistd.h>

// Midi represents a connection to a MIDI device.
typedef struct Midi *Midi;

// Midi_open_result enables us to return both the Midi instance and an error from Midi_open.
typedef struct Midi_open_result {
	Midi midi;
	int  error;
} Midi_open_result;

// Midi_open opens a MIDI connection to the specified device.
Midi_open_result Midi_open(const char *name);

// Midi_read reads bytes from the provided MIDI connection.
ssize_t Midi_read(Midi midi, char *buffer, size_t buffer_size);

// Midi_write writes bytes to the provided MIDI connection.
ssize_t Midi_write(Midi midi, const char *buffer, size_t buffer_size);

// Midi_close closes a MIDI connection.
int Midi_close(Midi midi);

#endif
