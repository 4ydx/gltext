// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gltext

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io/ioutil"
	"os"
	"time"
)

// Direction represents the direction in which strings should be rendered.
type Direction uint8

// FontConfig describes raster font metadata.
//
// It can be loaded from, or saved to a JSON encoded file,
// which should come with any bitmap font image.
type FontConfig struct {
	// The range of glyphs covered by this fontconfig
	Low, High rune

	// Glyphs holds a set of glyph descriptors, defining the location,
	// size and advance of each glyph in the sprite sheet.
	Glyphs Charset

	Image *image.NRGBA `json:"-"`
}

// Load reads font configuration data from the given JSON encoded stream.
func (fc *FontConfig) Load(rootPath string) (err error) {
	file := fmt.Sprintf("%s/font.config", rootPath)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, fc)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", time.Now())
	fc.Image, err = LoadImage(rootPath)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", time.Now())
	fc.Glyphs.Scale(1)
	return nil
}

// Save writes font configuration data to the given stream as JSON data.
func (fc *FontConfig) Save(rootPath string) (err error) {
	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return
	}
	file := fmt.Sprintf("%s/font.config", rootPath)
	err = ioutil.WriteFile(file, data, 0600)
	if err != nil {
		return
	}
	err = SaveImage(rootPath, fc.Image)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(file, data, 0600)
	return
}

func LoadImage(rootPath string) (*image.NRGBA, error) {
	file := fmt.Sprintf("%s/image.png", rootPath)
	img, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	pix, _, err := image.Decode(img)
	if err != nil {
		return nil, err
	}
	p, ok := pix.(*image.NRGBA)
	if ok {
		return p, nil
	}
	return nil, errors.New("Not a NRGBA image.")
}

func SaveImage(rootPath string, img *image.NRGBA) error {
	file := fmt.Sprintf("%s/image.png", rootPath)
	image, err := os.Create(file)
	if err != nil {
		return err
	}
	defer image.Close()

	b := bufio.NewWriter(image)
	err = png.Encode(b, img)
	if err != nil {
		return err
	}
	err = b.Flush()
	if err != nil {
		return err
	}
	return nil
}
