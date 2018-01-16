package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	axe "github.com/JamesDunne/axewitcher"
	"github.com/gvalkov/golang-evdev"

	"github.com/JamesDunne/golang-nanovg/nvg"
	"github.com/JamesDunne/golang-nanovg/nvgui"
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

	// Initialize UI:
	ui := nvgui.NewUI()
	err := ui.InitDisplay()
	if err != nil {
		panic(err)
	}

	err = ui.CreateFont("sans", "sans.ttf")
	if err != nil {
		panic(err)
	}
	ui.FontFace("sans")

	// Create MIDI interface:
	midi, err := axe.NewMidi()
	if err != nil {
		log.Println(err)
		// Use null driver:
		midi, err = axe.NewNullMidi()
	}
	defer midi.Close()

	// Initialize controller:
	controller := axe.NewController(midi)
	err = controller.Load()
	if err != nil {
		log.Fatal("Unable to load programs: ", err)
	}
	controller.Init()

	// Create window to represent display:
	w := ui.Window()

	touchSlot := 0

	const size = 28

	amps := [...]string{"MG", "JD"}

mainloop:
	for {
		ui.BeginFrame()

		ui.FillColor(ui.Palette(0))
		ui.BeginPath()
		ui.Rect(w)
		ui.Fill()

		top, bottom := w.SplitH(size + 8)

		ui.Label(top, "Trippin on a Hole in a Paper Heart", nvg.AlignLeft|nvg.AlignTop)

		// Split screen for MG v JD:
		mg, jd := bottom.SplitV(bottom.W * 0.5)

		drawAmp := func(w nvgui.Window, ampNo int) {
			amp := &controller.Curr.Amp[ampNo]

			ui.StrokeWidth(1.0)
			ui.StrokeColor(ui.Palette(3))
			ui.Pane(w)

			// Amp label at top center:
			label, w := w.SplitH(size + 8)
			ui.FillColor(ui.Palette(4))
			ui.Text(label, size, nvg.AlignCenter|nvg.AlignTop, amps[ampNo])

			// Tri-state buttons:
			top, bottom := w.SplitH(size + 16)
			btnHeight := top.W * 0.33333333
			btnDirty, top := top.SplitV(btnHeight)
			btnClean, btnAcoustic := top.SplitV(btnHeight)

			if t := ui.Button(btnDirty, amp.Mode == axe.AmpDirty, "dirty"); t != nil {

			}
			ui.Button(btnClean, amp.Mode == axe.AmpClean, "clean")
			ui.Button(btnAcoustic, amp.Mode == axe.AmpAcoustic, "acoustic")

			// FX toggles:
			fxWidth := bottom.W / 5.0
			mid, bottom := bottom.SplitH(bottom.H - (size + 16))
			fxNames := [...]string{"pit1", "rtr1", "phr1", "cho1", "dly1"}
			for i := 0; i < 5; i++ {
				var btnFX nvgui.Window
				btnFX, bottom = bottom.SplitV(fxWidth)
				ui.Button(btnFX, amp.Fx[i].Enabled, fxNames[i])
			}

			ui.StrokeColor(ui.Palette(3))
			ui.Pane(mid)

			gain, volume := mid.SplitV(mid.W * 0.5)
			g := float32(amp.DirtyGain) / 127.0
			ui.Dial(gain, "Gain", g, "0.68")
			v := float32(amp.Volume) / 127.0
			ui.Dial(volume, "Volume", v, "0 dB")
		}
		drawAmp(mg, 0)
		drawAmp(jd, 1)

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

		ui.EndFrame()

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
