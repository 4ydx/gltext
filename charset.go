// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gltext

// A Glyph describes metrics for a single font glyph.
// These indicate which area of a given image contains the
// glyph data and how the glyph should be spaced in a rendered string.
type Point struct {
	X float32
	Y float32
}

type Glyph struct {
	X      int `json:"x"`      // The x location of the glyph on a sprite sheet.
	Y      int `json:"y"`      // The y location of the glyph on a sprite sheet.
	Width  int `json:"width"`  // The width of the glyph on a sprite sheet.
	Height int `json:"height"` // The height of the glyph on a sprite sheet.

	// Advance determines the distance to the next glyph.
	// This is used to properly align non-monospaced fonts.
	Advance int `json:"advance"`
}

func (g *Glyph) GetIndices(font *Font) (tP1, tP2 Point) {
	// Quad width/height
	vw := float32(g.Width)
	vh := float32(g.Height)

	// texture point 1
	tP1 = Point{X: float32(g.X) / font.textureWidth, Y: float32(g.Y) / font.textureHeight}

	// texture point 2
	tP2 = Point{X: (float32(g.X) + vw) / font.textureWidth, Y: (float32(g.Y) + vh) / font.textureHeight}

	return
}

// A Charset represents a set of glyph descriptors for a font.
// Each glyph descriptor holds glyph metrics which are used to
// properly align the given glyph in the resulting rendered string.
type Charset []Glyph

// Scale scales all glyphs by the given factor and repositions them
// appropriately. A scale of 1 retains the original size. A scale of 2
// doubles the size of each glyph, etc.
//
// This is useful when the accompanying sprite sheet is scaled by the
// same factor. In this case, we want the glyph data to match up with the
// new image.
func (c Charset) Scale(factor int) {
	if factor <= 1 {
		// A factor of zero results in zero-sized glyphs and
		// is therefore not valid. A factor of 1 does not change
		// the glyphs, so we can ignore it.
		return
	}

	// Multiply each glyph field by the given factor
	// to scale them up to the new size.
	for i := range c {
		c[i].X *= factor
		c[i].Y *= factor
		c[i].Width *= factor
		c[i].Height *= factor
		c[i].Advance *= factor
	}
}
