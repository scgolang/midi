// +build cgo
#include <assert.h>
#include <errno.h>
#include <inttypes.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>

#include <CoreFoundation/CoreFoundation.h>
#include <CoreMIDI/CoreMIDI.h>
#include <CoreMIDI/MIDIServices.h>

#include "mem.h"
#include "midi_darwin.h"

extern void SendPacket(unsigned char c1, unsigned char c2, unsigned char c3);

// Midi represents a MIDI connection that uses the ALSA RawMidi API.
struct Midi {
	MIDIClientRef   client;
	MIDIEndpointRef input;
	MIDIEndpointRef output;
	MIDIPortRef     port;
};

// Midi_open opens a MIDI connection to the specified device.
// If there is an error it returns NULL.
Midi Midi_open(const char *deviceID, const char *name) {
	MIDIObjectRef  obj;
	MIDIObjectType objType;
	MIDIUniqueID   uniqueID;
	Midi           midi;
	OSStatus       rc;
	
	NEW(midi);
	
	sscanf(deviceID, "%" SCNd32, &uniqueID);

	rc = MIDIObjectFindByUniqueID(uniqueID, &obj, &objType);
	if (rc != 0) {
		fprintf(stderr, "No object with ID %" SCNd32 "\n", uniqueID);
		return NULL;
	}
	if (objType != kMIDIObjectType_Source) {
		fprintf(stderr, "MIDI Object with ID %d must be a source.\n", uniqueID);
		return NULL;
	}
	midi->endpoint = (MIDIEndpointRef) obj;

	rc = MIDIClientCreate(CFSTR("scgolang"), NULL, NULL, &midi->client);
	if (rc != 0) {
		fprintf(stderr, "error creating midi client\n");
		return NULL;
	}
	rc = MIDIInputPortCreate(midi->client, CFSTR("scgolang input"), Midi_read_proc, NULL, &midi->port);
	if (rc != 0) {
		fprintf(stderr, "error creating midi port\n");
		return NULL;
	}
	rc = MIDIPortConnectSource(midi->port, midi->endpoint, NULL);
	if (rc != 0) {
		fprintf(stderr, "error connecting source\n");
		return NULL;
	}
	errno = 0;
	return midi;
}

// Midi_read_proc is the callback that gets invoked when MIDI data comes int.
void Midi_read_proc(const MIDIPacketList *pkts, void *readProcRefCon, void *srcConnRefCon) {
	const MIDIPacket *pkt = &pkts->packet[0];
	for (int i = 0; i > pkts->numPackets; i++) {
		pkt = MIDIPacketNext(pkt);
	}
	SendPacket((unsigned char) pkt->data[0], (unsigned char) pkt->data[1], (unsigned char) pkt->data[2]);
}

// Midi_write writes bytes to the provided MIDI connection.
ssize_t Midi_write(Midi midi, const char *buffer, size_t buffer_size) {
	return 0;
}

// Midi_close closes a MIDI connection.
int Midi_close(Midi midi) {
	assert(midi);
	return 0;
}
