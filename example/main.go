package main

import (
	"fmt"
	"github.com/dumkin/gltext"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/image/math/fixed"
	"os"
	"runtime"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	err := glfw.Init()
	if err != nil {
		panic("glfw error")
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 5)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

	window, err := glfw.CreateWindow(640, 480, "Testing", nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	fmt.Println("Opengl version", gl.GoStr(gl.GetString(gl.VERSION)))

	fd, err := os.Open("font/Roboto.ttf")
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	font, err := gltext.LoadTruetype("fontconfigs")
	defer font.Release()

	if err == nil {
		fmt.Println("Font loaded from disk...")
	} else {
		// scale := fixed.Int26_6(32)
		scale := fixed.Int26_6(24)
		runesPerRow := fixed.Int26_6(128)

		runeRanges := make(gltext.RuneRanges, 0)

		runeRange := gltext.RuneRange{Low: 0x3000, High: 0x3030}
		runeRanges = append(runeRanges, runeRange)
		runeRange = gltext.RuneRange{Low: 0x3040, High: 0x309f}
		runeRanges = append(runeRanges, runeRange)
		runeRange = gltext.RuneRange{Low: 0x30a0, High: 0x30ff}
		runeRanges = append(runeRanges, runeRange)
		runeRange = gltext.RuneRange{Low: 0x4e00, High: 0x9faf}
		runeRanges = append(runeRanges, runeRange)
		runeRange = gltext.RuneRange{Low: 0xff00, High: 0xffef}
		runeRanges = append(runeRanges, runeRange)

		font, err = gltext.NewTruetype(fd, scale, runeRanges, runesPerRow)
		if err != nil {
			panic(err)
		}
		err = font.Config.Save("fontconfigs")
		if err != nil {
			panic(err)
		}
	}

	width, height := window.GetSize()
	font.ResizeWindow(float32(width), float32(height))

	scaleMin, scaleMax := float32(1.0), float32(1.1)

	text := gltext.NewText(font, scaleMin, scaleMax)
	defer text.Release()

	str := "梅干しが大好き。ウメボシガダイスキ。"
	for _, s := range str {
		fmt.Printf("%c: %d\n", s, rune(s))
	}

	text.SetString(str)
	text.SetColor(mgl32.Vec3{0.0, 0.0, 0.0})

	gl.ClearColor(1.0, 1.0, 1.0, 0.0)

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)

		text.SetPosition(mgl32.Vec2{0, 0})
		text.Draw()

		window.SwapBuffers()
		glfw.PollEvents()
	}
}