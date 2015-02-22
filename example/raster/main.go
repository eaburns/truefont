// © 2015 The truefont Authors. See AUTHORS file for a list of authors.
//
// Copyright 2010 The Freetype-Go Authors. All rights reserved.
// Use of this source code is governed by your choice of either the
// FreeType License or the GNU General Public License version 2 (or
// any later version), both of which can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	"code.google.com/p/freetype-go/freetype/raster"
)

type node struct {
	x, y, degree int
}

// These contours "outside" and "inside" are from the `A' glyph from the Droid
// Serif Regular font.

var outside = []node{
	{414, 489, 1},
	{336, 274, 2},
	{327, 250, 0},
	{322, 226, 2},
	{317, 203, 0},
	{317, 186, 2},
	{317, 134, 0},
	{350, 110, 2},
	{384, 86, 0},
	{453, 86, 1},
	{500, 86, 1},
	{500, 0, 1},
	{0, 0, 1},
	{0, 86, 1},
	{39, 86, 2},
	{69, 86, 0},
	{90, 92, 2},
	{111, 99, 0},
	{128, 117, 2},
	{145, 135, 0},
	{160, 166, 2},
	{176, 197, 0},
	{195, 246, 1},
	{649, 1462, 1},
	{809, 1462, 1},
	{1272, 195, 2},
	{1284, 163, 0},
	{1296, 142, 2},
	{1309, 121, 0},
	{1326, 108, 2},
	{1343, 96, 0},
	{1365, 91, 2},
	{1387, 86, 0},
	{1417, 86, 1},
	{1444, 86, 1},
	{1444, 0, 1},
	{881, 0, 1},
	{881, 86, 1},
	{928, 86, 2},
	{1051, 86, 0},
	{1051, 184, 2},
	{1051, 201, 0},
	{1046, 219, 2},
	{1042, 237, 0},
	{1034, 260, 1},
	{952, 489, 1},
	{414, 489, -1},
}

var inside = []node{
	{686, 1274, 1},
	{453, 592, 1},
	{915, 592, 1},
	{686, 1274, -1},
}

func p(n node) raster.Point {
	x, y := 20+n.x/4, 380-n.y/4
	return raster.Point{
		X: raster.Fix32(x * 256),
		Y: raster.Fix32(y * 256),
	}
}

func contour(r *raster.Rasterizer, ns []node) {
	if len(ns) == 0 {
		return
	}
	i := 0
	r.Start(p(ns[i]))
	for {
		switch ns[i].degree {
		case -1:
			// -1 signifies end-of-contour.
			return
		case 1:
			i++
			r.Add1(p(ns[i]))
		case 2:
			i += 2
			r.Add2(p(ns[i-1]), p(ns[i]))
		default:
			panic("bad degree")
		}
	}
}

func showNodes(m *image.RGBA, ns []node) {
	for _, n := range ns {
		p := p(n)
		x, y := int(p.X)/256, int(p.Y)/256
		if !(image.Point{x, y}).In(m.Bounds()) {
			continue
		}
		var c color.Color
		switch n.degree {
		case 0:
			c = color.RGBA{0, 255, 255, 255}
		case 1:
			c = color.RGBA{255, 0, 0, 255}
		case 2:
			c = color.RGBA{255, 0, 0, 255}
		}
		if c != nil {
			m.Set(x, y, c)
		}
	}
}

func main() {
	// Rasterize the contours to a mask image.
	const (
		w = 400
		h = 400
	)
	r := raster.NewRasterizer(w, h)
	contour(r, outside)
	contour(r, inside)
	mask := image.NewAlpha(image.Rect(0, 0, w, h))
	p := raster.NewAlphaSrcPainter(mask)
	r.Rasterize(p)

	// Draw the mask image (in gray) onto an RGBA image.
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	gray := image.NewUniform(color.Alpha{0x1f})
	draw.Draw(rgba, rgba.Bounds(), image.Black, image.ZP, draw.Src)
	draw.DrawMask(rgba, rgba.Bounds(), gray, image.ZP, mask, image.ZP, draw.Over)
	showNodes(rgba, outside)
	showNodes(rgba, inside)

	// Save that RGBA image to disk.
	f, err := os.Create("out.png")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	b := bufio.NewWriter(f)
	err = png.Encode(b, rgba)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = b.Flush()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	fmt.Println("Wrote out.png OK.")
}
