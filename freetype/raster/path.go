// Copyright 2010 The Freetype-Go Authors. All rights reserved.
// Use of this source code is governed by your choice of either the
// FreeType License or the GNU General Public License version 2 (or
// any later version), both of which can be found in the LICENSE file.

package raster

import (
	"fmt"

	"github.com/eaburns/truefont/freetype/geom"
)

// An Adder accumulates points on a curve.
type Adder interface {
	// Start starts a new curve at the given point.
	Start(a geom.Point)
	// Add1 adds a linear segment to the current curve.
	Add1(b geom.Point)
	// Add2 adds a quadratic segment to the current curve.
	Add2(b, c geom.Point)
	// Add3 adds a cubic segment to the current curve.
	Add3(b, c, d geom.Point)
}

// A Path is a sequence of curves, and a curve is a start point followed by a
// sequence of linear, quadratic or cubic segments.
type Path []geom.Fix32

// String returns a human-readable representation of a Path.
func (p Path) String() string {
	s := ""
	for i := 0; i < len(p); {
		if i != 0 {
			s += " "
		}
		switch p[i] {
		case 0:
			s += "S0" + fmt.Sprint([]geom.Fix32(p[i+1:i+3]))
			i += 4
		case 1:
			s += "A1" + fmt.Sprint([]geom.Fix32(p[i+1:i+3]))
			i += 4
		case 2:
			s += "A2" + fmt.Sprint([]geom.Fix32(p[i+1:i+5]))
			i += 6
		case 3:
			s += "A3" + fmt.Sprint([]geom.Fix32(p[i+1:i+7]))
			i += 8
		default:
			panic("freetype/raster: bad path")
		}
	}
	return s
}

// grow adds n elements to p.
func (p *Path) grow(n int) {
	n += len(*p)
	if n > cap(*p) {
		old := *p
		*p = make([]geom.Fix32, n, 2*n+8)
		copy(*p, old)
		return
	}
	*p = (*p)[0:n]
}

// Clear cancels any previous calls to p.Start or p.AddXxx.
func (p *Path) Clear() {
	*p = (*p)[0:0]
}

// Start starts a new curve at the given point.
func (p *Path) Start(a geom.Point) {
	n := len(*p)
	p.grow(4)
	(*p)[n] = 0
	(*p)[n+1] = a.X
	(*p)[n+2] = a.Y
	(*p)[n+3] = 0
}

// Add1 adds a linear segment to the current curve.
func (p *Path) Add1(b geom.Point) {
	n := len(*p)
	p.grow(4)
	(*p)[n] = 1
	(*p)[n+1] = b.X
	(*p)[n+2] = b.Y
	(*p)[n+3] = 1
}

// Add2 adds a quadratic segment to the current curve.
func (p *Path) Add2(b, c geom.Point) {
	n := len(*p)
	p.grow(6)
	(*p)[n] = 2
	(*p)[n+1] = b.X
	(*p)[n+2] = b.Y
	(*p)[n+3] = c.X
	(*p)[n+4] = c.Y
	(*p)[n+5] = 2
}

// Add3 adds a cubic segment to the current curve.
func (p *Path) Add3(b, c, d geom.Point) {
	n := len(*p)
	p.grow(8)
	(*p)[n] = 3
	(*p)[n+1] = b.X
	(*p)[n+2] = b.Y
	(*p)[n+3] = c.X
	(*p)[n+4] = c.Y
	(*p)[n+5] = d.X
	(*p)[n+6] = d.Y
	(*p)[n+7] = 3
}

// AddPath adds the Path q to p.
func (p *Path) AddPath(q Path) {
	n, m := len(*p), len(q)
	p.grow(m)
	copy((*p)[n:n+m], q)
}

// AddStroke adds a stroked Path.
func (p *Path) AddStroke(q Path, width geom.Fix32, cr Capper, jr Joiner) {
	Stroke(p, q, width, cr, jr)
}

// firstgeom.Point returns the first point in a non-empty Path.
func (p Path) firstPoint() geom.Point {
	return geom.Pt(p[1], p[2])
}

// lastgeom.Point returns the last point in a non-empty Path.
func (p Path) lastPoint() geom.Point {
	return geom.Pt(p[len(p)-3], p[len(p)-2])
}

// addPathReversed adds q reversed to p.
// For example, if q consists of a linear segment from A to B followed by a
// quadratic segment from B to C to D, then the values of q looks like:
// index: 01234567890123
// value: 0AA01BB12CCDD2
// So, when adding q backwards to p, we want to Add2(C, B) followed by Add1(A).
func addPathReversed(p Adder, q Path) {
	if len(q) == 0 {
		return
	}
	i := len(q) - 1
	for {
		switch q[i] {
		case 0:
			return
		case 1:
			i -= 4
			p.Add1(geom.Pt(q[i-2], q[i-1]))
		case 2:
			i -= 6
			p.Add2(geom.Pt(q[i+2], q[i+3]), geom.Pt(q[i-2], q[i-1]))
		case 3:
			i -= 8
			p.Add3(geom.Pt(q[i+4], q[i+5]), geom.Pt(q[i+2], q[i+3]), geom.Pt(q[i-2], q[i-1]))
		default:
			panic("freetype/raster: bad path")
		}
	}
}
