package main

import (
    "time"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

//GO_NOTE// we can organise variables (var, const, etc.) into blocks
const (
	width  = 500
	height = 500
    rows = 100
    columns = 100
    threshold = 0.15
    //fps = 10
    
    // code to be run in the GPU Driver
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

//GO_NOTE// we can organise variables (var, const, etc.) into blocks
var (
	square = []float32 {   //GO_NOTE// array declaration of type float32
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

    alive bool
    aliveNext bool

    x int
    y int
}
//GO_NOTE// This is the syntax for creating a method
//GO_NOTE// Pointer in the brackets refers to the object the method is for
// c.draw() to be used within the draw() function
func (c *cell) draw() {
    if !c.alive {
        return
    }

    gl.BindVertexArray(c.drawable)
    gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square) / 3))
}

// checkState determines the state of the cell for the next tick of the game.
func (c *cell) checkState(cells [][]*cell) {
    c.alive = c.aliveNext
    c.aliveNext = c.alive

    liveCount := c.liveNeighbours(cells)
    if c.alive {
        // 1. Any live cell with fewer than two live neighbours dies, as if
        // caused by underpopulation.
        if liveCount < 2 {
            c.aliveNext = false
        }

        // 2. Any live cell with two or three live neighbours lives on to the
        // next generation.
        if liveCount == 2 || liveCount == 3 {
            c.aliveNext = true
        }

        // 3. Any live cell with more than three live neighbours dies, as if by
        // overpopulation.
        if liveCount > 3 {
            c.aliveNext = false
        }
    } else {
        // 4. Any dead cell with exactly three live neighbours becomes a live
        // cell, as if by reproduction.
        if liveCount == 3 {
            c.aliveNext = true
        }
    }
}

// liveNeighbours returns the number of live neighbours for a cell.
func (c *cell) liveNeighbours(cells [][]*cell) int {
    var liveCount int
    add := func(x, y int) {
        // If we're at an edge, check the other side of the board.
        if x == len(cells) {
            x = 0
        } else if x == -1 {
            x = len(cells) - 1
        }

        if y == len(cells[x]) {
            y = 0
        } else if y == -1 {
            y = len(cells[x]) - 1
        }

        if cells[x][y].alive {
            liveCount++
        }
    }

    add(c.x-1 ,c.y)     // To the left
    add(c.x+1 ,c.y)     // To the right
    add(c.x ,c.y+1)     // Up
    add(c.x ,c.y-1)     // Down
    add(c.x-1 ,c.y+1)   // Top-left
    add(c.x+1 ,c.y+1)   // Top-right
    add(c.x-1 ,c.y-1)   // Bottom-left
    add(c.x+1 ,c.y-1)   // Bottom-right

    return liveCount
}




func main() {
	runtime.LockOSThread()

	window := initGlfw()
	defer glfw.Terminate()

	program := initOpenGL()

	//vao := makeVao(square)
    cells := makeCells()


    // Loop terminates when the window is closed
	for !window.ShouldClose() { 
     //   t := time.Now()

        for x := range cells {
            for _, c := range cells[x] {
                c.checkState(cells)
            }
        }

        draw(cells, window, program)        

       // time.Sleep(time.Second/time.Duration(fps) - time.Since(t))
	}
}







// makeCells creates a multi-dimensional slice to represent the board
func makeCells () [][]*cell {
    rand.Seed(time.Now().UnixNano())

    cells := make([][]*cell, rows, rows)
    for x := 0; x < rows; x++ {
        for y := 0; y < columns; y++ {
            c := newCell(x, y)

            c.alive = rand.Float64() < threshold
            c.aliveNext = c.alive

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

    // Allow window to be resized
	glfw.WindowHint(glfw.Resizable, glfw.False)
    // ContextVersionMajor and ContextVersionMinor refer to the OpenGL Version,
    // which is version 4.1 (major.minor)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
    // OpenGLProfile refers to which type of OpenGL we want to use, Core is current
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
    // Compatible with future versions?
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
    

//GO_NOTE// first return(?) value for range is the index, the second is the
// object/value at the index
    for x := range cells {               //GO_NOTE// x is the index
        for _, c := range cells[x] {     // c is the object/value at the index
           c.draw()                      // _ is used to skip the index to avoid
        }                                // an error as it is not used 
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
