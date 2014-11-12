// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gltext

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
)

// Direction represents the direction in which strings should be rendered.
type Direction uint8

// TODO: Known directions.
const (
	LeftToRight Direction = iota // E.g.: Latin
)

// FontConfig describes raster font metadata.
//
// It can be loaded from, or saved to a JSON encoded file,
// which should come with any bitmap font image.
type FontConfig struct {
	// The direction determines the orientation of rendered strings and should
	// hold any of the pre-defined Direction constants.
	Dir Direction `json:"direction"`

	// Lower rune boundary
	Low rune `json:"rune_low"`

	// Upper rune boundary.
	High rune `json:"rune_high"`

	// Glyphs holds a set of glyph descriptors, defining the location,
	// size and advance of each glyph in the sprite sheet.
	Glyphs Charset `json:"glyphs"`
}

// Load reads font configuration data from the given JSON encoded stream.
func (fc *FontConfig) Load(r io.Reader) (err error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return
	}
	return json.Unmarshal(data, fc)
}

// Save writes font configuration data to the given stream as JSON data.
func (fc *FontConfig) Save(w io.Writer) (err error) {
	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return
	}
	_, err = io.Copy(w, bytes.NewBuffer(data))
	return
}
