package main

import (
	"log"
	"time"

	"github.com/JamesDunne/rpi-egl/bcm"
	gl "github.com/JamesDunne/rpi-egl/gles2"
)

func main() {
	display := bcm.OpenDisplay()
	if display == nil {
		return
	}
	defer display.Close()

	err := gl.Init()
	if err != nil {
		panic(err)
	}

	log.Println(display.Width(), display.Height())
	gl.Viewport(0, 0, int32(display.Width()), int32(display.Height()))

	gl.ClearColor(0, 0, 1.0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	for i := 0; i < 10; i++ {
		gl.Clear(gl.COLOR_BUFFER_BIT)
		display.SwapBuffers()
	}

	time.Sleep(5 * time.Second)
}
