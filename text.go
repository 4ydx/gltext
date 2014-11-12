// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gltext

import (
	"fmt"
	"github.com/go-gl/glow/gl-core/3.3/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Text struct {
	font *Font

	// The desired color of the text
	color mgl32.Vec4

	// scaling the text
	Scale       float32
	ScaleMax    float32
	scaleMatrix mgl32.Mat4

	vao           uint32
	vbo           uint32
	ebo           uint32
	windowWidth   float32
	windowHeight  float32
	vboData       []float32
	vboIndexCount int
	eboData       []int32
	eboIndexCount int

	// X1, X2: the lower left and upper right points of a box that bounds the text
	X1 Point
	X2 Point
}

func LoadText(f *Font) (t *Text, err error) {
	t = new(Text)
	t.font = f

	// text hover values - implicit ScaleMin of 1.0
	t.ScaleMax = 1.1
	t.SetScale(1)

	// size of glfloat
	glfloat_size := int32(4)

	// stride of the buffered data
	xy_count := int32(2)
	stride := xy_count + int32(2)

	gl.GenVertexArrays(1, &t.vao)
	gl.GenBuffers(1, &t.vbo)
	gl.GenBuffers(1, &t.ebo)

	// vao
	gl.BindVertexArray(t.vao)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, t.font.textureID)

	// vbo
	// specify the buffer for which the VertexAttribPointer calls apply
	gl.BindBuffer(gl.ARRAY_BUFFER, t.vbo)

	gl.EnableVertexAttribArray(t.font.position)
	gl.VertexAttribPointer(
		t.font.position,
		2,
		gl.FLOAT,
		false,
		glfloat_size*stride,
		gl.PtrOffset(0),
	)

	gl.EnableVertexAttribArray(t.font.uv)
	gl.VertexAttribPointer(
		t.font.uv,
		2,
		gl.FLOAT,
		false,
		glfloat_size*stride,
		gl.PtrOffset(int(glfloat_size*xy_count)),
	)

	// ebo
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, t.ebo)

	// i am guessing that order is important here
	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)

	return t, nil
}

// Release releases font resources.
// A font can no longer be used for rendering after this call completes.
func (t *Text) Release() {
	gl.DeleteBuffers(1, &t.vbo)
	gl.DeleteBuffers(1, &t.ebo)
	gl.DeleteBuffers(1, &t.vao)
}

func (t *Text) SetScale(s float32) (changed bool) {
	if s > t.ScaleMax || s < 1.0 {
		return
	}
	changed = true
	t.Scale = s
	t.scaleMatrix = mgl32.Scale3D(s, s, s)
	return
}

func (t *Text) AddScale(s float32) (changed bool) {
	if s < 0 && t.Scale <= 1.0 {
		return
	}
	if s > 0 && t.Scale >= t.ScaleMax {
		return
	}
	changed = true
	t.Scale += s
	t.scaleMatrix = mgl32.Scale3D(t.Scale, t.Scale, t.Scale)
	return
}

func (t *Text) SetColor(r, g, b, a float32) {
	t.color = mgl32.Vec4{r, g, b, a}
}

func (t *Text) SetString(f *Font, fs string, argv ...interface{}) (Point, Point) {
	indices := []rune(fmt.Sprintf(fs, argv...))
	if len(indices) == 0 {
		return Point{}, Point{}
	}
	// ebo, vbo data
	t.vboIndexCount = len(indices) * 4 * 2 * 2 // 4 indexes per rune (containing 2 position + 2 texture)
	t.eboIndexCount = len(indices) * 6         // each rune requires 6 triangle indices for a quad
	t.vboData = make([]float32, t.vboIndexCount, t.vboIndexCount)
	t.eboData = make([]int32, t.eboIndexCount, t.eboIndexCount)
	t.makeBufferData(indices)
	return t.X1, t.X2
}

func (t *Text) SetPosition(x, y float32) {
	t.setDataPosition(x, y)
	if debug {
		fmt.Printf("vbo data\n%v\n", t.vboData)
		fmt.Printf("ebo data\n%v\n", t.eboData)
	}
	glfloat_size := int32(4)

	// setup context
	gl.BindVertexArray(t.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, t.vbo)
	gl.BufferData(
		gl.ARRAY_BUFFER, int(glfloat_size)*t.vboIndexCount, gl.Ptr(t.vboData), gl.DYNAMIC_DRAW)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, t.ebo)
	gl.BufferData(
		gl.ELEMENT_ARRAY_BUFFER, int(glfloat_size)*t.eboIndexCount, gl.Ptr(t.eboData), gl.DYNAMIC_DRAW)
	gl.BindVertexArray(0)

	// not necesssary, but i just want to better understand using vertex arrays
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
	return
}

func (t *Text) Draw() {
	gl.UseProgram(t.font.program)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, t.font.textureID)

	// uniforms
	gl.Uniform1i(t.font.fragmentTextureUniform, 0)
	gl.Uniform1f(t.font.textLowerBoundUniform, t.font.textLowerBound)
	gl.Uniform4fv(t.font.colorUniform, 1, &t.color[0])
	gl.UniformMatrix4fv(t.font.orthographicMatrixUniform, 1, false, &t.font.orthographicMatrix[0])
	gl.UniformMatrix4fv(t.font.scaleMatrixUniform, 1, false, &t.scaleMatrix[0])

	// draw
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.BindVertexArray(t.vao)
	gl.DrawElements(gl.TRIANGLES, int32(t.eboIndexCount), gl.UNSIGNED_INT, nil)
	gl.BindVertexArray(0)
	gl.Disable(gl.BLEND)
}

func (t *Text) getBoundingBox(vboIndex int) {
	// index -4: x, index -3: y, index -2: uv's x, index -1 uv's y
	x := t.vboData[vboIndex-4]
	y := t.vboData[vboIndex-3]

	if vboIndex-4 == 0 {
		t.X1.X = x
		t.X1.Y = y
	} else {
		if x < t.X1.X {
			t.X1.X = x
		}
		if y < t.X1.Y {
			t.X1.Y = y
		}
		if x > t.X2.X {
			t.X2.X = x
		}
		if y > t.X2.Y {
			t.X2.Y = y
		}
	}
}

// all text originally sits at point (0,0) which is the
// lower left hand corner of the screen.
func (t *Text) setDataPosition(x, y float32) {
	length := len(t.vboData)
	for index := 0; index < length; {
		// index (0,0)
		t.vboData[index] += x
		index++
		t.vboData[index] += y
		index += 3 // skip texture data

		// index (1,0)
		t.vboData[index] += x
		index++
		t.vboData[index] += y
		index += 3

		// index (1,1)
		t.vboData[index] += x
		index++
		t.vboData[index] += y
		index += 3

		// index (0,1)
		t.vboData[index] += x
		index++
		t.vboData[index] += y
		index += 3
	}
	// update screen position
	t.X1.X += x
	t.X2.X += x
	t.X1.Y += y
	t.X2.Y += y
}

// currently only supports left to right text flow
func (t *Text) makeBufferData(indices []rune) {
	glyphs := t.font.config.Glyphs
	low := t.font.config.Low

	vboIndex := 0
	eboIndex := 0
	lineX := float32(0)
	eboOffset := int32(0)
	for _, r := range indices {
		r -= low
		if r >= 0 && int(r) < len(glyphs) {
			vw := float32(glyphs[r].Width)
			vh := float32(glyphs[r].Height)
			tP1, tP2 := glyphs[r].GetIndices(t.font)

			// counter-clockwise quad

			// index (0,0)
			t.vboData[vboIndex] = lineX // position
			vboIndex++
			t.vboData[vboIndex] = 0
			vboIndex++
			t.vboData[vboIndex] = tP1.X // texture uv
			vboIndex++
			t.vboData[vboIndex] = tP2.Y
			vboIndex++
			t.getBoundingBox(vboIndex)

			// index (1,0)
			t.vboData[vboIndex] = lineX + vw
			vboIndex++
			t.vboData[vboIndex] = 0
			vboIndex++
			t.vboData[vboIndex] = tP2.X
			vboIndex++
			t.vboData[vboIndex] = tP2.Y
			vboIndex++
			t.getBoundingBox(vboIndex)

			// index (1,1)
			t.vboData[vboIndex] = lineX + vw
			vboIndex++
			t.vboData[vboIndex] = vh
			vboIndex++
			t.vboData[vboIndex] = tP2.X
			vboIndex++
			t.vboData[vboIndex] = tP1.Y
			vboIndex++
			t.getBoundingBox(vboIndex)

			// index (0,1)
			t.vboData[vboIndex] = lineX
			vboIndex++
			t.vboData[vboIndex] = vh
			vboIndex++
			t.vboData[vboIndex] = tP1.X
			vboIndex++
			t.vboData[vboIndex] = tP1.Y
			vboIndex++
			t.getBoundingBox(vboIndex)

			advance := float32(glyphs[r].Advance)
			lineX += advance

			// ebo data
			t.eboData[eboIndex] = 0 + eboOffset
			eboIndex++
			t.eboData[eboIndex] = 1 + eboOffset
			eboIndex++
			t.eboData[eboIndex] = 2 + eboOffset
			eboIndex++

			t.eboData[eboIndex] = 0 + eboOffset
			eboIndex++
			t.eboData[eboIndex] = 2 + eboOffset
			eboIndex++
			t.eboData[eboIndex] = 3 + eboOffset
			eboIndex++
			eboOffset += 4
		}
	}
	return
}