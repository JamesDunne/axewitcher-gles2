//+build rpi

package main

import (
	"strings"

	axe "github.com/JamesDunne/axewitcher"

	"github.com/JamesDunne/golang-nanovg/nvgui"
	"github.com/gvalkov/golang-evdev"
)

type EventListener struct {
	fsw   <-chan []evdev.InputEvent
	touch <-chan []evdev.InputEvent

	Touches   []nvgui.Touch
	touchSlot int

	FswEvents []axe.FswEvent
}

func InitEventListener() *EventListener {
	ev := &EventListener{
		Touches: make([]nvgui.Touch, 10),
	}

	// Listen for footswitch events:
	fswDev := FindDeviceByName("PCsensor FootSwitch3")
	if fswDev != nil {
		ev.fsw = ListenDevice(fswDev)
	}

	// Listen for touch events:
	touchDev := FindAbsDevice()
	if touchDev != nil {
		ev.touch = ListenDevice(touchDev)
	}

	return ev
}

func (l *EventListener) Await() {
	// Await an event:
	select {
	case evs := <-l.touch:
		// Process touch events with absolute coordinates:
		//fmt.Println("[")
		for _, ev := range evs {
			if ev.Type != evdev.EV_ABS {
				continue
			}

			//fmt.Println(evdev.ABS[int(ev.Code)], ev.Value)

			switch ev.Code {
			case evdev.ABS_MT_SLOT:
				l.touchSlot = int(ev.Value)
			case evdev.ABS_MT_POSITION_X:
				l.Touches[l.touchSlot].X = float32(ev.Value)
			case evdev.ABS_MT_POSITION_Y:
				l.Touches[l.touchSlot].Y = float32(ev.Value)
			case evdev.ABS_MT_TRACKING_ID:
				l.Touches[l.touchSlot].ID = ev.Value
			}
		}
		//fmt.Println("]")

	case evs := <-l.fsw:
		fswEvents := make([]axe.FswEvent, 0, len(evs))

		// Process footswitch (keyboard) events:
		for i := range evs {
			ev := &evs[i]
			if ev.Type != evdev.EV_KEY {
				//fmt.Println(ev)
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

			fswEvents = append(fswEvents, axe.FswEvent{
				Fsw:   button,
				State: key.State == evdev.KeyDown,
			})
		}
		l.FswEvents = fswEvents
	}
}

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
		for key := range dev.Capabilities {
			if key.Type == evdev.EV_ABS {
				return dev
			}
		}
	}

	return nil
}

func ListenDevice(dev *evdev.InputDevice) (ch chan []evdev.InputEvent) {
	ch = make(chan []evdev.InputEvent)

	go func() {
		defer close(ch)

		for {
			evs, err := dev.Read()
			if err != nil {
				break
			}

			ch <- evs
		}
	}()

	return
}
