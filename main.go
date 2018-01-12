package main

import (
	"errors"
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

func main() {
	// Lock main goroutine to OS thread for EGL safety:
	runtime.LockOSThread()

	var (
		fsw   <-chan []evdev.InputEvent
		touch <-chan []evdev.InputEvent
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
	display.SwapInterval(0)

	// Initialize NVG:
	vg := nvg.CreateGLES2(nvg.Antialias | nvg.Debug)
	defer nvg.DeleteGLES2(vg)

	fontSans := nvg.CreateFont(vg, "sans", "sans.ttf")
	if fontSans == -1 {
		panic(errors.New("could not load sans.ttf"))
	}
	nvg.FontFace(vg, "sans")

	winWidth := int32(display.Width())
	winHeight := int32(display.Height())

	// Set up GL viewport:
	gl.Viewport(0, 0, winWidth, winHeight)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)

	x, y := float32(0), float32(0)
	w := NewWindow(0, 0, float32(winWidth), float32(winHeight))

	ui := NewUI(vg)

mainloop:
	for {
		// Clear background:
		gl.Clear(gl.COLOR_BUFFER_BIT)

		nvg.BeginFrame(vg, winWidth, winHeight, 1.0)

		const pad = 2
		const size = 28.0

		ui.FillColor(ui.Palette(0))
		ui.Rect(w)
		ui.Fill()

		top, bottom := w.SplitH(34)

		song := top.Inner(pad, pad, pad, pad)
		ui.RoundedRect(song, 3.0)
		ui.FillColor(ui.Palette(2))
		ui.Fill()

		songText := song.Inner(pad*2, 0, pad*2, 0)
		ui.FillColor(nvg.RGB(0, 0, 0))
		ui.Text(songText, size, nvg.AlignLeft|nvg.AlignTop, "Trippin on a Hole in a Paper Heart")

		// Split screen for MG v JD:
		mg, jd := bottom.SplitH(bottom.H * 0.5)
		ui.StrokeWidth(2.0)
		ui.StrokeColor(ui.Palette(1))
		ui.RoundedRect(mg, 3.0)
		ui.Stroke()
		ui.RoundedRect(jd, 3.0)
		ui.Stroke()

		//for i := 0; i < 4; i++ {
		//	nvg.BeginPath(vg)
		//	nvg.RoundedRect(vg, 200*float32(i), 0, 200, 240, 3.0)
		//	nvg.FillColor(vg, palette[i])
		//	nvg.Fill(vg)
		//}

		nvg.FillColor(vg, nvg.RGBA(255, 255, 255, 160))
		nvg.TextAlign(vg, nvg.AlignCenter|nvg.AlignMiddle)
		nvg.Text(vg, x, y, "Hello, world!")

		nvg.EndFrame(vg)

		// Swap current surface to display:
		display.SwapBuffers()

		// Await an event:
		select {
		case evs := <-touch:
			// Process touch events with absolute coordinates:
			for _, ev := range evs {
				if ev.Type != evdev.EV_ABS {
					continue
				}

				switch ev.Code {
				case evdev.ABS_X:
					x = float32(ev.Value)
				case evdev.ABS_Y:
					y = float32(ev.Value)
				}
			}
		case evs := <-fsw:
			// Process footswitch (keyboard) events:
			for i := range evs {
				ev := &evs[i]
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
}
