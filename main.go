package main

import (
	"runtime"

	axe "github.com/JamesDunne/axewitcher"
	"github.com/JamesDunne/rpi-egl/bcm"
	gl "github.com/JamesDunne/rpi-egl/gles2"
)

func main() {
	// Lock main goroutine to OS thread for EGL safety:
	runtime.LockOSThread()

	// Listen for footswitch events:
	fsw, err := ListenFootswitch()
	if err != nil {
		panic(err)
	}

	// Set up BCM display directly with an EGL context:
	display, err := bcm.OpenDisplay()
	if err != nil {
		panic(err)
	}
	defer display.Close()

	// Set up GL viewport:
	gl.Viewport(0, 0, int32(display.Width()), int32(display.Height()))

	gl.ClearColor(0.0, 0.0, 1.0, 1.0)

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
			switch ev.Fsw {
			case axe.FswReset:
				break mainloop
			}
		}
	}
}
