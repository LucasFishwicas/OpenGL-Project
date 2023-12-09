package main

import (
	"log"
    "fmt"
    "strings"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

const (
	width  = 500
	height = 500
    rows = 10
    columns = 10

    vertexShaderSource = `
        #version 410
        in vec3 vp;
        void main() {
            gl_Position = vec4(vp, 1.0);
        }
    ` + "\x00"

    fragmentShaderSource = `
        #version 410
        out vec4 frag_colour;
        void main() {
            frag_colour = vec4(4, 0, 1, 0);
        }
    ` + "\x00"
)

var (
	square = []float32 {
		-0.5, 0.5, 0,
        -0.5, -0.5, 0,
		0.5, -0.5, 0,

        -0.5, 0.5, 0,
        0.5, 0.5, 0,
        0.5, -0.5, 0,
	}
)




// Cell struct to represent each cell in the board
type cell struct {
    drawable uint32

    x int
    y int
}
// This is a method for the Cell struct
func (c *cell) draw() {
    gl.BindVertexArray(c.drawable)
    gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square) / 3))
}





func main() {
	runtime.LockOSThread()

	window := initGlfw()
	defer glfw.Terminate()

	program := initOpenGL()

	//vao := makeVao(square)
    cells := makeCells()

	for !window.ShouldClose() {
		//TODO
		draw(cells, window, program)
	}
}







// makeCells creates a multi-dimensional slice to represent the board
func makeCells () [][]*cell {
    cells := make([][]*cell, rows, rows)
    for x := 0; x < rows; x++ {
        for y := 0; y < columns; y++ {
            c := newCell(x, y)
            cells[x] = append(cells[x], c)
        }
    }

    return cells
}






func newCell(x, y int) *cell {
    points := make([]float32, len(square), len(square))
    copy(points, square)

    for i := 0; i < len(points); i++ {
        var position float32
        var size float32
        switch i % 3 {
            case 0:
                size = 1.0 / float32(columns)
                position = float32(x) * size
            case 1:
                size = 1.0 / float32(rows)
                position = float32(y) * size
            default:
                continue
        }

        if points[i] < 0 {
            points[i] = (position * 2) - 1
        } else {
            points[i] = ((position + size) * 2) - 1
        }
    }

    return &cell {
        drawable: makeVao(points),

        x: x,
        y: y,
    }
}





// compileShader receives shader source code as a string and its type
// then returns a pointer to the compiled shader
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






// initGlfw initialises glfw and returns a Window to use.
func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "Conway's Game of Life", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}





// initOpenGL initialises OpenGl and returns an initialised program.
func initOpenGL() uint32 {
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

    vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
    if err != nil {
        panic(err)
    }
    fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
    if err != nil {
        panic(err)
    }

	prog := gl.CreateProgram()
    gl.AttachShader(prog, vertexShader)
    gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)
	return prog
}





// draw will draw the cells, using the cells method, to visualise the game
func draw(cells [][]*cell, window *glfw.Window, prog uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(prog)

    for x:= range cells {
        for _, c := range cells[x] {
            c.draw()
        }
    }

	glfw.PollEvents()
	window.SwapBuffers()
}

// makeVao initialises and returns a vertex array from the points provided.
func makeVao(points []float32) uint32 {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	return vao
}
