package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/scgolang/midi"
)

// Config represents the application's configuration.
type Config struct {
	List bool
}

func main() {
	var config Config

	flag.BoolVar(&config.List, "l", false, "list midi devices")
	flag.Parse()

	exitCode := 0

	if config.List {
		exitCode = list()
	}
	os.Exit(exitCode)
}

func list() int {
	devices, err := midi.Devices()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	fmt.Printf("found %d devices\n", len(devices))

	for _, device := range devices {
		fmt.Printf("%s %s %s\n", device.ID, device.Name, device.Type)
	}
	return 0
}
