// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package v45

import (
	"fmt"
	"github.com/4ydx/gltext"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

var boxVertexShaderSource = `
#version 330

uniform mat4 orthographic_matrix;
uniform vec2 final_position;

in vec4 centered_position;

void main() {
  vec4 center = orthographic_matrix * centered_position;
  gl_Position = vec4(center.x + final_position.x, center.y + final_position.y, center.z, center.w);
}
` + "\x00"

var boxFragmentShaderSource = `
#version 330

out vec4 fragment_color;

void main() {
  fragment_color = vec4(0.3,0.3,0.3,1);
}
` + "\x00"

// BoundingBox is the GL implementation of the bounding box of an object
type BoundingBox struct {
	program uint32 // program compiled from shaders

	// font holds our orthographic matrix
	font *Font

	// attributes
	centeredPosition uint32 // vertex position

	// the final screen position post-scaling
	finalPositionUniform int32
	finalPosition        mgl32.Vec2

	// transform to orthographic projection
	orthographicMatrixUniform int32

	vao           uint32
	vbo           uint32
	ebo           uint32
	windowWidth   float32
	windowHeight  float32
	vboData       []float32
	vboIndexCount int
	eboData       []int32
	eboIndexCount int
}

func loadBoundingBox(f *Font, lowerLeft gltext.Point, upperRight gltext.Point) (b *BoundingBox, err error) {
	b = new(BoundingBox)
	b.font = f

	// create shader program and define attributes and uniforms
	b.program, err = NewProgram(boxVertexShaderSource, boxFragmentShaderSource)
	if err != nil {
		return b, err
	}

	// ebo, vbo data
	b.vboIndexCount = 4 * 2 // 4 indexes per bounding box (containing 2 position)
	b.eboIndexCount = 6     // each rune requires 6 triangle indices for a quad
	b.vboData = make([]float32, b.vboIndexCount, b.vboIndexCount)
	b.eboData = make([]int32, b.eboIndexCount, b.eboIndexCount)
	b.makeBufferData(lowerLeft, upperRight)

	if gltext.IsDebug {
		prefix := gltext.DebugPrefix()
		fmt.Printf("%s bounding %v %v\n", prefix, lowerLeft, upperRight)
		fmt.Printf("%s bounding vbo data\n%v\n", prefix, b.vboData)
		fmt.Printf("%s bounding ebo data\n%v\n", prefix, b.eboData)
	}

	// attributes
	b.centeredPosition = uint32(gl.GetAttribLocation(b.program, gl.Str("centered_position\x00")))

	// uniforms
	b.finalPositionUniform = gl.GetUniformLocation(b.program, gl.Str("final_position\x00"))
	b.orthographicMatrixUniform = gl.GetUniformLocation(b.program, gl.Str("orthographic_matrix\x00"))

	// size of glfloat
	glfloatSize := int32(4)

	gl.GenVertexArrays(1, &b.vao)
	gl.GenBuffers(1, &b.vbo)
	gl.GenBuffers(1, &b.ebo)

	// vao
	gl.BindVertexArray(b.vao)

	// vbo
	// specify the buffer for which the VertexAttribPointer calls apply
	gl.BindBuffer(gl.ARRAY_BUFFER, b.vbo)

	gl.EnableVertexAttribArray(b.centeredPosition)
	gl.VertexAttribPointer(
		b.centeredPosition,
		2,
		gl.FLOAT,
		false,
		0,
		gl.PtrOffset(0),
	)
	gl.BufferData(gl.ARRAY_BUFFER, int(glfloatSize)*b.vboIndexCount, gl.Ptr(b.vboData), gl.DYNAMIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, b.ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, int(glfloatSize)*b.eboIndexCount, gl.Ptr(b.eboData), gl.DYNAMIC_DRAW)
	gl.BindVertexArray(0)

	// not necesssary, but i just want to better understand using vertex arrays
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)

	return b, nil
}

// Release opengl objects
func (b *BoundingBox) Release() {
	gl.DeleteBuffers(1, &b.vbo)
	gl.DeleteBuffers(1, &b.ebo)
	gl.DeleteBuffers(1, &b.vao)
}

// Draw the bounding box
func (b *BoundingBox) Draw() {
	gl.UseProgram(b.program)

	// uniforms
	gl.Uniform2fv(b.finalPositionUniform, 1, &b.finalPosition[0])
	gl.UniformMatrix4fv(b.orthographicMatrixUniform, 1, false, &b.font.OrthographicMatrix[0])

	// draw
	gl.BindVertexArray(b.vao)
	gl.DrawElements(gl.TRIANGLES, int32(b.eboIndexCount), gl.UNSIGNED_INT, nil)
	gl.BindVertexArray(0)
}

func (b *BoundingBox) makeBufferData(lowerLeft, upperRight gltext.Point) {
	// counter-clockwise quad

	// index (0,0)
	b.vboData[0] = lowerLeft.X // position
	b.vboData[1] = lowerLeft.Y

	// index (1,0)
	b.vboData[2] = upperRight.X
	b.vboData[3] = lowerLeft.Y

	// index (1,1)
	b.vboData[4] = upperRight.X
	b.vboData[5] = upperRight.Y

	// index (0,1)
	b.vboData[6] = lowerLeft.X
	b.vboData[7] = upperRight.Y

	// ebo data
	b.eboData[0] = 0
	b.eboData[1] = 1
	b.eboData[2] = 2
	b.eboData[3] = 0
	b.eboData[4] = 2
	b.eboData[5] = 3
	return
}
