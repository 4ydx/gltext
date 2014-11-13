## Modern opengl text rendering 

A simple package for rendering a string using modern opengl.  Based on the bounding
box of a string, positioning of the string on screen prior to rendering is possible.
There do seem to be issues with the dimensions reported by freetype-go unfortunately.

- Unicode support.
- Dynamic text zooming along the z-axis.
- Dynamic text positioning within the orthographic projection space.
- Dynamic color changes.

Unicode support is based on the underlying truetype font being used (or bitmap).

![Alt text](/example.gif?raw=true "Simple Screenshot")

### Install

* go get github.com/4ydx/gltext

### Example

* Refer to https://github.com/4ydx/gltext_example as an example.

### Dependencies

This packages uses [freetype-go](https://code.google.com/p/freetype-go) which is licensed 
under GPLv2 and FTL licenses. You can choose which one is a better fit for your 
use case but FTL requires you to give some form of credit to Freetype.org

You can read the [GPLv2](https://code.google.com/p/freetype-go/source/browse/licenses/gpl.txt)
and [FTL](https://code.google.com/p/freetype-go/source/browse/licenses/ftl.txt)
licenses for more information about the requirements.

### License

Copyright 2012 The go-gl Authors. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.

