// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gltext

import (
	"bufio"
	"fmt"
	"github.com/go-gl/glow/gl-core/3.3/gl"
	"github.com/go-gl/mathgl/mgl32"
	"image"
	"image/png"
	"os"
)

const debug = false

var vertexShaderSource string = `
#version 330

uniform mat4 matrix;

in vec4 position;
in vec2 uv;

out vec2 fragment_uv;

void main() {
  fragment_uv = uv;
  gl_Position = matrix * position;
}
` + "\x00"

var fragmentShaderSource string = `
#version 330

uniform sampler2D fragment_texture;
uniform float text_lowerbound;
uniform vec4 fragment_color_adjustment;

in vec2 fragment_uv;
out vec4 fragment_color;

void main() {
  vec4 color = texture(fragment_texture, fragment_uv);
  if(color.w > text_lowerbound) {
    color = fragment_color_adjustment;
  }
  fragment_color = color;
}
` + "\x00"

type Font struct {
	config                 *FontConfig // Character set for this font.
	textureID              uint32      // Holds the glyph texture id.
	maxGlyphWidth          int         // Largest glyph width.
	maxGlyphHeight         int         // Largest glyph height.
	program                uint32      // program compiled from shaders
	position               uint32      // Position of the shaders 'position' variable
	uv                     uint32      // Position of the shaders uv variable
	fragmentTextureUniform int32       // Position of the shaders fragment texture variable

	// The desired color of the text
	colorUniform int32
	color        mgl32.Vec4

	// The background of the image is transparent.  Using an arbitrary
	// lower limit to distinguish between the background and the text.
	// There must be a better way that preserves the gradient-like
	// appearance of the text that the freetype-go library produces.
	textLowerBoundUniform int32
	textLowerBound        float32

	glMatrix      int32
	vao           uint32
	vbo           uint32
	ebo           uint32
	textureWidth  float32
	textureHeight float32
	windowWidth   float32
	windowHeight  float32
	ortho         mgl32.Mat4
	vboData       []float32
	vboIndexCount int
	eboData       []int32
	eboIndexCount int

	// X1, X2: the lower left and upper right points of a box that bounds the text
	X1 Point
	X2 Point
}

func loadFont(img *image.RGBA, config *FontConfig) (f *Font, err error) {
	f = new(Font)
	f.config = config
	f.textLowerBound = 0.4 // lower numbers make fatter text

	// Resize image to next power-of-two.
	img = Pow2Image(img).(*image.RGBA)
	ib := img.Bounds()

	f.textureWidth = float32(ib.Dx())
	f.textureHeight = float32(ib.Dy())

	for _, glyph := range config.Glyphs {
		if glyph.Width > f.maxGlyphWidth {
			f.maxGlyphWidth = glyph.Width
		}
		if glyph.Height > f.maxGlyphHeight {
			f.maxGlyphHeight = glyph.Height
		}
	}

	// save to disk for testing
	if debug {
		file, err := os.Create("out.png")
		if err != nil {
			panic(err)
		}
		defer file.Close()

		b := bufio.NewWriter(file)
		err = png.Encode(b, img)
		if err != nil {
			panic(err)
		}
		err = b.Flush()
		if err != nil {
			panic(err)
		}
	}

	// generate texture
	gl.GenTextures(1, &f.textureID)
	gl.BindTexture(gl.TEXTURE_2D, f.textureID)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(ib.Dx()),
		int32(ib.Dy()),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(img.Pix),
	)
	gl.BindTexture(gl.TEXTURE_2D, 0)

	// create shader program and define attributes and uniforms
	f.program, err = NewProgram(vertexShaderSource, fragmentShaderSource)
	if err != nil {
		return f, err
	}

	f.glMatrix = gl.GetUniformLocation(f.program, gl.Str("matrix\x00"))
	f.position = uint32(gl.GetAttribLocation(f.program, gl.Str("position\x00")))
	f.uv = uint32(gl.GetAttribLocation(f.program, gl.Str("uv\x00")))
	f.fragmentTextureUniform = gl.GetUniformLocation(f.program, gl.Str("fragment_texture\x00"))
	f.colorUniform = gl.GetUniformLocation(f.program, gl.Str("fragment_color_adjustment\x00"))
	f.textLowerBoundUniform = gl.GetUniformLocation(f.program, gl.Str("text_lowerbound\x00"))

	// size of glfloat
	glfloat_size := int32(4)

	// stride of the buffered data
	xy_count := int32(2)
	stride := xy_count + int32(2)

	gl.GenVertexArrays(1, &f.vao)
	gl.GenBuffers(1, &f.vbo)
	gl.GenBuffers(1, &f.ebo)

	// vao
	gl.BindVertexArray(f.vao)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, f.textureID)

	// vbo
	// specify the buffer for which the VertexAttribPointer calls apply
	gl.BindBuffer(gl.ARRAY_BUFFER, f.vbo)

	gl.EnableVertexAttribArray(f.position)
	gl.VertexAttribPointer(
		f.position,
		2,
		gl.FLOAT,
		false,
		glfloat_size*stride,
		gl.PtrOffset(0),
	)

	gl.EnableVertexAttribArray(f.uv)
	gl.VertexAttribPointer(
		f.uv,
		2,
		gl.FLOAT,
		false,
		glfloat_size*stride,
		gl.PtrOffset(int(glfloat_size*xy_count)),
	)

	// ebo
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, f.ebo)

	// i am guessing that order is important here
	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)

	return f, nil
}

func (f *Font) ResizeWindow(width float32, height float32) {
	f.windowWidth = width
	f.windowHeight = height
	f.ortho = mgl32.Ortho2D(0, f.windowWidth, 0, f.windowHeight)
}

// Low returns the font's lower rune bound.
func (f *Font) Low() rune { return f.config.Low }

// High returns the font's upper rune bound.
func (f *Font) High() rune { return f.config.High }

// Glyphs returns the font's glyph descriptors.
func (f *Font) Glyphs() Charset { return f.config.Glyphs }

// Release releases font resources.
// A font can no longer be used for rendering after this call completes.
func (f *Font) Release() {
	gl.DeleteTextures(1, &f.textureID)
	gl.DeleteBuffers(1, &f.vbo)
	gl.DeleteBuffers(1, &f.ebo)
	gl.DeleteBuffers(1, &f.vao)
	f.config = nil
}

func (f *Font) SetColor(r, g, b, a float32) {
	f.color = mgl32.Vec4{r, g, b, a}
}

func (f *Font) SetTextLowerBound(v float32) {
	f.textLowerBound = v
}

func (f *Font) SetString(fs string, argv ...interface{}) (Point, Point) {
	indices := []rune(fmt.Sprintf(fs, argv...))
	if len(indices) == 0 {
		return Point{}, Point{}
	}
	// ebo, vbo data
	f.vboIndexCount = len(indices) * 4 * 2 * 2 // 4 indexes per rune (containing 2 position + 2 texture)
	f.eboIndexCount = len(indices) * 6         // each rune requires 6 triangle indices for a quad
	f.vboData = make([]float32, f.vboIndexCount, f.vboIndexCount)
	f.eboData = make([]int32, f.eboIndexCount, f.eboIndexCount)
	f.makeBufferData(indices)
	return f.X1, f.X2
}

func (f *Font) SetPosition(x, y float32) {
	f.setDataPosition(x, y)
	if debug {
		fmt.Printf("ortho matrix\n%v\n", f.ortho)
		fmt.Printf("vbo data\n%v\n", f.vboData)
		fmt.Printf("ebo data\n%v\n", f.eboData)
	}
	glfloat_size := int32(4)

	// setup context
	gl.BindVertexArray(f.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, f.vbo)
	gl.BufferData(
		gl.ARRAY_BUFFER, int(glfloat_size)*f.vboIndexCount, gl.Ptr(f.vboData), gl.DYNAMIC_DRAW)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, f.ebo)
	gl.BufferData(
		gl.ELEMENT_ARRAY_BUFFER, int(glfloat_size)*f.eboIndexCount, gl.Ptr(f.eboData), gl.DYNAMIC_DRAW)
	gl.BindVertexArray(0)
	// completed context

	// not necesssary, but i just want to better understand using vertex arrays
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
	return
}

func (f *Font) Draw() {
	gl.UseProgram(f.program)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, f.textureID)
	gl.Uniform1i(f.fragmentTextureUniform, 0)
	gl.Uniform1f(f.textLowerBoundUniform, f.textLowerBound)
	gl.Uniform4fv(f.colorUniform, 1, &f.color[0])
	gl.UniformMatrix4fv(f.glMatrix, 1, false, &f.ortho[0])

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.BindVertexArray(f.vao)
	gl.DrawElements(gl.TRIANGLES, int32(f.eboIndexCount), gl.UNSIGNED_INT, nil)
	gl.BindVertexArray(0)
	gl.Disable(gl.BLEND)
}

func (f *Font) getBoundingBox(vboIndex int) {
	// index -4: x, index -3: y, index -2: uv's x, index -1 uv's y
	x := f.vboData[vboIndex-4]
	y := f.vboData[vboIndex-3]

	if vboIndex-4 == 0 {
		f.X1.X = x
		f.X1.Y = y
	} else {
		if x < f.X1.X {
			f.X1.X = x
		}
		if y < f.X1.Y {
			f.X1.Y = y
		}
		if x > f.X2.X {
			f.X2.X = x
		}
		if y > f.X2.Y {
			f.X2.Y = y
		}
	}
}

// all text originally sits at point (0,0) which is the
// lower left hand corner of the screen.
func (f *Font) setDataPosition(x, y float32) {
	length := len(f.vboData)
	for index := 0; index < length; {
		// index (0,0)
		f.vboData[index] += x
		index++
		f.vboData[index] += y
		index += 3 // skip texture data

		// index (1,0)
		f.vboData[index] += x
		index++
		f.vboData[index] += y
		index += 3

		// index (1,1)
		f.vboData[index] += x
		index++
		f.vboData[index] += y
		index += 3

		// index (0,1)
		f.vboData[index] += x
		index++
		f.vboData[index] += y
		index += 3
	}
	// update screen position
	f.X1.X += x
	f.X2.X += x
	f.X1.Y += y
	f.X2.Y += y
}

// currently only supports left to right text flow
func (f *Font) makeBufferData(indices []rune) {
	glyphs := f.config.Glyphs
	low := f.config.Low

	vboIndex := 0
	eboIndex := 0
	lineX := float32(0)
	eboOffset := int32(0)
	for _, r := range indices {
		r -= low
		if r >= 0 && int(r) < len(glyphs) {
			vw := float32(glyphs[r].Width)
			vh := float32(glyphs[r].Height)
			tP1, tP2 := glyphs[r].GetIndices(f)

			// counter-clockwise quad

			// index (0,0)
			f.vboData[vboIndex] = lineX // position
			vboIndex++
			f.vboData[vboIndex] = 0
			vboIndex++
			f.vboData[vboIndex] = tP1.X // texture uv
			vboIndex++
			f.vboData[vboIndex] = tP2.Y
			vboIndex++
			f.getBoundingBox(vboIndex)

			// index (1,0)
			f.vboData[vboIndex] = lineX + vw
			vboIndex++
			f.vboData[vboIndex] = 0
			vboIndex++
			f.vboData[vboIndex] = tP2.X
			vboIndex++
			f.vboData[vboIndex] = tP2.Y
			vboIndex++
			f.getBoundingBox(vboIndex)

			// index (1,1)
			f.vboData[vboIndex] = lineX + vw
			vboIndex++
			f.vboData[vboIndex] = vh
			vboIndex++
			f.vboData[vboIndex] = tP2.X
			vboIndex++
			f.vboData[vboIndex] = tP1.Y
			vboIndex++
			f.getBoundingBox(vboIndex)

			// index (0,1)
			f.vboData[vboIndex] = lineX
			vboIndex++
			f.vboData[vboIndex] = vh
			vboIndex++
			f.vboData[vboIndex] = tP1.X
			vboIndex++
			f.vboData[vboIndex] = tP1.Y
			vboIndex++
			f.getBoundingBox(vboIndex)

			advance := float32(glyphs[r].Advance)
			lineX += advance

			// ebo data
			f.eboData[eboIndex] = 0 + eboOffset
			eboIndex++
			f.eboData[eboIndex] = 1 + eboOffset
			eboIndex++
			f.eboData[eboIndex] = 2 + eboOffset
			eboIndex++

			f.eboData[eboIndex] = 0 + eboOffset
			eboIndex++
			f.eboData[eboIndex] = 2 + eboOffset
			eboIndex++
			f.eboData[eboIndex] = 3 + eboOffset
			eboIndex++
			eboOffset += 4
		}
	}
	return
}

// GlyphBounds returns the largest width and height for any of the glyphs
// in the font. This constitutes the largest possible bounding box
// a single glyph will have.
func (f *Font) GlyphBounds() (int, int) {
	return f.maxGlyphWidth, f.maxGlyphHeight
}
