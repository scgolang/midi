#include <assert.h>
#include <inttypes.h>
#include <stddef.h>
#include <string.h>

#include <CoreFoundation/CoreFoundation.h>
#include <CoreMIDI/CoreMIDI.h>
#include <mach/mach_time.h>

#include "mem.h"
#include "midi_darwin.h"

extern void SendPacket(Midi midi, unsigned char c1, unsigned char c2, unsigned char c3);

Midi_device_endpoints  find_device_endpoints(const char *device);
char                  *CFStringToUTF8(CFStringRef aString);

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
Midi_open_result Midi_open(const char *name) {
	Midi           midi;
	OSStatus       rc;
	
	NEW(midi);

	// Read input and output endpoints.
	Midi_device_endpoints device_endpoints = find_device_endpoints(name);
	if (device_endpoints.error != 0) {
		return (Midi_open_result) { .midi = NULL, .error = device_endpoints.error };
	}
	midi->input  = device_endpoints.input;
	midi->output = device_endpoints.output;

	rc = MIDIClientCreate(CFSTR("scgolang"), NULL, NULL, &midi->client);
	if (rc != 0) {
		return (Midi_open_result) { .midi = NULL, .error = rc };
	}
	rc = MIDIInputPortCreate(midi->client, CFSTR("scgolang input"), Midi_read_proc, NULL, &midi->inputPort);
	if (rc != 0) {
		return (Midi_open_result) { .midi = NULL, .error = rc };
	}
	rc = MIDIOutputPortCreate(midi->client, CFSTR("scgolang output"), &midi->outputPort);
	if (rc != 0) {
		return (Midi_open_result) { .midi = NULL, .error = rc };
	}
	rc = MIDIPortConnectSource(midi->inputPort, midi->input, midi);
	if (rc != 0) {
		return (Midi_open_result) { .midi = NULL, .error = rc };
	}
	return (Midi_open_result) { .midi =  midi, .error = 0 };
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
Midi_write_result Midi_write(Midi midi, const char *buffer, size_t buffer_size) {
	assert(midi);

	MIDIPacketList pkts;
	MIDIPacket    *cur        = MIDIPacketListInit(&pkts);
	MIDITimeStamp  now        = mach_absolute_time();
	size_t         numPackets = (buffer_size / 256) + 1;
	ByteCount      listSize   = numPackets * 256;

	for (size_t i = 0; i < numPackets; i++) {
		Byte data[3];
		for (int j = 0; j < 3; j++) {
			data[j] = buffer[i+j];
		}
		cur = MIDIPacketListAdd(&pkts, listSize, cur, now, 3, data);
		if (cur == NULL) {
			return (Midi_write_result) { .n = 0, .error = -10900 };
		}
	}
	OSStatus rc = MIDISend(midi->outputPort, midi->output, &pkts);
	if (rc != 0) {
		return (Midi_write_result) { .n = 0, .error = rc };
	}
	return (Midi_write_result) { .n = listSize, .error = 0 };
}

// Midi_close closes a MIDI connection.
int Midi_close(Midi midi) {
	assert(midi);

	OSStatus rc1, rc2, rc3;

	rc1 = MIDIPortDispose(midi->inputPort);
	rc2 = MIDIPortDispose(midi->outputPort);
	rc3 = MIDIClientDispose(midi->client);

	if      (rc1 != 0) return rc1;
	else if (rc2 != 0) return rc2;
	else if (rc3 != 0) return rc3;
	else               return 0;
}

Midi_device_endpoints find_device_endpoints(const char *name) {
	ItemCount numDevices = MIDIGetNumberOfDevices();
	OSStatus  rc;

	for (int i = 0; i < numDevices; i++) {
		CFStringRef   deviceName;
		MIDIDeviceRef deviceRef = MIDIGetDevice(i);
		
		rc = MIDIObjectGetStringProperty(deviceRef, kMIDIPropertyName, &deviceName);
		if (rc != 0) {
			return (Midi_device_endpoints) { .device = 0, .input = 0, .output = 0, .error = rc };
		}
		if (strcmp(CFStringToUTF8(deviceName), name) != 0) {
			continue;
		}
		ItemCount numEntities = MIDIDeviceGetNumberOfEntities(deviceRef);
		
		for (int i = 0; i < numEntities; i++) {
			MIDIEntityRef entityRef       = MIDIDeviceGetEntity(deviceRef, i);
			ItemCount     numDestinations = MIDIGetNumberOfDestinations(entityRef);
			ItemCount     numSources      = MIDIGetNumberOfSources(entityRef);

			if (numDestinations < 1 || numSources < 1) {
				continue;
			}
			MIDIEndpointRef input  = MIDIGetSource(0);
			MIDIEndpointRef output = MIDIGetDestination(0);

			return (Midi_device_endpoints) { .device = deviceRef, .input = input, .output = output, .error = 0 };
		}
	}
	return (Midi_device_endpoints) { .device = 0, .input = 0, .output = 0, .error = -10901 };
}

char *CFStringToUTF8(CFStringRef aString) {
	if (aString == NULL) {
		return NULL;
	}

	CFIndex length = CFStringGetLength(aString);
	CFIndex maxSize =
		CFStringGetMaximumSizeForEncoding(length, kCFStringEncodingUTF8) + 1;
	char *buffer = (char *)malloc(maxSize);
	if (CFStringGetCString(aString, buffer, maxSize,
			       kCFStringEncodingUTF8)) {
		return buffer;
	}
	free(buffer); // If we failed
	return NULL;
}
