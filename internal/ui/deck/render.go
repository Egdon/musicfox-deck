package deck

import (
	"math"
	"runtime"
	"sync"
)

// The rest pose turns the cassette slightly to show its top and left walls.
const (
	restYaw   = -14 * math.Pi / 180
	restPitch = 9 * math.Pi / 180
)

// camDist is the camera's distance from the cassette in face units, far
// enough for a gentle perspective.
const camDist = 2.8 * faceW

// Pose places the cassette in front of the camera.
type Pose struct {
	Yaw, Pitch, Roll float64 // radians
	X, Y             float64 // offset from the center of the fit box, face units
	Scale            float64 // 0 counts as 1
}

// RestPose is the cassette lying still with its A side up.
func RestPose() Pose { return Pose{Yaw: restYaw, Pitch: restPitch, Scale: 1} }

// State is what moves inside the cassette.
type State struct {
	Radii  [2]float64 // tape pack radius of each reel, see ReelRadii
	Angles [2]float64 // reel angles in object space
	Blur   float64    // 0 shows sharp teeth, 1 a blurred ring
	Lift   float64    // height above the backdrop in face units: the shadow falls farther and softer
}

// Rect is a rectangle in cells.
type Rect struct{ X, Y, W, H float64 }

// View is what to draw into: a W×H cell area, the cell split and shape, and
// the box the cassette is fitted to (the whole area when empty).
type View struct {
	W, H   int
	Tier   Tier
	Aspect float64 // cell height over width; 0 means 2
	Fit    Rect
}

// Frame is a rendered view in row-major cells. A cell with an empty Ch is
// not drawn, which leaves the caller's background.
type Frame struct {
	W, H  int
	Cells []Cell
}

// Renderer draws cassettes. It keeps its buffers between frames and returns
// the previous frame when nothing changed; a Frame is valid until the next
// call.
type Renderer struct {
	key   renderKey
	valid bool
	mats  materials
	pal   Palette
	frame Frame

	col  []rgb     // object color per sub-pixel
	cov  []float64 // object coverage per sub-pixel
	mat  []mat
	edge []bool
	shad []float64
	tmp  []float64
}

type renderKey struct {
	v    View
	pal  Palette
	pose Pose
	st   State
}

// camera holds the projection of one frame.
type camera struct {
	aspect float64
	cx, cy float64 // principal point in physical units (cell widths)
	focal  float64 // physical units per unit of x/z
	k      float64 // physical units per face unit at the cassette
}

func newCamera(v View) camera {
	aspect := v.Aspect
	if aspect <= 0 {
		aspect = 2
	}
	fit := v.Fit
	if fit.W <= 0 || fit.H <= 0 {
		fit = Rect{0, 0, float64(v.W), float64(v.H)}
	}
	// Leave room for the turned walls, the float and the shadow.
	k := min(0.9*fit.W/(faceW+12), 0.9*fit.H*aspect/(faceH+12))
	return camera{
		aspect: aspect,
		cx:     fit.X + fit.W/2,
		cy:     (fit.Y + fit.H/2) * aspect,
		focal:  k * camDist,
		k:      k,
	}
}

// project maps a camera-space point to physical screen units.
func (c camera) project(p vec3) (float64, float64) {
	return c.cx + c.focal*p.x/p.z, c.cy + c.focal*p.y/p.z
}

// Bounds returns the cells the cassette covers at a pose, without its
// shadow, as x0, y0, x1, y1.
func Bounds(v View, p Pose) (x0, y0, x1, y1 float64) {
	cam := newCamera(v)
	rot, s := rotation(p.Yaw, p.Pitch, p.Roll), scaleOf(p)
	t := vec3{p.X, p.Y, camDist}
	x0, y0, x1, y1 = math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for i := range 8 {
		c := vec3{faceW / 2, faceH / 2, depth / 2}
		if i&1 != 0 {
			c.x = -c.x
		}
		if i&2 != 0 {
			c.y = -c.y
		}
		if i&4 != 0 {
			c.z = -c.z
		}
		px, py := cam.project(rot.apply(c.mul(s)).add(t))
		x0, x1 = min(x0, px), max(x1, px)
		y0, y1 = min(y0, py/cam.aspect), max(y1, py/cam.aspect)
	}
	return x0, y0, x1, y1
}

func scaleOf(p Pose) float64 {
	if p.Scale <= 0 {
		return 1
	}
	return p.Scale
}

// Render draws the cassette.
func (r *Renderer) Render(v View, pal Palette, pose Pose, st State) *Frame {
	key := renderKey{v, pal, pose, st}
	if r.valid && key == r.key {
		return &r.frame
	}
	if !r.valid || pal != r.pal {
		r.mats, r.pal = materialsFor(pal), pal
	}
	r.key, r.valid = key, true

	n := v.W * v.H
	if cap(r.frame.Cells) < n {
		r.frame.Cells = make([]Cell, n)
	}
	r.frame.W, r.frame.H, r.frame.Cells = v.W, v.H, r.frame.Cells[:n]
	clear(r.frame.Cells)
	if v.W <= 0 || v.H <= 0 {
		return &r.frame
	}

	cam := newCamera(v)
	rot, s := rotation(pose.Yaw, pose.Pitch, pose.Roll), scaleOf(pose)
	t := vec3{pose.X, pose.Y, camDist}
	sc := &scene{
		m:      &r.mats,
		rot:    rot,
		scale:  s,
		o:      rot.applyT(t.mul(-1)).mul(1 / s),
		radii:  st.Radii,
		angles: st.Angles,
		blur:   st.Blur,
		lObj:   rot.applyT(keyLight),
	}

	gc, gr := v.Tier.Grid()
	sc.detail = cam.k * s * float64(gr) / cam.aspect
	// The shadow falls down and to the right, farther and softer the
	// higher the cassette floats.
	lift := max(st.Lift, 0)
	scale := cam.k * s
	offX, offY := (0.9+0.35*lift)*scale, (1.2+0.45*lift)*scale // physical units
	blurR := (0.8 + 0.3*lift) * scale

	bx0, by0, bx1, by1 := Bounds(v, pose)
	x0 := max(int(math.Floor(bx0))-1, 0)
	y0 := max(int(math.Floor(by0))-1, 0)
	x1 := min(int(math.Ceil(bx1+offX+2*blurR))+1, v.W)
	y1 := min(int(math.Ceil(by1+(offY+2*blurR)/cam.aspect))+1, v.H)
	if x0 >= x1 || y0 >= y1 {
		return &r.frame
	}
	bw, bh := x1-x0, y1-y0
	sw, sh := bw*gc, bh*gr
	r.grow(sw * sh)

	// Pass 1: one ray through the center of every sub-pixel.
	ray := func(fx, fy float64) vec3 {
		px := float64(x0) + fx/float64(gc)
		py := (float64(y0) + fy/float64(gr)) * cam.aspect
		return vec3{(px - cam.cx) / cam.focal, (py - cam.cy) / cam.focal, 1}
	}
	parallel(sh, func(j0, j1 int) {
		for j := j0; j < j1; j++ {
			for i := range sw {
				k := j*sw + i
				c, m, ok := sc.sample(ray(float64(i)+0.5, float64(j)+0.5))
				r.col[k], r.mat[k], r.cov[k] = c, m, 0
				if ok {
					r.cov[k] = 1
				}
			}
		}
	})
	// Pass 2: four more rays where neighboring materials differ.
	edge := r.edge
	for j := range sh {
		for i := range sw {
			k, m := j*sw+i, r.mat[j*sw+i]
			edge[k] = (i > 0 && r.mat[k-1] != m) || (i+1 < sw && r.mat[k+1] != m) ||
				(j > 0 && r.mat[k-sw] != m) || (j+1 < sh && r.mat[k+sw] != m)
		}
	}
	parallel(sh, func(j0, j1 int) {
		for j := j0; j < j1; j++ {
			for i := range sw {
				k := j*sw + i
				if !edge[k] {
					continue
				}
				var sum rgb
				hits := 0
				for _, o := range [4][2]float64{{0.25, 0.25}, {0.75, 0.25}, {0.25, 0.75}, {0.75, 0.75}} {
					if c, _, ok := sc.sample(ray(float64(i)+o[0], float64(j)+o[1])); ok {
						sum, hits = sum.add(c), hits+1
					}
				}
				r.cov[k] = float64(hits) / 4
				if hits > 0 {
					r.col[k] = sum.scale(1 / float64(hits))
				}
			}
		}
	})

	// Pass 3: the shadow is the coverage, shifted and blurred.
	dx, dy := int(math.Round(offX*float64(gc))), int(math.Round(offY*float64(gr)/cam.aspect))
	rx := max(int(math.Round(blurR*float64(gc))), 1)
	ry := max(int(math.Round(blurR*float64(gr)/cam.aspect)), 1)
	for j := range sh {
		for i := range sw {
			c := 0.0
			if si, sj := i-dx, j-dy; si >= 0 && sj >= 0 {
				c = r.cov[sj*sw+si]
			}
			r.shad[j*sw+i] = c
		}
	}
	for range 2 {
		boxBlur(r.shad, r.tmp, sw, sh, rx, 1, 0)
		boxBlur(r.tmp, r.shad, sw, sh, ry, 0, 1)
	}

	// Pass 4: compose and fit each cell.
	bg := r.mats.bg
	parallel(bh, func(c0, c1 int) {
		px := make([]rgb, gc*gr)
		for cy := c0; cy < c1; cy++ {
			for cx := range bw {
				blank := true
				for q := range gc * gr {
					k := (cy*gr+q/gc)*sw + cx*gc + q%gc
					cov, shadow := r.cov[k], r.shad[k]
					under := bg.scale(1 - 0.45*shadow)
					px[q] = r.col[k].scale(cov).add(under.scale(1 - cov))
					blank = blank && cov == 0 && shadow < 0.01
				}
				if blank {
					continue
				}
				r.frame.Cells[(y0+cy)*v.W+x0+cx] = fitCell(px, v.Tier)
			}
		}
	})
	return &r.frame
}

func (r *Renderer) grow(n int) {
	if cap(r.col) < n {
		r.col, r.cov, r.mat, r.edge = make([]rgb, n), make([]float64, n), make([]mat, n), make([]bool, n)
		r.shad, r.tmp = make([]float64, n), make([]float64, n)
	}
	r.col, r.cov, r.mat, r.edge = r.col[:n], r.cov[:n], r.mat[:n], r.edge[:n]
	r.shad, r.tmp = r.shad[:n], r.tmp[:n]
}

// boxBlur averages src over a window of 2r+1 samples along (dx, dy) into dst.
func boxBlur(src, dst []float64, w, h, r, dx, dy int) {
	for j := range h {
		for i := range w {
			var sum float64
			for k := -r; k <= r; k++ {
				x, y := i+k*dx, j+k*dy
				if x >= 0 && y >= 0 && x < w && y < h {
					sum += src[y*w+x]
				}
			}
			dst[j*w+i] = sum / float64(2*r+1)
		}
	}
}

var workers = min(runtime.GOMAXPROCS(0), 4)

// parallel splits 0..n into bands and runs f on them concurrently.
func parallel(n int, f func(lo, hi int)) {
	if workers <= 1 || n < 8 {
		f(0, n)
		return
	}
	var wg sync.WaitGroup
	step := (n + workers - 1) / workers
	for lo := 0; lo < n; lo += step {
		wg.Add(1)
		go func(lo, hi int) {
			defer wg.Done()
			f(lo, hi)
		}(lo, min(lo+step, n))
	}
	wg.Wait()
}
