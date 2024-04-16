package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/bbeni/tomato"
	// "github.com/go-gl/gl/v4.2-core/gl" // Use this version of OpenGL
)

func main() {
	err := tomato.Create(1080, 720, "Hello Tomato");
	if err != nil {
		panic(err)
	}

	// Tomato!
    where1 := image.Rectangle{image.Pt(20, 300), image.Pt(420, 700)}
    img1   := image.NewUniform(color.RGBA{200, 30, 30, 255})
    where2 := image.Rectangle{image.Pt(120, 400), image.Pt(320, 600)}
    img2   := image.NewUniform(color.RGBA{255, 0, 0, 255})
    where3 := image.Rectangle{image.Pt(220, 250), image.Pt(260, 310)}
    img3   := image.NewUniform(color.RGBA{23, 200, 23, 255})

    for tomato.Alive() {
    	// gather input
    	loop:
    	for {
			select {
			case event := <-tomato.Events():
				switch event.Kind {
				case tomato.KeyDown:
					switch event.Key {
					case tomato.Escape:
						tomato.Die()
					default:
						fmt.Println(event.Key)
						break
					}
				case tomato.RuneTyped:
					fmt.Println(string(event.Rune))
				case tomato.MouDown:
					fmt.Println(event)
				}
				// move tomato?!
				where1 = where1.Add(image.Pt(1,0))
				where2 = where2.Add(image.Pt(1,0))
				where3 = where3.Add(image.Pt(1,0))
			default:
				break loop
			}
		}

    	// do logic
		// ...

		// draw your Game
		// ...

    	// draw Gui Tomato!
		tomato.ToDraw(where1, img1)
		tomato.ToDraw(where2, img2)
		tomato.ToDraw(where3, img3)

        tomato.Draw()
        tomato.Win.SwapBuffers()
    }
}

