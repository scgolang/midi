// +build cgo
#include <assert.h>
#include <stddef.h>

#include <alsa/asoundlib.h>

#include "mem.h"
#include "midi_linux.h"

// Midi represents a MIDI connection that uses the ALSA RawMidi API.
struct Midi {
	snd_rawmidi_t *in;
	snd_rawmidi_t *out;
};

// Midi_open opens a MIDI connection to the specified device.
// If there is an error it returns NULL.
Midi_open_result Midi_open(const char *name) {
	Midi midi;
	int  rc;
	
	NEW(midi);
	
	rc = snd_rawmidi_open(&midi->in, &midi->out, name, SND_RAWMIDI_SYNC);
	if (rc != 0) {
		return (Midi_open_result) { .midi = NULL, .error = rc };
	}
	return (Midi_open_result) { .midi = midi, .error = 0 };
}

// Midi_read reads bytes from the provided MIDI connection.
ssize_t Midi_read(Midi midi, char *buffer, size_t buffer_size) {
	assert(midi);
	assert(midi->in);
	return snd_rawmidi_read(midi->in, (void *) buffer, buffer_size);
}

// Midi_write writes bytes to the provided MIDI connection.
ssize_t Midi_write(Midi midi, const char *buffer, size_t buffer_size) {
	assert(midi);
	assert(midi->out);
	return snd_rawmidi_write(midi->out, (void *) buffer, buffer_size);
}

// Midi_close closes a MIDI connection.
int Midi_close(Midi midi) {
	assert(midi);
	assert(midi->in);
	assert(midi->out);
	
	int inrc, outrc;
	
	inrc = snd_rawmidi_close(midi->in);
	outrc = snd_rawmidi_close(midi->out);

	if (inrc != 0) {
		return inrc;
	}
	return outrc;
}
