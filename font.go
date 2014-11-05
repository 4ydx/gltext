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

in vec2 fragment_uv;
out vec4 fragment_color;

void main() {
  vec4 color     = texture(fragment_texture, fragment_uv);
  color.a        = max(color.a, 0.4);
  fragment_color = color;
}
` + "\x00"

type Font struct {
	config         *FontConfig // Character set for this font.
	textureID      uint32      // Holds the glyph texture id.
	maxGlyphWidth  int         // Largest glyph width.
	maxGlyphHeight int         // Largest glyph height.
	program        uint32      // program compiled from shaders
	vboSize        int32
	position       uint32
	uv             uint32
	fragmentTexure int32
	glMatrix       int32
	vao            uint32
	vbo            uint32
	ebo            uint32
	textureWidth   float32
	textureHeight  float32
	windowWidth    float32
	windowHeight   float32
	ortho          mgl32.Mat4
}

func loadFont(img *image.RGBA, config *FontConfig) (f *Font, err error) {
	f = new(Font)
	f.config = config

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
	f.vboSize = 0
	f.program, err = newProgram(vertexShaderSource, fragmentShaderSource)
	if err != nil {
		return f, err
	}

	f.glMatrix = gl.GetUniformLocation(f.program, gl.Str("matrix\x00"))
	f.position = uint32(gl.GetAttribLocation(f.program, gl.Str("position\x00")))
	f.uv = uint32(gl.GetAttribLocation(f.program, gl.Str("uv\x00")))
	f.fragmentTexure = gl.GetUniformLocation(f.program, gl.Str("fragment_texture\x00"))

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

func (f *Font) Printf(x, y float32, fs string, argv ...interface{}) error {
	indices := []rune(fmt.Sprintf(fs, argv...))
	if len(indices) == 0 {
		return nil
	}
	// ebo, vbo data
	vboIndexCount := len(indices) * 4 * 2 * 2 // 4 indexes per rune (containing 2 position + 2 texture)
	eboIndexCount := len(indices) * 6         // each rune requires 6 triangle edges to complete a quad
	vboData := make([]float32, vboIndexCount, vboIndexCount)
	eboData := make([]int32, eboIndexCount, eboIndexCount)
	f.makeData(x, y, indices, vboData, eboData)

	if debug {
		fmt.Printf("ortho matrix\n%v\n", f.ortho)
		fmt.Printf("vbo data\n%v\n", vboData)
		fmt.Printf("ebo data\n%v\n", eboData)
	}

	glfloat_size := int32(4)

	// setup context
	gl.BindVertexArray(f.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, f.vbo)
	gl.BufferData(
		gl.ARRAY_BUFFER, int(glfloat_size)*vboIndexCount, gl.Ptr(vboData), gl.DYNAMIC_DRAW)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, f.ebo)
	gl.BufferData(
		gl.ELEMENT_ARRAY_BUFFER, int(glfloat_size)*eboIndexCount, gl.Ptr(eboData), gl.DYNAMIC_DRAW)
	gl.BindVertexArray(0)
	// completed context

	// not necesssary, but i just want to better understand using vertex arrays
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)

	// draw
	gl.UseProgram(f.program)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, f.textureID)
	gl.Uniform1i(f.fragmentTexure, 0)
	gl.UniformMatrix4fv(f.glMatrix, 1, false, &f.ortho[0])

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.BindVertexArray(f.vao)
	gl.DrawElements(gl.TRIANGLES, int32(eboIndexCount), gl.UNSIGNED_INT, nil)
	gl.BindVertexArray(0)
	gl.Disable(gl.BLEND)

	return nil
}

func (f *Font) makeData(x, y float32, indices []rune, vboData []float32, eboData []int32) {
	glyphs := f.config.Glyphs
	low := f.config.Low

	lineX := float32(0)
	vboIndex := 0
	eboIndex := 0
	eboOffset := int32(0)
	for _, r := range indices {
		r -= low
		if r >= 0 && int(r) < len(glyphs) {
			vw := float32(glyphs[r].Width)
			vh := float32(glyphs[r].Height)
			tP1, tP2 := glyphs[r].GetIndices(f)

			// counter-clockwise quad

			// index (0,0)
			vboData[vboIndex] = lineX + x // position
			vboIndex++
			vboData[vboIndex] = 0 + y
			vboIndex++
			vboData[vboIndex] = tP1.X // texture uv
			vboIndex++
			vboData[vboIndex] = tP2.Y
			vboIndex++

			// index (1,0)
			vboData[vboIndex] = lineX + vw + x
			vboIndex++
			vboData[vboIndex] = 0 + y
			vboIndex++
			vboData[vboIndex] = tP2.X
			vboIndex++
			vboData[vboIndex] = tP2.Y
			vboIndex++

			// index (1,1)
			vboData[vboIndex] = lineX + vw + x
			vboIndex++
			vboData[vboIndex] = vh + y
			vboIndex++
			vboData[vboIndex] = tP2.X
			vboIndex++
			vboData[vboIndex] = tP1.Y
			vboIndex++

			// index (0,1)
			vboData[vboIndex] = lineX + x
			vboIndex++
			vboData[vboIndex] = vh + y
			vboIndex++
			vboData[vboIndex] = tP1.X
			vboIndex++
			vboData[vboIndex] = tP1.Y
			vboIndex++

			advance := float32(glyphs[r].Advance)
			lineX += advance

			// ebo data
			eboData[eboIndex] = 0 + eboOffset
			eboIndex++
			eboData[eboIndex] = 1 + eboOffset
			eboIndex++
			eboData[eboIndex] = 2 + eboOffset
			eboIndex++

			eboData[eboIndex] = 0 + eboOffset
			eboIndex++
			eboData[eboIndex] = 2 + eboOffset
			eboIndex++
			eboData[eboIndex] = 3 + eboOffset
			eboIndex++
			eboOffset += 4
		}
	}
}

// GlyphBounds returns the largest width and height for any of the glyphs
// in the font. This constitutes the largest possible bounding box
// a single glyph will have.
func (f *Font) GlyphBounds() (int, int) {
	return f.maxGlyphWidth, f.maxGlyphHeight
}
