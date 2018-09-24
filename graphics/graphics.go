package graphics

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

const screenWidth = 1024
const screenHeight = 512

var (
	window   sdl.Window
	renderer *sdl.Renderer
)

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

// Init initializes sdl
func Init() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	panicOnError(err)

	window, err := sdl.CreateWindow(
		"CHIP-8",
		1000,
		20,
		screenWidth,
		screenHeight,
		sdl.WINDOW_SHOWN,
	)
	panicOnError(err)

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	panicOnError(err)
}

// RenderBuffer clears the screen and renders a buffer
func RenderBuffer(buffer []bool, bufferWidth int, pixelRatio int) {
	// Clear screen
	renderer.Clear()

	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.FillRect(nil)

	renderer.SetDrawColor(255, 255, 255, 255)
	// Loop over pixel values in buffer and render
	for i := range buffer {
		if buffer[i] {
			var x = (i % bufferWidth) * pixelRatio
			var y = ((i - i%bufferWidth) / bufferWidth) * pixelRatio
			rect := sdl.Rect{
				X: int32(x),
				Y: int32(y),
				W: int32(pixelRatio),
				H: int32(pixelRatio),
			}
			renderer.FillRect(&rect)
		}
	}
	renderer.Present()
}

// GetKeyboardState returns an array of booleans for 16 keys
func GetKeyboardState() [16]bool {
	sdl.PumpEvents()
	var state = sdl.GetKeyboardState()

	fmt.Printf("1 %d\t 2 %d\t 3 %d 4 %d\n Q %d\t W %d\t E %d R %d\n A %d\t S %d\t D %d F %d\n Z %d\t X %d\t C %d V %d\n",
		state[sdl.SCANCODE_1],
		state[sdl.SCANCODE_2],
		state[sdl.SCANCODE_3],
		state[sdl.SCANCODE_4],
		state[sdl.SCANCODE_Q],
		state[sdl.SCANCODE_W],
		state[sdl.SCANCODE_E],
		state[sdl.SCANCODE_R],
		state[sdl.SCANCODE_A],
		state[sdl.SCANCODE_S],
		state[sdl.SCANCODE_D],
		state[sdl.SCANCODE_F],
		state[sdl.SCANCODE_Z],
		state[sdl.SCANCODE_X],
		state[sdl.SCANCODE_C],
		state[sdl.SCANCODE_V])
	return [16]bool{
		state[sdl.SCANCODE_1] == 1,
		state[sdl.SCANCODE_2] == 1,
		state[sdl.SCANCODE_3] == 1,
		state[sdl.SCANCODE_4] == 1,
		state[sdl.SCANCODE_Q] == 1,
		state[sdl.SCANCODE_W] == 1,
		state[sdl.SCANCODE_E] == 1,
		state[sdl.SCANCODE_R] == 1,
		state[sdl.SCANCODE_A] == 1,
		state[sdl.SCANCODE_S] == 1,
		state[sdl.SCANCODE_D] == 1,
		state[sdl.SCANCODE_F] == 1,
		state[sdl.SCANCODE_Z] == 1,
		state[sdl.SCANCODE_X] == 1,
		state[sdl.SCANCODE_C] == 1,
		state[sdl.SCANCODE_V] == 1,
	}
}

// Teardown destroys the sdl session
func Teardown() {
	sdl.Quit()
	window.Destroy()
	renderer.Destroy()
}
