// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gltext

import (
	"bufio"
	"github.com/go-gl/glow/gl-core/3.3/gl"
	"github.com/go-gl/mathgl/mgl32"
	"image"
	"image/png"
	"os"
)

const debug = false

var vertexShaderSource string = `
#version 330

uniform mat4 scale_matrix;
uniform mat4 orthographic_matrix;

in vec4 position;
in vec2 uv;

out vec2 fragment_uv;

void main() {
  fragment_uv = uv;
  gl_Position = scale_matrix * orthographic_matrix * position;
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
	config         *FontConfig // Character set for this font.
	textureID      uint32      // Holds the glyph texture id.
	maxGlyphWidth  int         // Largest glyph width.
	maxGlyphHeight int         // Largest glyph height.
	program        uint32      // program compiled from shaders

	// attributes
	position uint32 // vertex position
	uv       uint32 // texture position

	// Position of the shaders fragment texture variable
	fragmentTextureUniform int32

	// The desired color of the text
	colorUniform int32

	// The background of the image is transparent.  Using an arbitrary
	// lower limit to distinguish between the background and the text.
	// There must be a better way that preserves the gradient-like
	// appearance of the text that the freetype-go library produces.
	textLowerBoundUniform int32
	textLowerBound        float32

	// View matrix
	orthographicMatrixUniform int32
	orthographicMatrix        mgl32.Mat4

	// Scale the resulting text
	scaleMatrixUniform int32

	textureWidth  float32
	textureHeight float32
	windowWidth   float32
	windowHeight  float32
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
	f.program, err = NewProgram(vertexShaderSource, fragmentShaderSource)
	if err != nil {
		return f, err
	}

	// attributes
	f.position = uint32(gl.GetAttribLocation(f.program, gl.Str("position\x00")))
	f.uv = uint32(gl.GetAttribLocation(f.program, gl.Str("uv\x00")))

	// uniforms
	f.orthographicMatrixUniform = gl.GetUniformLocation(f.program, gl.Str("orthographic_matrix\x00"))
	f.scaleMatrixUniform = gl.GetUniformLocation(f.program, gl.Str("scale_matrix\x00"))
	f.fragmentTextureUniform = gl.GetUniformLocation(f.program, gl.Str("fragment_texture\x00"))
	f.colorUniform = gl.GetUniformLocation(f.program, gl.Str("fragment_color_adjustment\x00"))
	f.textLowerBoundUniform = gl.GetUniformLocation(f.program, gl.Str("text_lowerbound\x00"))
	return f, nil
}

func (f *Font) ResizeWindow(width float32, height float32) {
	f.windowWidth = width
	f.windowHeight = height
	f.orthographicMatrix = mgl32.Ortho2D(0, f.windowWidth, 0, f.windowHeight)
}

// Release releases font resources.
// A font can no longer be used for rendering after this call completes.
func (f *Font) Release() {
	gl.DeleteTextures(1, &f.textureID)
	f.config = nil
}

func (f *Font) SetTextLowerBound(v float32) {
	f.textLowerBound = v
}
