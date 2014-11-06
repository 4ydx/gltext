## a 'modern' opengl rewrite of go-gl/gltext

A simple package for rendering a string using modern opengl.  Based on the bounding
box of a string, positioning of the string on screen prior to rendering is possible.
There do seem to be issues with the dimensions reported by freetype-go unfortunately.

Unicode support is based on the underlying truetype font being used (or bitmap).

![Alt text](/example.png?raw=true "Working Example")

### TODO

* Have a look at Valve's 'Signed Distance Field` techniques to render
  sharp font textures are different zoom levels.

  * [SIGGRAPH2007_AlphaTestedMagnification.pdf](http://www.valvesoftware.com/publications/2007/SIGGRAPH2007_AlphaTestedMagnification.pdf)
  * [Youtube video](http://www.youtube.com/watch?v=CGZRHJvJYIg)
  
  More links to info in the youtube video description.
  An alternative might be a port of [GLyphy](http://code.google.com/p/glyphy/)


### Known bugs

* Determining the height of truetype glyphs is not entirely accurate.
  It is unclear at this point how to get to this information reliably.
  Specifically the parts in `LoadTruetype` at truetype.go#L54+.
  The vertical glyph bounds computed by freetype-go are not correct for
  certain fonts. Right now we manually offset the value by added `4` to
  the height. This is an unreliable hack and should be fixed.
* `freetype-go` does not expose `AdvanceHeight` for vertically rendered fonts.
  This may mean that the Advance size for top-to-bottom fonts is incorrect.


### Dependencies

This packages uses [freetype-go](https://code.google.com/p/freetype-go) which is licensed 
under GPLv2 e FTL licenses. You can choose which one is a better fit for your 
use case but FTL requires you to give some form of credit to Freetype.org

You can read the [GPLv2](https://code.google.com/p/freetype-go/source/browse/licenses/gpl.txt)
and [FTL](https://code.google.com/p/freetype-go/source/browse/licenses/ftl.txt)
licenses for more information about the requirements.

### Usage

    go get github.com/4ydx/gltext

Refer to [4ydx/test_gltext][ex] for usage examples.

[ex]: https://github.com/4ydx/test_gltext


### License

Copyright 2012 The go-gl Authors. All rights reserved.
Use of this source code is governed by a BSD-style
license that can be found in the LICENSE file.

