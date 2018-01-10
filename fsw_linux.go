package main

import (
	"fmt"
	"strings"

	"github.com/gvalkov/golang-evdev"
)

func FindDeviceByName(name string) *evdev.InputDevice {
	// List all input devices:
	devs, err := evdev.ListInputDevices()
	if err != nil {
		return nil
	}
	for _, dev := range devs {
		// Find foot switch device:
		if strings.Contains(dev.Name, name) {
			return dev
		}
	}

	return nil
}

func FindAbsDevice() *evdev.InputDevice {
	// List all input devices:
	devs, err := evdev.ListInputDevices()
	if err != nil {
		return nil
	}
	for _, dev := range devs {
		fmt.Println(dev.Name, dev.Capabilities)
	}

	return nil
}

func ListenDevice(dev *evdev.InputDevice) (ch chan *evdev.InputEvent) {
	ch = make(chan *evdev.InputEvent)

	go func() {
		defer close(ch)

		for {
			ev, err := dev.ReadOne()
			if err != nil {
				break
			}

			ch <- ev
		}
	}()

	return
}
