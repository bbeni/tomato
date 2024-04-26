// graphics stuff
package tomato

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"runtime"
	"strings"
	"sync"

	"github.com/go-gl/gl/v4.2-core/gl" // I hope it is supported on most systems
	"github.com/go-gl/glfw/v3.3/glfw"
)

// Lock the thread needed for glfw (and gl?)
func init() {
	runtime.LockOSThread()
}

// Direct access to the glfw.Window
var Win *glfw.Window
var GuiImg *image.RGBA

// @Todo rename Gui, because its not only for Gui stuff, its used for any 2d rendering... it's more like an overlay over existing gl stuff, so maybe hud?
// gl stuff
var GuiShader uint32
var GuiTexture uint32
var GuiQuadVAO uint32
var GuiQuad = []float32{
	//  X, Y, Z, U, V
	-1.0, 1.0, 1.0, 0.0, 0.0,
	1.0, -1.0, 1.0, 1.0, 1.0,
	-1.0, -1.0, 1.0, 0.0, 1.0,
	-1.0, 1.0, 1.0, 0.0, 0.0,
	1.0, 1.0, 1.0, 1.0, 0.0,
	1.0, -1.0, 1.0, 1.0, 1.0,
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

func Die() {
	dead = true
}

func Events() <-chan Ev {
	return outEvents
}

// setup everything with this function
func Create(width, height int, title string) error {
	if err := glfw.Init(); err != nil {
		return err
	}
	// @Todo add options?
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	//glfw.WindowHint(glfw.Decorated, glfw.False)
	glfw.WindowHint(glfw.Resizable, glfw.False)
	w, err := glfw.CreateWindow(width, height, title, nil, nil)

	if err != nil {
		return err
	}

	Win = w
	Win.MakeContextCurrent()

	if err = gl.Init(); err != nil {
		return err
	}

	if err = openGLSetup(); err != nil {
		return err
	}

	gl.ClearColor(0.9, 0.85, 0.3, 1.0)
	gl.Clear(gl.DEPTH_BUFFER_BIT | gl.COLOR_BUFFER_BIT)
	Win.SwapBuffers()
	gl.Clear(gl.DEPTH_BUFFER_BIT | gl.COLOR_BUFFER_BIT)

	eventsSetup()
	return nil
}

var dead bool
var inEvents chan Ev
var outEvents chan Ev

// The programmer is responsible for using the appropriate Fields
// I know this is kinda ugly, but whatever..
// @Todo: We can still later introduce an interface for type checking
type Ev struct {
	Kind        EvKind
	image.Point        // MouMove, MouScroll, MouUp, MouDown
	Button      Button // MouUp,   MouDown
	Key         Key    // KeyDown, KeyUp,     KeyRepeat
	Rune        rune   // RuneTyped
}

func (ev Ev) String() string {
	return fmt.Sprintf("[%v Ev]{%v %v %v %v}", ev.Kind, ev.Key, string(ev.Rune), ev.Point, ev.Button)
}

//go:generate stringer -type=EvKind
type EvKind uint8

const (
	_ EvKind = iota
	WinClose
	MouMove
	MouDown
	MouUp
	MouScroll
	KeyDown
	KeyUp
	KeyRepeat
	RuneTyped
)

//go:generate stringer -type=Button
type Button uint8

const (
	MouseLeft Button = iota
	MouseRight
	MouseMiddle
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

//
// @Speed: Is a map efficient enough?
//

var buttons = map[glfw.MouseButton]Button{
	glfw.MouseButtonLeft:   MouseLeft,
	glfw.MouseButtonRight:  MouseRight,
	glfw.MouseButtonMiddle: MouseMiddle,
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

// @Todo store in a struct
var MouseX, MouseY int
var MouseDownL, MouseDownM, MouseDownR bool

// function adapted from faiface/gui
func eventsSetup() {

	inEvents = make(chan Ev)
	outEvents = make(chan Ev)

	go func() {

		var queue []Ev

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
		MouseX, MouseY = int(x), int(y)
		inEvents <- Ev{
			Kind:  MouMove,
			Point: image.Pt(MouseX, MouseY),
		}
	})

	Win.SetMouseButtonCallback(func(_ *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
		b, ok := buttons[button]
		if !ok {
			return
		}
		switch action {
		case glfw.Press:
			switch b {
			case MouseLeft:
				MouseDownL = true
			case MouseMiddle:
				MouseDownM = true
			case MouseRight:
				MouseDownR = true
			}

			inEvents <- Ev{
				Kind:   MouDown,
				Point:  image.Pt(MouseX, MouseY),
				Button: b,
			}
		case glfw.Release:
			switch b {
			case MouseLeft:
				MouseDownL = false
			case MouseMiddle:
				MouseDownM = false
			case MouseRight:
				MouseDownR = false
			}
			inEvents <- Ev{
				Kind:   MouUp,
				Point:  image.Pt(MouseX, MouseY),
				Button: b,
			}
		}
	})

	Win.SetScrollCallback(func(_ *glfw.Window, xoff, yoff float64) {
		inEvents <- Ev{
			Kind:  MouScroll,
			Point: image.Pt(int(xoff), int(yoff)),
		}
	})

	Win.SetCharCallback(func(_ *glfw.Window, r rune) {
		inEvents <- Ev{
			Kind: RuneTyped,
			Rune: r,
		}
	})

	Win.SetKeyCallback(func(_ *glfw.Window, key glfw.Key, _ int, action glfw.Action, _ glfw.ModifierKey) {
		k, ok := keys[key]
		if !ok {
			return
		}
		switch action {
		case glfw.Press:
			inEvents <- Ev{
				Kind: KeyDown,
				Key:  k,
			}
		case glfw.Release:
			inEvents <- Ev{
				Kind: KeyUp,
				Key:  k,
			}
		case glfw.Repeat:
			inEvents <- Ev{
				Kind: KeyRepeat,
				Key:  k,
			}
		}
	})

	Win.SetFramebufferSizeCallback(func(_ *glfw.Window, width, height int) {
		//@Todo: handle resizing
	})

	Win.SetCloseCallback(func(_ *glfw.Window) {
		inEvents <- Ev{
			Kind: WinClose,
		}
	})
}

func openGLSetup() error {
	var err error
	var guiShaderSource = `
		#version 420

		in vec3 vert;
		in vec2 vertTexCoord;
		out vec2 fragTexCoord;

		void main() {
			fragTexCoord = vertTexCoord;
			gl_Position = vec4(vert.xy, 0.0, 1.0);
		}
		#define FRAGMENT_SHADER
		#version 420

		uniform sampler2D tex;
		in vec2 fragTexCoord;

		out vec4 outputColor;

		void main() {
			outputColor = texture(tex, fragTexCoord);
		}
	`

	GuiShader, err = NewGLProgram(guiShaderSource)

	if err != nil {
		fmt.Print("\nERROR making GuiShader: ")
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

	return nil
}

type drawOp struct {
	where image.Rectangle
	img   image.Image
}

// @Memory prealocate memory maybe?
var drawQueue []drawOp
var drawLock sync.Mutex

// An Image to draw on the screen at Rectangle r
// when Draw() is called all is rendered.
func ToDraw(r image.Rectangle, img image.Image) {
	drawLock.Lock()
	drawQueue = append(drawQueue, drawOp{
		where: r,
		img:   img,
	})
	drawLock.Unlock()
}

func Clear() {
	clearColor := []byte{255, 255, 255, 255}
	gl.ClearTexImage(GuiTexture, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&clearColor[0]))
	GuiImg = image.NewRGBA(GuiImg.Bounds())
	gl.Clear(gl.DEPTH_BUFFER_BIT | gl.COLOR_BUFFER_BIT)
}

// If the Gui is on the texture on the GPU, the additional draw calls don't matter anyway
// we could just Draw it using Alpha blending

func Draw() {
	gl.UseProgram(GuiShader)
	gl.Enable(gl.BLEND)
	//gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)       // Assume premultiplied alpha
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA) // Non-premultipled version

	drawLock.Lock()

	// @Speed use union of all bounds instead...
	bounds := GuiImg.Bounds()
	for _, op := range drawQueue {
		op.where = op.where.Intersect(bounds)
		if op.where.Empty() {
			continue
		}
		draw.Draw(GuiImg, op.where, op.img, image.ZP, draw.Src)
	}

	gl.TextureSubImage2D(
		GuiTexture,
		0,
		int32(bounds.Min.X),
		int32(bounds.Min.Y),
		int32(bounds.Dx()),
		int32(bounds.Dy()),
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(GuiImg.Pix))

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	{
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, GuiTexture)
		gl.BindVertexArray(GuiQuadVAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 6*2*3)
	}

	// reset draw queue
	drawQueue = drawQueue[:0]
	drawLock.Unlock()

	gl.Disable(gl.BLEND)
	gl.Disable(gl.DEPTH_TEST)
}

func NewGLProgram(shaderSource string) (uint32, error) {

	shaderSources := strings.Split(shaderSource, "#define FRAGMENT_SHADER")

	if len(shaderSources) != 2 {
		return 0, errors.New("Syntax Error: tomato style shader source needs `#define FRAGMENT_SHADER` that separates vertex/fragment shader! (this is because we only want one shader source file!)\n")
	}

	vertexShaderSource := shaderSources[0] + "\x00"
	fragmentShaderSource := shaderSources[1] + "\x00"

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

func newScreenTexture(width, height int) uint32 {

	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	if rgba.Stride != rgba.Rect.Size().X*4 {
		panic("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), image.NewUniform(color.RGBA{0, 0, 0, 0}), image.Point{0, 0}, draw.Src)

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
