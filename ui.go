/*
   It should be a kinda "immediate" Ui, see for example
   https://github.com/ocornut/imgui

   Here is the button, layout and all Ui related stuff implemented.

   We have layout (Vertical for now) that can hold buttons.
   The layout can be placed anywhere.

   the button function returns true once when it is clicked.
   @Todo should we change this to return true when the button is down and let
         the user handle the clicked logic?
*/

package tomato

import (
	//"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/math/fixed"
)

type Size = image.Point

const (
	TEXT_SIZE     float64 = 24
	BUTTON_HEIGHT float64 = 56
	Y_MARGIN      int     = 4
	MAX_BUTTONS   int     = 32
)

type button struct {
	Size     Size
	DrwUp    draw.Image
	DrwDown  draw.Image
	DrwHover draw.Image
}

type layout struct {
	Ori     Orientation
	Place   image.Rectangle
	Elems   [MAX_BUTTONS]button
	NextPos image.Point
}

type Ui_Frame struct {
	Layouts      []layout
	Active       int // maps to active layout
	DefaultTheme ButtonColorTheme
}

type Orientation uint8

const (
	Vertical Orientation = iota
	Horizontal
)

type ButtonColorTheme struct {
	Text     color.RGBA
	BgUp     color.RGBA
	BgHover  color.RGBA
	FontFace font.Face
	//Blink color.RGBA
}

var ui_frame Ui_Frame

func SetupUi() {

	ui_frame = Ui_Frame{}
	font, err := truetype.Parse(gomono.TTF)
	if err != nil {
		panic(err)
	}

	fontFace := truetype.NewFace(font, &truetype.Options{
		Size: TEXT_SIZE,
	})

	ui_frame.DefaultTheme = ButtonColorTheme{
		Text:     color.RGBA{255, 250, 240, 255}, // Floral White
		BgUp:     color.RGBA{36, 33, 36, 255},    // Raisin Black
		BgHover:  color.RGBA{45, 45, 45, 255},
		FontFace: fontFace,
	}
	ui_frame.Layouts = make([]layout, 0)
}

func Layout(id int, orientation Orientation, place image.Rectangle) {
	if id >= len(ui_frame.Layouts) {
		if id != len(ui_frame.Layouts) {
			panic("Need To call Layout(...) with ids starting from 0 in increaing manner.")
		}
		ui_frame.Layouts = append(ui_frame.Layouts, layout{
			Ori:     orientation,
			Place:   place,
			NextPos: place.Min,
		})
	}
	ui_frame.Active = id
}

// in current layout! delete the buttons for now
func InvalidateElements() {
	lay := &ui_frame.Layouts[ui_frame.Active]
	for id := range MAX_BUTTONS {
		lay.Elems[id] = button{}
	}
}

var previousDown bool // keeps track of the prevois MouseDownL state to detect clicks

// returns true if it has been clicked!
func TextButton(id int, text string, theme *ButtonColorTheme) bool { // use nil for default theme
	if len(ui_frame.Layouts) == 0 {
		panic("\ntomato ERROR: call ui.Layout(0, ui.Vertical, image.Rect(0,0,100,100)) at least before button!\n")
	}

	// @Todo make id independent of MAX_BUTTONS
	if id >= MAX_BUTTONS || id < 0 {
		panic("id must be from 0 to MAX_BUTTONS-1")
	}

	lay := &ui_frame.Layouts[ui_frame.Active]

	// create if it doesn't exist yet
	if lay.Elems[id].DrwUp == nil {
		size := Size{lay.Place.Dx(), int(math.Ceil(BUTTON_HEIGHT))}
		rect := image.Rectangle{image.Pt(0, 0), size}

		if theme == nil {
			theme = &ui_frame.DefaultTheme
		}
		u, h := button_render(text, theme, rect)

		ui_frame.Layouts[ui_frame.Active].Elems[id] = button{
			Size:     size,
			DrwUp:    u,
			DrwDown:  h,
			DrwHover: h,
		}
	}

	b := &ui_frame.Layouts[ui_frame.Active].Elems[id]
	zp := lay.NextPos
	mp := zp.Add(b.Size)
	target := image.Rectangle{zp, mp}
	mouse := image.Pt(MouseX, MouseY)

	if lay.Ori == Vertical {
		if mouse.In(target) {
			ToDraw(target, b.DrwHover)
		} else {
			ToDraw(target, b.DrwUp)
		}
		lay.NextPos = zp.Add(image.Pt(0, b.Size.Y+Y_MARGIN))

	} else {
		panic("not implemented!")
	}

	// We know we clicked and are inside this button
	if MouseDownL && previousDown != MouseDownL && mouse.In(target) {
		return true
	}

	return false
}

func DrawUi() {
	// reset all layout next positions to their origin
	for i := range ui_frame.Layouts {
		ui_frame.Layouts[i].NextPos = ui_frame.Layouts[i].Place.Min
	}
	previousDown = MouseDownL

	// @Todo should it call it?
	Draw()
}

func RenderText(text string, textColor, btnColor color.RGBA, fontFace font.Face) draw.Image {

	drawer := &font.Drawer{
		Src:  &image.Uniform{textColor},
		Face: fontFace,
		Dot:  fixed.P(0, 0),
	}

	b26_6, _ := drawer.BoundString(text)
	bounds := image.Rect(
		b26_6.Min.X.Floor(),
		b26_6.Min.Y.Floor(),
		b26_6.Max.X.Ceil(),
		b26_6.Max.Y.Ceil(),
	)

	drawer.Dst = image.NewRGBA(bounds)
	btnUpUniform := image.NewUniform(btnColor)
	draw.Draw(drawer.Dst, bounds, btnUpUniform, image.ZP, draw.Src)
	drawer.DrawString(text)
	return drawer.Dst
}

func RenderTextMulti(text string, textColor, bgColor color.RGBA, fontFace font.Face, maxWidth int) draw.Image {

	drawer := &font.Drawer{
		Src:  &image.Uniform{textColor},
		Face: fontFace,
		Dot:  fixed.P(0, 0),
	}

	lines := make([]string, 0)

	j := 0
	i := 0
	for i = 0; i < len(text)-1; i++ {
		b26_6, _ := drawer.BoundString(text[j : i+1])
		if b26_6.Max.X.Ceil()-b26_6.Min.X.Floor() > maxWidth {
			lines = append(lines, text[j:i])
			j = i
		}
	}

	if i != j {
		lines = append(lines, text[j:i])
	}

	maxW := 0
	lineH := 0
	for _, line := range lines {
		b26_6, _ := drawer.BoundString(line)
		bounds := image.Rect(
			b26_6.Min.X.Floor(),
			b26_6.Min.Y.Floor(),
			b26_6.Max.X.Ceil(),
			b26_6.Max.Y.Ceil(),
		)
		if bounds.Dx() > maxW {
			maxW = bounds.Dx()
		}
		if bounds.Dy() > lineH {
			lineH = bounds.Dy()
		}
	}

	bounds := image.Rect(0, 0, maxW, (len(lines)+1)*lineH)
	result := image.NewRGBA(bounds)
	bgUniform := image.NewUniform(bgColor)
	draw.Draw(result, bounds, bgUniform, image.ZP, draw.Src)

	for i, line := range lines {
		tImage := RenderText(line, textColor, bgColor, fontFace)
		draw.Draw(result, bounds, tImage, bounds.Min.Sub(image.Pt(0, lineH*(i+1))), draw.Src)
	}
	return result
}

func button_render(text string, colorTheme *ButtonColorTheme, r image.Rectangle) (draw.Image, draw.Image) {
	var textImageUp image.Image
	var textImageHover image.Image
	{
		textImageUp = RenderText(text, colorTheme.Text, colorTheme.BgUp, colorTheme.FontFace)
		textImageHover = RenderText(text, colorTheme.Text, colorTheme.BgHover, colorTheme.FontFace)
	}

	redraw := func(hover bool) draw.Image {
		img := image.NewRGBA(r)
		var textImage image.Image
		var buttonBg image.Image
		if hover {
			buttonBg = image.NewUniform(colorTheme.BgHover)
			textImage = textImageHover
		} else {
			buttonBg = image.NewUniform(colorTheme.BgUp)
			textImage = textImageUp
		}
		draw.Draw(img, r, buttonBg, image.ZP, draw.Src)
		textRect := r
		textRect.Min.Y += textRect.Dy()/2 - textImage.Bounds().Dy()/2
		textRect.Min.X += textRect.Dx()/2 - textImage.Bounds().Dx()/2

		draw.Draw(img, textRect, textImage, textImage.Bounds().Min, draw.Src)
		return img
	}

	normalImg := redraw(false)
	hoveredImg := redraw(true)
	return normalImg, hoveredImg
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
