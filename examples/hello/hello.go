package main

import (
	"time"
	"fmt"
	"image"

	"github.com/bbeni/tomato"
	//"github.com/go-gl/gl/v4.2-core/gl"
)

func main() {
	err := tomato.Create(1080, 720, "Hello Tomato");

	if err != nil {
		panic(err)
	}

	var step int;
    for tomato.Alive() {

    	// gather input
    	loop:
    	for {
			select {
			case event := <-tomato.Events():
				switch event := event.(type) {
				case tomato.KeyDown,
					 tomato.KeyType:
					tomato.Die()
					fmt.Println(event)
				default:
					fmt.Println(event)
				}
			default:
				break loop
			}
		}

    	// do logic

        //fmt.Println(step)
        step++

    	// draw stuff

    	// draw Gui

        r := image.Rectangle{image.Pt(0, 0),image.Pt(100, 100)}
        tomato.Draw(r)
        tomato.Win.SwapBuffers()

        time.Sleep(time.Second / 120)
    }
}

