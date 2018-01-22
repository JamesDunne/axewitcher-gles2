package main

import (
	"log"
	"runtime"
	"time"

	axe "github.com/JamesDunne/axewitcher"

	"github.com/JamesDunne/golang-nanovg/nvg"
	"github.com/JamesDunne/golang-nanovg/nvgui"

	// for profiling:
	"os"
	"runtime/pprof"
)

func main() {
	// Lock main goroutine to OS thread for EGL safety:
	runtime.LockOSThread()

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

	// Initialize event handlers for fsw, touchscreen:
	eventListener := InitEventListener()

	// Create window to represent display:
	w := ui.Window()

	const size = 28

	amps := [...]string{"MG", "JD"}

	{
		f, err := os.Create("profile.log")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

mainloop:
	for f := 0; f < 60; f++ {
		start := time.Now()

		ui.BeginFrame()

		//ui.FillColor(ui.Palette(0))
		//ui.BeginPath()
		//ui.Rect(w)
		//ui.Fill()

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

		elapsed := time.Since(start)
		log.Printf("%s\n", elapsed)

		// Await an event:
		//eventListener.Await()
		copy(ui.Touches, eventListener.Touches)

		// Process fsw events:
		for _, ev := range eventListener.FswEvents {
			if ev.State && ev.Fsw == axe.FswReset {
				break mainloop
			}
		}
	}
}
