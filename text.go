// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gltext

import (
	"fmt"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Align int

const (
	AlignLeft Align = iota
	AlignRight
	AlignCenter
)

// CharacterSide shows which side of a character is
// clicked
type CharacterSide int

const (
	CSLeft CharacterSide = iota
	CSRight
	CSUnknown
)

type Text struct {
	font *Font

	// final position on screen
	finalPosition mgl32.Vec2

	// text color
	color mgl32.Vec3

	// scaling the text
	Scale       float32
	ScaleMin    float32
	ScaleMax    float32
	scaleMatrix mgl32.Mat4

	// bounding box of text
	BoundingBox *BoundingBox

	// general opengl values
	vao           uint32
	vbo           uint32
	ebo           uint32
	vboData       []float32
	vboIndexCount int
	eboData       []int32
	eboIndexCount int

	// determines how many prefix characters are drawn on screen
	RuneCount int

	// no longer than this string
	MaxRuneCount int

	// X1, X2: the lower left and upper right points of a box that bounds the text with a center point (0,0)

	// lower left
	X1 Point
	// upper right
	X2 Point

	SetPositionX float32
	SetPositionY float32

	String      string
	CharSpacing []float32
}

func (t *Text) GetLength() int {
	return t.eboIndexCount / 6
}

// NewText creates a new text object with scaling boundaries
// the rest state of the text when not being interacted with
// is scaleMin.  most likely one wants to use 1.0.
func NewText(f *Font, scaleMin, scaleMax float32) (t *Text) {
	t = &Text{}
	t.font = f

	// text hover values
	// "resting state" of a text object is the min scale
	t.ScaleMin, t.ScaleMax = scaleMin, scaleMax
	t.SetScale(1)
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

	gl.EnableVertexAttribArray(t.font.centeredPosition)
	gl.VertexAttribPointer(
		t.font.centeredPosition,
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
	return t
}

// Release releases font resources.
// A font can no longer be used for rendering after this call completes.
func (t *Text) Release() {
	gl.DeleteBuffers(1, &t.vbo)
	gl.DeleteBuffers(1, &t.ebo)
	gl.DeleteVertexArrays(1, &t.vao)
}

// SetScale returns true when a change occured
func (t *Text) SetScale(s float32) bool {
	if s > t.ScaleMax || s < t.ScaleMin {
		return false
	}
	t.Scale = s
	t.scaleMatrix = mgl32.Scale3D(s, s, s)
	return true
}

// AddScale returns true when a change occured
func (t *Text) AddScale(s float32) bool {
	if s < 0 && t.Scale <= t.ScaleMin {
		return false
	}
	if s > 0 && t.Scale >= t.ScaleMax {
		return false
	}
	t.Scale += s
	t.scaleMatrix = mgl32.Scale3D(t.Scale, t.Scale, t.Scale)
	return true
}

func (t *Text) SetColor(r, g, b float32) {
	t.color = mgl32.Vec3{r, g, b}
}

// SetString performs creates new vbo and ebo objects as well as to perform all
// binding required for displaying text to screen
func (t *Text) SetString(fs string, argv ...interface{}) {
	indices := []rune(fmt.Sprintf(fs, argv...))
	if len(indices) == 0 {
		t.String = ""
		t.RuneCount = 0
		return
	}
	if t.MaxRuneCount > 0 && len(indices) > t.MaxRuneCount+1 {
		indices = indices[0:t.MaxRuneCount]
	}
	t.String = string(indices)

	// ebo, vbo data
	glfloat_size := int32(4)

	t.vboIndexCount = len(indices) * 4 * 2 * 2 // 4 indexes per rune (containing 2 position + 2 texture)
	t.eboIndexCount = len(indices) * 6         // each rune requires 6 triangle indices for a quad
	t.RuneCount = len(indices)
	t.vboData = make([]float32, t.vboIndexCount, t.vboIndexCount)
	t.eboData = make([]int32, t.eboIndexCount, t.eboIndexCount)

	// generate the basic vbo data and bounding box
	// center the vbo data around the orthographic (0,0) point
	t.X1 = Point{0, 0}
	t.X2 = Point{0, 0}
	t.makeBufferData(indices)
	t.centerTheData(t.getLowerLeft())

	if IsDebug {
		prefix := DebugPrefix()
		fmt.Printf("%s bounding box %v %v\n", prefix, t.X1, t.X2)
		fmt.Printf("%s text vbo data\n%v\n", prefix, t.vboData)
		fmt.Printf("%s text ebo data\n%v\n", prefix, t.eboData)
	}
	gl.BindVertexArray(t.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, t.vbo)
	gl.BufferData(
		gl.ARRAY_BUFFER, int(glfloat_size)*t.vboIndexCount, gl.Ptr(t.vboData), gl.DYNAMIC_DRAW)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, t.ebo)
	gl.BufferData(
		gl.ELEMENT_ARRAY_BUFFER, int(glfloat_size)*t.eboIndexCount, gl.Ptr(t.eboData), gl.DYNAMIC_DRAW)
	gl.BindVertexArray(0)

	// possibly not necesssary?
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)

	// SetString can be called at anytime.  we want to make sure that if the user is updating the text,
	// the previous position will be maintained
	t.SetPosition(t.SetPositionX, t.SetPositionY)
}

// The block of text is positioned around the center of the screen, which in this case must
// be considered (0,0).  This is necessary for orthographic projection and scaling to work
// well together.  If the text is *not* at (0,0), then scaling doesnt produce a direct zoom effect.
func (t *Text) getLowerLeft() (lowerLeft Point) {
	lineWidthHalf := (t.X2.X - t.X1.X) / 2
	lineHeightHalf := (t.X2.Y - t.X1.Y) / 2

	lowerLeft.X = -lineWidthHalf
	lowerLeft.Y = -lineHeightHalf
	return
}

// Requirement prior to calling SetPosition:
// The text's X1 and X2 values must be the bounding box with center (0,0)
func (t *Text) SetPosition(x, y float32) {
	if t.X1.X == 0 && t.X1.Y == 0 && t.X2.X == 0 && t.X2.Y == 0 {
		TextDebug("Bounding Box Is Unset")
	}
	// transform to orthographic coordinates ranged -1 to 1 for the shader
	t.finalPosition[0] = x / (t.font.WindowWidth / 2)
	t.finalPosition[1] = y / (t.font.WindowHeight / 2)
	if IsDebug {
		t.BoundingBox.finalPosition[0] = x / (t.font.WindowWidth / 2)
		t.BoundingBox.finalPosition[1] = y / (t.font.WindowHeight / 2)
	}

	// used for detecting clicks, hovers, etc
	t.X1.X += x
	t.X1.Y += y
	t.X2.X += x
	t.X2.Y += y

	// used to build shadow data and for calling SetPosition again when needed
	t.SetPositionX = x
	t.SetPositionY = y
}

func (t *Text) Justify(align Align) {
	// calculate left aligned text location
	sign := 1
	if align == AlignRight {
		sign = -1
	}
	x := t.SetPositionX + float32(sign)*(t.X2.X-t.X1.X)/2
	y := t.SetPositionY

	// SetPosition requires the text bounding box be centered on (0,0).
	// so calculate its original value
	tX2 := Point{X: (t.X2.X - t.X1.X) / 2, Y: (t.X2.Y - t.X1.Y) / 2}
	tX1 := Point{X: -tX2.X, Y: -tX2.Y}
	t.X2 = tX2
	t.X1 = tX1

	t.SetPosition(x, y)
}

func (t *Text) Draw() {
	if IsDebug {
		t.BoundingBox.Draw()
	}
	gl.UseProgram(t.font.program)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, t.font.textureID)

	// uniforms
	gl.Uniform1i(t.font.fragmentTextureUniform, 0)
	gl.Uniform4fv(t.font.colorUniform, 1, &t.color[0])
	gl.Uniform2fv(t.font.finalPositionUniform, 1, &t.finalPosition[0])
	gl.UniformMatrix4fv(t.font.orthographicMatrixUniform, 1, false, &t.font.OrthographicMatrix[0])
	gl.UniformMatrix4fv(t.font.scaleMatrixUniform, 1, false, &t.scaleMatrix[0])

	// draw
	drawCount := int32(t.RuneCount * 6)
	if drawCount > int32(t.eboIndexCount) {
		drawCount = int32(t.eboIndexCount)
	}
	if drawCount < 0 {
		return
	}
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.BindVertexArray(t.vao)
	gl.DrawElements(gl.TRIANGLES, drawCount, gl.UNSIGNED_INT, nil)
	gl.BindVertexArray(0)
	gl.Disable(gl.BLEND)
}

// centerTheData prepares the value "centered_position" found in the font shader
// as named, the function centers the text around the orthographic center of the screen
func (t *Text) centerTheData(lowerLeft Point) (err error) {
	length := len(t.vboData)
	for index := 0; index < length; {
		// index (0,0)
		t.vboData[index] += lowerLeft.X
		index++
		t.vboData[index] += lowerLeft.Y
		index += 3 // skip texture data

		// index (1,0)
		t.vboData[index] += lowerLeft.X
		index++
		t.vboData[index] += lowerLeft.Y
		index += 3

		// index (1,1)
		t.vboData[index] += lowerLeft.X
		index++
		t.vboData[index] += lowerLeft.Y
		index += 3

		// index (0,1)
		t.vboData[index] += lowerLeft.X
		index++
		t.vboData[index] += lowerLeft.Y
		index += 3
	}

	// update bounding box
	t.X1.X += lowerLeft.X
	t.X2.X += lowerLeft.X
	t.X1.Y += lowerLeft.Y
	t.X2.Y += lowerLeft.Y

	// prepare objects for drawing the bounding box
	if IsDebug {
		t.BoundingBox, err = loadBoundingBox(t.font, t.X1, t.X2)
	}
	return
}

func (t *Text) Width() float32 {
	return t.X2.X - t.X1.X
}

func (t *Text) Height() float32 {
	return t.X2.Y - t.X1.Y
}

// PrintCharSpacing is used for debugging
func (t *Text) PrintCharSpacing() {
	fmt.Printf("\n%s:\n", t.String)
	at := t.X1.X
	for i, cs := range t.CharSpacing {
		at = cs + at
		fmt.Printf("'%c': %.2f ", t.String[i], at)
	}
}

// ClickedCharacter should only be called after a bounding box hit is confirmed because
// it does not check y-axis values at all.  Returns the index and side of the char clicked.
func (t *Text) ClickedCharacter(xPos float64) (index int, side CharacterSide) {
	// transform from screen coordinates to... window coordinates?
	xPos = xPos - float64(t.font.WindowWidth/2)

	// could do a binary search...
	at := float64(t.X1.X)
	for i, cs := range t.CharSpacing {
		at = float64(cs) + at
		if i == 0 && xPos <= at-float64(cs) {
			return i, CSLeft
		}
		if i == len(t.CharSpacing)-1 && xPos > at {
			return i, CSRight
		}
		if xPos <= at && xPos > at-float64(cs) {
			if xPos-(at-float64(cs)) > float64(cs)/2 {
				return i, CSRight
			} else {
				return i, CSLeft
			}
		}
	}
	return -1, CSUnknown
}

func (t *Text) CharPosition(index int) float64 {
	at := float64(t.X1.X)
	for i, cs := range t.CharSpacing {
		if i == index {
			break
		}
		at = float64(cs) + at
	}
	return at
}

func (t *Text) HasRune(r rune) bool {
	for _, runes := range t.font.Config.RuneRanges {
		if r >= runes.Low && r <= runes.High {
			return true
		}
	}
	return false
}

// makeBufferData positions quads for drawing the text in the indices parameter using glyph dimensions
func (t *Text) makeBufferData(indices []rune) {
	glyphs := t.font.Config.Glyphs

	vboIndex := 0
	eboIndex := 0
	lineX := float32(0)
	eboOffset := int32(0)

	t.CharSpacing = make([]float32, 0)
	for i, r := range indices {
		glyphIndex := t.font.Config.RuneRanges.GetGlyphIndex(r)
		if glyphIndex >= 0 {
			if IsDebug {
				prefix := DebugPrefix()
				fmt.Println("glyphIndex", glyphIndex)
				fmt.Printf("%s rune %+v line at %f", prefix, glyphs[glyphIndex], lineX)
			}
			advance := float32(glyphs[glyphIndex].Advance)
			vw := float32(glyphs[glyphIndex].Width)
			vh := float32(glyphs[glyphIndex].Height)

			// used to determine which character inside of the text was clicked
			//t.CharSpacing = append(t.CharSpacing, vw)
			t.CharSpacing = append(t.CharSpacing, advance)

			// variable width characters will produce a bounding box that is just
			// a bit too long on the right-hand side unless we trim off the excess
			// when processing the right-most character
			trim := float32(0)
			if i == len(indices)-1 {
				trim = vw - advance
			}
			tP1, tP2 := glyphs[glyphIndex].GetTexturePositions(t.font)

			// counter-clockwise quad
			// the bounding box value X2 is being expanded as characters are added

			// index (0,0)
			t.vboData[vboIndex] = lineX // position
			vboIndex++
			t.vboData[vboIndex] = 0
			vboIndex++
			t.vboData[vboIndex] = tP1.X // texture uv
			vboIndex++
			t.vboData[vboIndex] = tP2.Y
			vboIndex++

			// index (1,0) - expanding X2
			t.vboData[vboIndex], t.X2.X = lineX+vw, lineX+vw-trim
			vboIndex++
			t.vboData[vboIndex] = 0
			vboIndex++
			t.vboData[vboIndex] = tP2.X
			vboIndex++
			t.vboData[vboIndex] = tP2.Y
			vboIndex++

			// index (1,1) - expanding X2
			t.vboData[vboIndex] = lineX + vw
			vboIndex++
			t.vboData[vboIndex], t.X2.Y = vh, vh
			vboIndex++
			t.vboData[vboIndex] = tP2.X
			vboIndex++
			t.vboData[vboIndex] = tP1.Y
			vboIndex++

			// index (0,1)
			t.vboData[vboIndex] = lineX
			vboIndex++
			t.vboData[vboIndex] = vh
			vboIndex++
			t.vboData[vboIndex] = tP1.X
			vboIndex++
			t.vboData[vboIndex] = tP1.Y
			vboIndex++

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

			// shift to the right
			lineX += advance
			if IsDebug {
				fmt.Printf("-> %f\n", lineX)
			}
		}
	}
	return
}
