package main

import (
	"errors"
	"fmt"
	"runtime"

	axe "github.com/JamesDunne/axewitcher"
	"github.com/JamesDunne/rpi-egl/bcm"
	gl "github.com/JamesDunne/rpi-egl/gles2"
	"github.com/gvalkov/golang-evdev"
)

func main() {
	// Lock main goroutine to OS thread for EGL safety:
	runtime.LockOSThread()

	// Listen for footswitch events:
	fswDev := FindDeviceByName("PCsensor FootSwitch3")
	if fswDev == nil {
		panic(errors.New("No footswitch device found"))
	}
	fsw := ListenDevice(fswDev)

	// Listen for touch events:
	touchDev := FindAbsDevice()
	_ = touchDev

	// Set up BCM display directly with an EGL context:
	display, err := bcm.OpenDisplay()
	if err != nil {
		panic(err)
	}
	defer display.Close()

	// Set up GL viewport:
	gl.Viewport(0, 0, int32(display.Width()), int32(display.Height()))

	gl.ClearColor(0.0, 0.0, 0.2, 1.0)

mainloop:
	for {
		// Clear background:
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// Swap current surface to display:
		err = display.SwapBuffers()
		if err != nil {
			panic(err)
		}

		// Await a footswitch event:
		select {
		case ev := <-fsw:
			if ev.Type != evdev.EV_KEY {
				fmt.Println(ev)
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

			fswEvent := axe.FswEvent{
				Fsw:   button,
				State: key.State == evdev.KeyDown,
			}

			switch fswEvent.Fsw {
			case axe.FswReset:
				break mainloop
			}
		}
	}
}
