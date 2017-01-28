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
#include <mach/mach_time.h>

#include "mem.h"
#include "midi_darwin.h"

extern void SendPacket(Midi midi, unsigned char c1, unsigned char c2, unsigned char c3);

MIDIEndpointRef get_endpoint_by_id(const char *id,  MIDIObjectType expectedObjType, const char *typeName);

// Midi represents a MIDI connection that uses the ALSA RawMidi API.
struct Midi {
	MIDIClientRef   client;
	MIDIEndpointRef input;
	MIDIEndpointRef output;
	MIDIPortRef     inputPort;
	MIDIPortRef     outputPort;
};

// Midi_open opens a MIDI connection to the specified device.
// If there is an error it returns NULL.
Midi Midi_open(const char *inputID, const char *outputID, const char *name) {
	Midi           midi;
	OSStatus       rc;
	
	NEW(midi);

	// Read input and output endpoints.
	midi->input  = get_endpoint_by_id(inputID, kMIDIObjectType_Source, "source");
	midi->output = get_endpoint_by_id(outputID, kMIDIObjectType_Destination, "destination");

	rc = MIDIClientCreate(CFSTR("scgolang"), NULL, NULL, &midi->client);
	if (rc != 0) {
		fprintf(stderr, "error creating midi client\n");
		return NULL;
	}
	rc = MIDIInputPortCreate(midi->client, CFSTR("scgolang input"), Midi_read_proc, NULL, &midi->inputPort);
	if (rc != 0) {
		fprintf(stderr, "error creating midi input port\n");
		return NULL;
	}
	rc = MIDIOutputPortCreate(midi->client, CFSTR("scgolang output"), &midi->outputPort);
	if (rc != 0) {
		fprintf(stderr, "error creating midi output port\n");
		return NULL;
	}
	rc = MIDIPortConnectSource(midi->inputPort, midi->input, midi);
	if (rc != 0) {
		fprintf(stderr, "error connecting source\n");
		return NULL;
	}
	errno = 0;
	return midi;
}

MIDIEndpointRef get_endpoint_by_id(const char *id,  MIDIObjectType expectedObjType, const char *typeName) {
	MIDIObjectRef  obj;
	MIDIObjectType objType;
	OSStatus       rc;
	MIDIUniqueID   uniqueID;
	
	sscanf(id, "%" SCNd32, &uniqueID);
	rc = MIDIObjectFindByUniqueID(uniqueID, &obj, &objType);
	if (rc != 0) {
		fprintf(stderr, "No object with ID %" SCNd32 "\n", uniqueID);
		return rc;
	}
	if (objType != expectedObjType) {
		fprintf(stderr, "MIDI Object with ID %d must be a %s.\n", uniqueID, typeName);
		return rc;
	}
	return (MIDIEndpointRef) obj;
}

// Midi_read_proc is the callback that gets invoked when MIDI data comes int.
void Midi_read_proc(const MIDIPacketList *pkts, void *readProcRefCon, void *srcConnRefCon) {
	const MIDIPacket *pkt = &pkts->packet[0];

	Midi midi = (Midi) srcConnRefCon;
	
	for (int i = 0; i > pkts->numPackets; i++) {
		pkt = MIDIPacketNext(pkt);
	}
	SendPacket(midi, (unsigned char) pkt->data[0], (unsigned char) pkt->data[1], (unsigned char) pkt->data[2]);
}

// Midi_write writes bytes to the provided MIDI connection.
ssize_t Midi_write(Midi midi, const char *buffer, size_t buffer_size) {
	assert(midi);

	MIDIPacketList pkts;
	MIDIPacket    *cur        = MIDIPacketListInit(&pkts);
	MIDITimeStamp  now        = mach_absolute_time();
	size_t         numPackets = buffer_size / 3;
	ByteCount      listSize   = numPackets * 32;

	for (size_t i = 0; i < numPackets; i++) {
		Byte data[3];
		for (int j = 0; j < 3; j++) {
			data[j] = buffer[i+j];
		}
		cur = MIDIPacketListAdd(&pkts, listSize, cur, now, 3, data);
		if (cur == NULL) {
			fprintf(stderr, "error adding packet to list\n");
			return 0;
		}
	}
	OSStatus rc = MIDISend(midi->outputPort, midi->output, &pkts);
	if (rc != 0) {
		errno = rc;
		return 0;
	}
	return listSize;
}

// Midi_close closes a MIDI connection.
int Midi_close(Midi midi) {
	assert(midi);
	return 0;
}
