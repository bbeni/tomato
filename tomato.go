// graphics stuff
package tomato

import (
	"fmt"
	"image"
	"strings"
	"image/draw"
	"image/color"
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/gl/v4.2-core/gl"
)

var Win 	  *glfw.Window
var GuiImg 	   image.Image

// gl stuff
var GuiShader  uint32
var GuiTexture uint32
var GuiQuadVAO uint32
var GuiQuad = []float32 {
	//  X, Y, Z, U, V
	-1.0,  1.0, 1.0,  0.0, 0.0,
	1.0,  -1.0, 1.0,  1.0, 1.0,
	-1.0, -1.0, 1.0,  0.0, 1.0,
	-1.0,  1.0, 1.0,  0.0, 0.0,
	1.0,   1.0, 1.0,  1.0, 0.0,
	1.0,  -1.0, 1.0,  1.0, 1.0,
}

func Alive() bool {
	if !Win.ShouldClose() && !dead {
        glfw.PollEvents()
        return true
	} else {
		Win.Destroy()
	    glfw.Terminate()
		return false
	}
}

var dead bool
func Die() {
	dead = true
}

type Event interface{}

var inEvents  chan Event
var outEvents chan Event

func Events() <-chan Event {
	return outEvents
}


func Create(width, height int, title string) (error) {
	err := glfw.Init()
	if err != nil {
		return err
	}
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	//glfw.WindowHint(glfw.Decorated, glfw.False)
	w, err := glfw.CreateWindow(width, height, title, nil, nil)

	if err != nil {
		return err
	}

	Win = w
    Win.MakeContextCurrent()

	if err = gl.Init(); err != nil {
		return err
	}

	openGLSetup()
	eventsSetup()
	return nil
}

func eventsSetup() {
	var mouseX, mouseY int

	inEvents  = make(chan Event)
	outEvents = make(chan Event)

	go func() {

		var queue []Event;

		for {
			in, success := <-inEvents
			if !success {
				close(outEvents)
			}
			queue = append(queue, in)

			for len(queue) > 0 {
				select {
				case outEvents <- queue[0]:
					queue = queue[1:]
				case in, success := <-inEvents:
					if !success {
						for _, in := range queue {
							outEvents <- in
						}
						close(outEvents)
						return
					}
					queue = append(queue, in)
				}
			}
		}
	}()


	Win.SetCursorPosCallback(func(_ *glfw.Window, x, y float64) {
		mouseX, mouseY = int(x), int(y)
		inEvents <- MouMove{image.Pt(mouseX, mouseY)}
	})

	Win.SetMouseButtonCallback(func(_ *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
		b, ok := buttons[button]
		if !ok {
			return
		}
		switch action {
		case glfw.Press:
			inEvents <- MouDown{image.Pt(mouseX, mouseY), b}
		case glfw.Release:
			inEvents <- MouUp{image.Pt(mouseX, mouseY), b}
		}
	})

	Win.SetScrollCallback(func(_ *glfw.Window, xoff, yoff float64) {
		inEvents <- MouScroll{image.Pt(int(xoff), int(yoff))}
	})

	Win.SetCharCallback(func(_ *glfw.Window, r rune) {
		fmt.Println(r)
		inEvents <- KeyType{r}
	})

	Win.SetKeyCallback(func(_ *glfw.Window, key glfw.Key, _ int, action glfw.Action, _ glfw.ModifierKey) {
		k, ok := keys[key]
		if !ok {
			return
		}
		switch action {
		case glfw.Press:
			inEvents <- KeyDown{k}
		case glfw.Release:
			inEvents <- KeyUp{k}
		case glfw.Repeat:
			inEvents <- KeyRepeat{k}
		}
	})

	Win.SetFramebufferSizeCallback(func(_ *glfw.Window, width, height int) {
		//r := image.Rect(0, 0, width, height)
		//w.newSize <- r
		//inEvents <- gui.Resize{Rectangle: r}
	})

	Win.SetCloseCallback(func(_ *glfw.Window) {
		inEvents <- WinClose{}
	})
}


func openGLSetup() error {
	var err error
	var screenVertShader = `
		#version 420

		in vec3 vert;
		in vec2 vertTexCoord;
		out vec2 fragTexCoord;

		void main() {
			fragTexCoord = vertTexCoord;
			gl_Position = vec4(vert.xy, 0.0, 1.0);
		}
	` + "\x00"

	var screenFragShader = `
		#version 420

		uniform sampler2D tex;
		in vec2 fragTexCoord;

		out vec4 outputColor;

		void main() {
			outputColor = texture(tex, fragTexCoord);
		}
	` + "\x00"

	GuiShader, err = NewGLProgram(screenVertShader, screenFragShader)

	if err != nil {
		fmt.Print("ERROR making GuiShader: ")
		fmt.Println(err)
		return err
	}

	width, height := Win.GetFramebufferSize()
	GuiTexture = newScreenTexture(width, height)

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}
	GuiImg = image.NewRGBA(image.Rectangle{upLeft, lowRight})

	textureUniform := gl.GetUniformLocation(GuiShader, gl.Str("tex\x00"))
	gl.Uniform1i(textureUniform, 0)
	gl.BindFragDataLocation(GuiShader, 0, gl.Str("outputColor\x00"))

	gl.GenVertexArrays(1, &GuiQuadVAO)
	gl.BindVertexArray(GuiQuadVAO)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(GuiQuad)*4, gl.Ptr(GuiQuad), gl.STATIC_DRAW)

	vertAttrib := uint32(gl.GetAttribLocation(GuiShader, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointerWithOffset(vertAttrib, 3, gl.FLOAT, false, 5*4, 0)

	texCoordAttrib := uint32(gl.GetAttribLocation(GuiShader, gl.Str("vertTexCoord\x00")))
	gl.EnableVertexAttribArray(texCoordAttrib)
	gl.VertexAttribPointerWithOffset(texCoordAttrib, 2, gl.FLOAT, false, 5*4, 3*4)

	gl.ClearColor(1.0, 1.0, 0.0, 1.0)
	return nil
}

func Draw(r image.Rectangle) {

	bounds := GuiImg.Bounds()
	r = r.Intersect(bounds)
	if r.Empty() {
		return
	}

	tmp := image.NewRGBA(r)
	draw.Draw(tmp, r, GuiImg, r.Min, draw.Src)

	gl.UseProgram(GuiShader)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)  		 // Assume premultiplied alpha
	//gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA) // Non-premultipled version
	//gl.Clear(gl.DEPTH_BUFFER_BIT | gl.COLOR_BUFFER_BIT)

	gl.TextureSubImage2D(
		GuiTexture,
		0,
		int32(r.Min.X),
		int32(r.Min.Y),
		int32(r.Dx()),
		int32(r.Dy()),
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(tmp.Pix))

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	// TODO: might be wrong, need to add ceil/floor to the values.
	// TODO: scissor array of rects?
	_, hei := Win.GetFramebufferSize()
	gl.Enable(gl.SCISSOR_TEST)
	gl.Scissor(int32(r.Min.X), int32(hei) - int32(r.Max.Y), int32(r.Dx()), int32(r.Dy()))

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, GuiTexture)

	//TODO: this is a dirty trick to draw the gui on both buffers
	//      double render and we are on the same buffer as before.
	for range 2 {
		gl.Clear(gl.DEPTH_BUFFER_BIT)
		gl.BindVertexArray(GuiQuadVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 6*2*3)

		Win.SwapBuffers()
	}

	gl.Disable(gl.BLEND)
	gl.Disable(gl.SCISSOR_TEST)
	gl.Disable(gl.DEPTH_TEST)
}



func NewGLProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {

	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source)

	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func newScreenTexture(width, height int) (uint32) {

	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	if rgba.Stride != rgba.Rect.Size().X*4 {
		panic("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), image.NewUniform(color.RGBA{0,0,0,0}), image.Point{0, 0}, draw.Src)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))

	return texture
}



/* adapted from github.com/faiface/gui/win
	TODO: include licence
*/

var buttons = map[glfw.MouseButton]Button{
	glfw.MouseButtonLeft:   ButtonLeft,
	glfw.MouseButtonRight:  ButtonRight,
	glfw.MouseButtonMiddle: ButtonMiddle,
}

var keys = map[glfw.Key]Key{
	glfw.KeyLeft:         Left,
	glfw.KeyRight:        Right,
	glfw.KeyUp:           Up,
	glfw.KeyDown:         Down,
	glfw.KeyEscape:       Escape,
	glfw.KeySpace:        Space,
	glfw.KeyBackspace:    Backspace,
	glfw.KeyDelete:       Delete,
	glfw.KeyEnter:        Enter,
	glfw.KeyTab:          Tab,
	glfw.KeyHome:         Home,
	glfw.KeyEnd:          End,
	glfw.KeyPageUp:       PageUp,
	glfw.KeyPageDown:     PageDown,
	glfw.KeyLeftShift:    Shift,
	glfw.KeyRightShift:   Shift,
	glfw.KeyLeftControl:  Ctrl,
	glfw.KeyRightControl: Ctrl,
	glfw.KeyLeftAlt:      Alt,
	glfw.KeyRightAlt:     Alt,
}

type Button string
const (
	ButtonLeft   Button = "left"
	ButtonRight  Button = "right"
	ButtonMiddle Button = "middle"
)

//go:generate stringer -type=Key
type Key uint8
const (
	Left Key = iota
	Right
	Up
	Down
	Escape
	Space
	Backspace
	Delete
	Enter
	Tab
	Home
	End
	PageUp
	PageDown
	Shift
	Ctrl
	Alt
)

type (
	WinClose struct {  }
	MouMove struct {image.Point}
	MouDown struct {
		image.Point
		Button Button
	}
	MouUp struct {
		image.Point
		Button Button
	}
	MouScroll struct {image.Point}
	KeyType struct {Rune rune}
	KeyDown struct {Key Key}
	KeyUp struct {Key Key}
	KeyRepeat struct {Key Key}
)

func (wc WinClose)  String() string { return "window-close" }
func (mm MouMove)   String() string { return fmt.Sprintf("mouse-move   {%d,%d}", mm.X, mm.Y) }
func (md MouDown)   String() string { return fmt.Sprintf("mouse-down   {%d,%d} %s", md.X, md.Y, md.Button) }
func (mu MouUp)     String() string { return fmt.Sprintf("mouse-up     {%d,%d} %s", mu.X, mu.Y, mu.Button) }
func (ms MouScroll) String() string { return fmt.Sprintf("mouse-scroll {%d,%d}", ms.X, ms.Y) }
func (kt KeyType)   String() string { return fmt.Sprintf("key-type     '%v'", string(kt.Rune)) }
func (kd KeyDown)   String() string { return fmt.Sprintf("key-down      %s", kd.Key) }
func (ku KeyUp)     String() string { return fmt.Sprintf("key-up        %s", ku.Key) }
func (kr KeyRepeat) String() string { return fmt.Sprintf("key-repeat    %s", kr.Key) }

func init() {
	runtime.LockOSThread()
}
