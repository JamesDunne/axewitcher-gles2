package main

import (
	"errors"
	"strings"

	axe "github.com/JamesDunne/axewitcher"
	"github.com/gvalkov/golang-evdev"
)

func ListenFootswitch() (ch chan axe.FswEvent, err error) {
	fsw := (*evdev.InputDevice)(nil)
	first := (*evdev.InputDevice)(nil)

	// List all input devices:
	devs, err := evdev.ListInputDevices()
	if err != nil {
		return
	}
	for _, dev := range devs {
		if first == nil {
			first = dev
		}

		// Find foot switch device:
		if strings.Contains(dev.Name, "PCsensor FootSwitch3") {
			fsw = dev
			break
		}
	}
	if fsw == nil {
		fsw = first
	}
	if fsw == nil {
		err = errors.New("No footswitch device found!")
		return
	}

	ch = make(chan axe.FswEvent)

	go func() {
		defer close(ch)

		for {
			ev, err := fsw.ReadOne()
			if err != nil {
				break
			}
			if ev.Type != evdev.EV_KEY {
				continue
			}

			key := evdev.NewKeyEvent(ev)
			if key.State == evdev.KeyHold {
				continue
			}

			// Determine which footswitch was pressed/released:
			// NOTE: unfortunately the footswitch driver does not allow multiple switches to be depressed simultaneously.
			button := axe.FswNone
			if key.Scancode == evdev.KEY_A {
				button = axe.FswReset
			} else if key.Scancode == evdev.KEY_B {
				button = axe.FswPrev
			} else if key.Scancode == evdev.KEY_C {
				button = axe.FswNext
			}

			ch <- axe.FswEvent{
				Fsw:   button,
				State: key.State == evdev.KeyDown,
			}
		}
	}()

	return
}
