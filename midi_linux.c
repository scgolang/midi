// +build cgo
#include <assert.h>
#include <errno.h>
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
Midi Midi_open(const char *device_id, const char *name) {
	Midi midi;
	int rc;
	NEW(midi);
	rc = snd_rawmidi_open(&midi->in, &midi->out, device_id, SND_RAWMIDI_SYNC);
	errno = rc; // Not sure if the rawmidi return codes map to errno values.
	if (rc != 0) return NULL;
	return midi;
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
