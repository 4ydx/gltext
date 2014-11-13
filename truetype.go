// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gltext

import (
	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"
	"image"
	"image/draw"
	"io"
	"io/ioutil"
)

// http://www.freetype.org/freetype2/docs/tutorial/step2.html

// LoadTruetype loads a truetype font from the given stream and
// applies the given font scale in points.
//
// The low and high values determine the lower and upper rune limits
// we should load for this font. For standard ASCII this would be: 32, 127.
func LoadTruetype(r io.Reader, scale int32, low, high rune) (*Font, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Read the truetype font.
	ttf, err := truetype.Parse(data)
	if err != nil {
		return nil, err
	}

	// Create our FontConfig type.
	var fc FontConfig
	fc.Low = low
	fc.High = high
	fc.Glyphs = make(Charset, high-low+1)

	// Create an image, large enough to store all requested glyphs.
	//
	// We limit the image to 16 glyphs per row. Then add as many rows as
	// needed to encompass all glyphs, while making sure the resulting image
	// has power-of-two dimensions.
	gc := int32(len(fc.Glyphs))
	glyphsPerRow := int32(16)
	glyphsPerCol := (gc / glyphsPerRow) + 1

	gb := ttf.Bounds(scale)
	gw := (gb.XMax - gb.XMin)
	//fmt.Println("max gb x", gb.XMax, "min gb x", gb.XMin)

	// why?
	gh := (gb.YMax - gb.YMin) + 5
	//fmt.Println("max gb y", gb.YMax, "min gb y", gb.YMin)

	iw := Pow2(uint32(gw * glyphsPerRow))
	ih := Pow2(uint32(gh * glyphsPerCol))

	fg, bg := image.White, image.Transparent
	rect := image.Rect(0, 0, int(iw), int(ih))
	img := image.NewRGBA(rect)
	draw.Draw(img, img.Bounds(), bg, image.ZP, draw.Src)

	// Use a freetype context to do the drawing.
	c := freetype.NewContext()
	c.SetDPI(72) // Do not change this.  It is required in order to have a properly aligned bounding box!!!
	c.SetFont(ttf)
	c.SetFontSize(float64(scale))
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(fg)

	// Iterate over all relevant glyphs in the truetype font and
	// draw them all to the image buffer.
	//
	// For each glyph, we also create a corresponding Glyph structure
	// for our Charset. It contains the appropriate glyph coordinate offsets.
	var gi int
	var gx, gy int32

	for ch := low; ch <= high; ch++ {
		index := ttf.Index(ch)
		metric := ttf.HMetric(scale, index)

		fc.Glyphs[gi].Advance = int(metric.AdvanceWidth)
		fc.Glyphs[gi].X = int(gx)
		fc.Glyphs[gi].Y = int(gy)
		fc.Glyphs[gi].Width = int(gw)
		fc.Glyphs[gi].Height = int(gh)

		pt := freetype.Pt(int(gx), int(gy)+int(c.PointToFix32(float64(scale))>>8))
		c.DrawString(string(ch), pt)

		if gi%16 == 0 {
			gx = 0
			gy += gh
		} else {
			gx += gw
		}
		gi++
	}
	return loadFont(img, &fc)
}
