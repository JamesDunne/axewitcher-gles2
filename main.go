package main

import (
	"fmt"
	"runtime"
	"strings"

	axe "github.com/JamesDunne/axewitcher"
	"github.com/JamesDunne/rpi-egl/bcm"
	gl "github.com/JamesDunne/rpi-egl/gles2"
	"github.com/gvalkov/golang-evdev"

	"github.com/JamesDunne/golang-nanovg/nvg"
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
		for key := range dev.Capabilities {
			if key.Type == evdev.EV_ABS {
				return dev
			}
		}
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

func main() {
	// Lock main goroutine to OS thread for EGL safety:
	runtime.LockOSThread()

	var (
		fsw   <-chan *evdev.InputEvent
		touch <-chan *evdev.InputEvent
	)
	// Listen for footswitch events:
	fswDev := FindDeviceByName("PCsensor FootSwitch3")
	if fswDev != nil {
		fsw = ListenDevice(fswDev)
	}

	// Listen for touch events:
	touchDev := FindAbsDevice()
	if touchDev != nil {
		touch = ListenDevice(touchDev)
	}

	// Set up BCM display directly with an EGL context:
	display, err := bcm.OpenDisplay()
	if err != nil {
		panic(err)
	}
	defer display.Close()

	// Initialize NVG:
	vg := nvg.CreateGLES2(nvg.Antialias | nvg.Debug)
	defer nvg.DeleteGLES2(vg)

	winWidth := int32(display.Width())
	winHeight := int32(display.Height())

	// Set up GL viewport:
	gl.Viewport(0, 0, int32(display.Width()), int32(display.Height()))

	gl.ClearColor(0.0, 0.0, 0.2, 1.0)

mainloop:
	for {
		// Clear background:
		gl.Clear(gl.COLOR_BUFFER_BIT)

		nvg.BeginFrame(vg, winWidth, winHeight, 1.0)

		// Window
		nvg.BeginPath(vg)
		nvg.RoundedRect(vg, 10, 10, 200, 300, 3.0)
		nvg.FillColor(vg, nvg.RGBA(28, 30, 34, 192))
		// nvg.FillColor(vg, nvg.RGBA(0,0,0,128));
		nvg.Fill(vg)

		nvg.EndFrame(vg)

		// Swap current surface to display:
		err = display.SwapBuffers()
		if err != nil {
			panic(err)
		}

		// Await a footswitch event:
		select {
		case ev := <-touch:
			fmt.Println(ev)
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
