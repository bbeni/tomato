package main

import (
	"fmt"
	"image"
	"os"

	//"image/color"
	"time"

	"github.com/bbeni/tomato"
	// "github.com/go-gl/gl/v4.2-core/gl" // Use this version of OpenGL
)

var open [4]bool

func main() {
	err := tomato.Create(1080, 720, "Hello Tomato/ui")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tomato.SetupUi()

	last := time.Now()

	for tomato.Alive() {

		// read input events and set variables
	loop:
		for {
			select {
			case event := <-tomato.Events():

				switch event.Kind {
				case tomato.KeyDown:
					if event.Key == tomato.Escape {
						tomato.Die()
					}
				default:
					//fmt.Println(event)
				}
			default:
				break loop
			}
		}

		// render frame

		// @Bug not working right now ... doesn't show stuff if used
		//tomato.Draw()

		// render ui
		for i := range 4 {

			w := 250
			tomato.Layout(i, tomato.Vertical, image.Rect(i*w, 0, (i+1)*w-2, 400))

			if tomato.TextButton(0, "Open/Close Me", nil) {
				open[i] = !open[i]
			}

			if open[i] {
				if tomato.TextButton(1, "Click toggles ->", nil) {
					open[(i+1)%4] = !open[(i+1)%4]
				}
				tomato.TextButton(2, "How", nil)
				tomato.TextButton(3, "is the", nil)
				tomato.TextButton(4, "Weather?", nil)
			}
		}

		time.Sleep(time.Millisecond * 30)
		tomato.DrawUi()
		tomato.Win.SwapBuffers()
		tomato.Clear() // prepare the other buffer for rendering again

		now := time.Now()
		dt := now.Sub(last)
		fmt.Println(1.0/dt.Seconds(), "FPS")
		last = time.Now()
	}

}
