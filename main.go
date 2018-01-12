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

type Touch struct {
	Point
	ID int32
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
	//display, err := bcm.OpenDisplay(5, 6, 5)
	display, err := bcm.OpenDisplay(8, 8, 8)
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

	w := NewWindow(0, 0, float32(winWidth), float32(winHeight))

	ui := NewUI(vg)
	touchSlot := 0

	const pad = 2
	const size = 28.0
	const round = 4.0

mainloop:
	for {
		// Clear background:
		gl.Clear(gl.COLOR_BUFFER_BIT)

		nvg.BeginFrame(vg, winWidth, winHeight, 1.0)

		ui.FillColor(ui.Palette(0))
		ui.BeginPath()
		ui.Rect(w)
		ui.Fill()

		top, bottom := w.SplitH(size + 8)

		song := top.Inner(pad, pad, pad, pad)
		ui.Label(song, "Trippin on a Hole in a Paper Heart")

		// Split screen for MG v JD:
		mg, jd := bottom.SplitV(bottom.W * 0.5)
		mg = mg.Inner(0, pad, 0, pad)
		jd = jd.Inner(0, pad, 0, pad)

		drawAmp := func(w Window, name string) {
			ui.StrokeWidth(1.0)
			ui.StrokeColor(ui.Palette(1))

			ui.Pane(w)

			// Amp label at top center:
			label, w := w.SplitH(size + 8)
			ui.FillColor(ui.Palette(5))
			ui.Text(label, size, nvg.AlignCenter|nvg.AlignTop, name)

			// Tri-state buttons:
			top, bottom := w.SplitH(size + 16)
			btnHeight := top.W * 0.33333333
			btnDirty, top := top.SplitV(btnHeight)
			btnClean, btnAcoustic := top.SplitV(btnHeight)

			ui.Button(btnDirty)
			ui.Button(btnClean)
			ui.Button(btnAcoustic)

			ui.FillColor(ui.Palette(0))
			ui.Text(btnDirty, size, nvg.AlignCenter|nvg.AlignMiddle, "dirty")
			ui.Text(btnClean, size, nvg.AlignCenter|nvg.AlignMiddle, "clean")
			ui.Text(btnAcoustic, size, nvg.AlignCenter|nvg.AlignMiddle, "acoustic")

			// FX toggles:
			fxWidth := bottom.W / 5.0
			mid, bottom := bottom.SplitH(bottom.H - (size + 16))
			fxNames := [...]string{"pit1", "rtr1", "phr1", "cho1", "dly1"}
			for i := 0; i < 5; i++ {
				var btnFX Window
				btnFX, bottom = bottom.SplitV(fxWidth)

				ui.Button(btnFX)

				ui.FillColor(ui.Palette(0))
				ui.Text(btnFX, size, nvg.AlignCenter|nvg.AlignMiddle, fxNames[i])
			}

			ui.StrokeColor(ui.Palette(1))
			ui.Pane(mid)
		}
		drawAmp(mg, "MG")
		drawAmp(jd, "JD")

		// Draw touch points:
		for _, tp := range ui.Touches {
			if tp.ID <= 0 {
				continue
			}

			ui.FillColor(nvg.RGBA(255, 255, 255, 160))
			ui.BeginPath()
			ui.Circle(tp.Point, 15.0)
			ui.Fill()
		}

		nvg.EndFrame(vg)

		// Swap current surface to display:
		display.SwapBuffers()

		// Await an event:
		select {
		case evs := <-touch:
			// Process touch events with absolute coordinates:
			//fmt.Println("[")
			for _, ev := range evs {
				if ev.Type != evdev.EV_ABS {
					continue
				}

				//fmt.Println(evdev.ABS[int(ev.Code)], ev.Value)

				switch ev.Code {
				case evdev.ABS_MT_SLOT:
					touchSlot = int(ev.Value)
				case evdev.ABS_MT_POSITION_X:
					ui.Touches[touchSlot].X = float32(ev.Value)
				case evdev.ABS_MT_POSITION_Y:
					ui.Touches[touchSlot].Y = float32(ev.Value)
				case evdev.ABS_MT_TRACKING_ID:
					ui.Touches[touchSlot].ID = ev.Value
				}
			}
			//fmt.Println("]")
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
