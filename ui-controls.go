package main

import (
	"github.com/JamesDunne/golang-nanovg/nvg"
)

const pad = 2
const size = 28.0
const round = 4.0

func (u *UI) isTouched(w Window) bool {
	for _, t := range u.Touches {
		// Skip released touch points:
		if t.ID <= 0 {
			continue
		}

		p := Point{t.X, t.Y}
		if w.IsPointInside(p) {
			return true
		}
	}
	return false
}

func (ui *UI) Button(w Window) bool {
	touched := ui.isTouched(w)
	if touched {
		ui.StrokeColor(ui.Palette(2))
		ui.FillColor(ui.Palette(1))
	} else {
		ui.StrokeColor(ui.Palette(1))
		ui.FillColor(ui.Palette(2))
	}

	ui.BeginPath()
	ui.RoundedRect(w, round)
	ui.Stroke()
	ui.Fill()

	// TODO: use state for latching behavior
	return touched
}

func (ui *UI) Pane(w Window) {
	ui.BeginPath()
	ui.RoundedRect(w, round)
	ui.Stroke()
}

func (ui *UI) Label(w Window, string string) {
	ui.BeginPath()
	ui.RoundedRect(w, round)
	ui.FillColor(ui.Palette(1))
	ui.Fill()

	lblText := w.Inner(pad*2, 0, pad*2, 0)
	ui.FillColor(ui.Palette(5))
	ui.Text(lblText, size, nvg.AlignLeft|nvg.AlignTop, string)
}

func (ui *UI) Dial(w Window, label string, value float32, valueStr string) {
	c := w.AlignedPoint(nvg.AlignCenter | nvg.AlignMiddle)
	r := w.RadiusMin()

	ui.Save()

	// Filled center:
	ui.BeginPath()
	ui.Circle(c, r-4.0)
	ui.FillColor(ui.Palette(1))
	ui.Fill()

	// Highlighted arc:
	ui.BeginPath()
	ui.Arc(c, r-2.0, nvg.Pi*0.75, nvg.Pi*2.25, nvg.Cw)
	ui.StrokeWidth(4.0)
	ui.MiterLimit(1.0)
	ui.LineCap(nvg.Square)
	ui.StrokeColor(ui.Palette(3))
	ui.Stroke()

	ui.Restore()
}
