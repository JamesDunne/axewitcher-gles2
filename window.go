package main

type Point struct {
	X float32
	Y float32
}

type Window struct {
	X float32
	Y float32
	W float32
	H float32
}

func NewWindow(x, y, w, h float32) Window {
	return Window{x, y, w, h}
}

func (n Window) Inner(l, t, r, b float32) Window {
	return Window{n.X + l, n.Y + t, n.W - r - l, n.H - t - b}
}

func (n Window) SplitH(t float32) (top Window, bottom Window) {
	top = Window{n.X, n.Y, n.W, t - 1}
	bottom = Window{n.X, n.Y + t, n.W, n.H - t}
	return
}

func (n Window) SplitV(l float32) (left Window, right Window) {
	left = Window{n.X, n.Y, l - 1, n.H}
	right = Window{l, n.Y, n.W - l, n.H}
	return
}
