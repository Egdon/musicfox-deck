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

// shadowDepth is how far the cassette's shadow darkens the backdrop.
const shadowDepth = 0.45

// fitMargin is the room left around the face in the fit box for the turned
// walls, the float and the shadow, in face units.
const fitMargin = 12

// FitAspect is the width over height of the box the cassette fills snugly,
// both measured in the same unit: a box rows tall holds it in
// rows × cell aspect × FitAspect columns.
const FitAspect = (faceW + fitMargin) / (faceH + fitMargin)

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
// the previous frame when nothing changed, or redraws only the hubs when just
// the reels turned; a Frame is valid until the next call.
type Renderer struct {
	key   renderKey
	valid bool
	mats  materials
	pal   Palette
	frame Frame
	pass  pass

	col  []rgb     // object color per sub-pixel
	cov  []float64 // object coverage per sub-pixel
	mat  []mat
	edge []bool
	hub  []bool // a sample of the sub-pixel landed on a hub, which turns
	mark []uint8
	shad []float64
	tmp  []float64
}

// pass is where the last full render drew: its scene and its box, in cells
// and in sub-pixels.
type pass struct {
	sc     scene
	cam    camera
	x0, y0 int
	bw, bh int
	gc, gr int
	sw, sh int
	bounds Rect // the cells the cassette covers: rays outside miss it
}

// ray is the camera-space direction through (fx, fy) in sub-pixels of the box.
func (p *pass) ray(fx, fy float64) vec3 {
	px := float64(p.x0) + fx/float64(p.gc)
	py := (float64(p.y0) + fy/float64(p.gr)) * p.cam.aspect
	return vec3{(px - p.cam.cx) / p.cam.focal, (py - p.cam.cy) / p.cam.focal, 1}
}

// subSamples are where an edge sub-pixel is sampled again.
var subSamples = [4][2]float64{{0.25, 0.25}, {0.75, 0.25}, {0.25, 0.75}, {0.75, 0.75}}

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
	k := min(0.9*fit.W/(faceW+fitMargin), 0.9*fit.H*aspect/(faceH+fitMargin))
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
	if r.valid && key.v == r.key.v && key.pal == r.key.pal && key.pose == r.key.pose &&
		key.st.Radii == r.key.st.Radii && key.st.Lift == r.key.st.Lift {
		r.key = key
		if r.turnReels(v, st) {
			return &r.frame
		}
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
	r.pass = pass{}
	if v.W <= 0 || v.H <= 0 {
		return &r.frame
	}

	cam := newCamera(v)
	rot, s := rotation(pose.Yaw, pose.Pitch, pose.Roll), scaleOf(pose)
	t := vec3{pose.X, pose.Y, camDist}
	sc := scene{
		m:      &r.mats,
		rot:    rot,
		inv:    1 / s,
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
	r.pass = pass{sc: sc, cam: cam, x0: x0, y0: y0, bw: bw, bh: bh, gc: gc, gr: gr, sw: sw, sh: sh,
		bounds: Rect{bx0, by0, bx1 - bx0, by1 - by0}}
	p := &r.pass

	// Pass 1: one ray through the center of every sub-pixel.
	parallel(sh, func(j0, j1 int) {
		for j := j0; j < j1; j++ {
			for i := range sw {
				k := j*sw + i
				r.center(p, i, j, k)
			}
		}
	})
	// Pass 2: four more rays where neighboring materials differ.
	for j := range sh {
		for i := range sw {
			r.edge[j*sw+i] = r.isEdge(p, i, j)
		}
	}
	parallel(sh, func(j0, j1 int) {
		for j := j0; j < j1; j++ {
			for i := range sw {
				if k := j*sw + i; r.edge[k] {
					r.supersample(p, i, j, k)
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
	parallel(bh, func(c0, c1 int) {
		px := make([]rgb, gc*gr)
		for cy := c0; cy < c1; cy++ {
			for cx := range bw {
				r.compose(p, v, cx, cy, px)
			}
		}
	})
	return &r.frame
}

// center shades the ray through the center of sub-pixel (i, j), at index k.
func (r *Renderer) center(p *pass, i, j, k int) {
	fx, fy := float64(i)+0.5, float64(j)+0.5
	if !p.mayHit(fx, fy) {
		r.col[k], r.mat[k], r.cov[k], r.hub[k] = r.mats.bg, matBg, 0, false
		return
	}
	c, m, ok := p.sc.sample(p.ray(fx, fy))
	r.col[k], r.mat[k], r.cov[k], r.hub[k] = c, m, 0, m.turns()
	if ok {
		r.cov[k] = 1
	}
}

// mayHit reports whether the ray through (fx, fy) passes within the cells the
// cassette covers; the shadow's margin lies beyond them.
func (p *pass) mayHit(fx, fy float64) bool {
	const e = 1e-6
	x := float64(p.x0) + fx/float64(p.gc)
	y := float64(p.y0) + fy/float64(p.gr)
	b := p.bounds
	return x >= b.X-e && x <= b.X+b.W+e && y >= b.Y-e && y <= b.Y+b.H+e
}

// isEdge reports whether sub-pixel (i, j) borders another material.
func (r *Renderer) isEdge(p *pass, i, j int) bool {
	k := j*p.sw + i
	m := r.mat[k]
	return (i > 0 && r.mat[k-1] != m) || (i+1 < p.sw && r.mat[k+1] != m) ||
		(j > 0 && r.mat[k-p.sw] != m) || (j+1 < p.sh && r.mat[k+p.sw] != m)
}

// supersample replaces the center sample of sub-pixel (i, j) with four.
func (r *Renderer) supersample(p *pass, i, j, k int) {
	var sum rgb
	hits := 0
	for _, o := range subSamples {
		c, m, ok := p.sc.sample(p.ray(float64(i)+o[0], float64(j)+o[1]))
		if ok {
			sum, hits = sum.add(c), hits+1
		}
		r.hub[k] = r.hub[k] || m.turns()
	}
	r.cov[k] = float64(hits) / 4
	if hits > 0 {
		r.col[k] = sum.scale(1 / float64(hits))
	}
}

// compose fits cell (cx, cy) of the box from its sub-pixels over the shadowed
// backdrop; px is scratch space for one cell.
func (r *Renderer) compose(p *pass, v View, cx, cy int, px []rgb) {
	bg := r.mats.bg
	covered, dark := false, 0.0
	for q := range p.gc * p.gr {
		k := (cy*p.gr+q/p.gc)*p.sw + cx*p.gc + q%p.gc
		cov, shadow := r.cov[k], r.shad[k]
		under := bg.scale(1 - shadowDepth*shadow)
		px[q] = r.col[k].scale(cov).add(under.scale(1 - cov))
		covered = covered || cov > 0
		dark += shadow
	}
	dark /= float64(p.gc * p.gr)
	cell := &r.frame.Cells[(p.y0+cy)*v.W+p.x0+cx]
	if !covered && dark < 0.01 {
		*cell = Cell{}
		return
	}
	c := fitCell(px, v.Tier)
	if !covered {
		c.Shadow = shadowDepth * dark
	}
	*cell = c
}

// Marks turnReels keeps per sub-pixel.
const (
	markHub  uint8 = 1 << iota // on a hub before the reels turned
	markNear                   // on a hub or next to one: may change
)

// turnReels redraws the frame for new reel angles and blur, which change
// only the hubs: the sub-pixels with a sample on one and, as their materials
// may change, the edges next to them. It reports false when the coverage
// changed after all, which moves the shadow and needs a full render.
func (r *Renderer) turnReels(v View, st State) bool {
	p := &r.pass
	if p.sw == 0 {
		return true // nothing is drawn
	}
	p.sc.m, p.sc.angles, p.sc.blur = &r.mats, st.Angles, st.Blur
	sw, sh := p.sw, p.sh
	i0, j0, i1, j1 := sw, sh, -1, -1
	for k, h := range r.hub {
		if h {
			i, j := k%sw, k/sw
			i0, j0, i1, j1 = min(i0, i), min(j0, j), max(i1, i), max(j1, j)
		}
	}
	if i1 < 0 {
		return true // no hub in sight
	}
	i0, j0, i1, j1 = max(i0-1, 0), max(j0-1, 0), min(i1+1, sw-1), min(j1+1, sh-1)
	if len(r.mark) < sw*sh {
		r.mark = make([]uint8, sw*sh)
	}
	mark := r.mark
	for j := j0; j <= j1; j++ {
		for i := i0; i <= i1; i++ {
			k := j*sw + i
			mark[k] = 0
			if r.hub[k] {
				mark[k] = markHub
			}
		}
	}
	for j := j0; j <= j1; j++ {
		for i := i0; i <= i1; i++ {
			k := j*sw + i
			if r.hub[k] || (i > 0 && r.hub[k-1]) || (i+1 < sw && r.hub[k+1]) ||
				(j > 0 && r.hub[k-sw]) || (j+1 < sh && r.hub[k+sw]) {
				mark[k] |= markNear
			}
		}
	}
	// The hubs' own center samples first: the edges depend on them.
	for j := j0; j <= j1; j++ {
		for i := i0; i <= i1; i++ {
			if k := j*sw + i; mark[k]&markNear != 0 {
				r.tmp[k] = r.cov[k]
				if mark[k]&markHub != 0 {
					r.center(p, i, j, k)
				}
			}
		}
	}
	for j := j0; j <= j1; j++ {
		for i := i0; i <= i1; i++ {
			k := j*sw + i
			if mark[k]&markNear == 0 {
				continue
			}
			was, edge := r.edge[k], r.isEdge(p, i, j)
			r.edge[k] = edge
			switch {
			case mark[k]&markHub != 0:
				if edge {
					r.supersample(p, i, j, k)
				}
			case edge && !was:
				r.supersample(p, i, j, k)
			case was && !edge:
				r.center(p, i, j, k)
			}
			if r.cov[k] != r.tmp[k] {
				return false
			}
		}
	}
	px := make([]rgb, p.gc*p.gr)
	for cy := j0 / p.gr; cy <= j1/p.gr; cy++ {
		for cx := i0 / p.gc; cx <= i1/p.gc; cx++ {
			if r.near(p, cx, cy) {
				r.compose(p, v, cx, cy, px)
			}
		}
	}
	return true
}

// near reports whether a sub-pixel of cell (cx, cy) is marked markNear.
func (r *Renderer) near(p *pass, cx, cy int) bool {
	for q := range p.gc * p.gr {
		if r.mark[(cy*p.gr+q/p.gc)*p.sw+cx*p.gc+q%p.gc]&markNear != 0 {
			return true
		}
	}
	return false
}

func (r *Renderer) grow(n int) {
	if cap(r.col) < n {
		r.col, r.cov, r.mat, r.edge = make([]rgb, n), make([]float64, n), make([]mat, n), make([]bool, n)
		r.hub, r.shad, r.tmp = make([]bool, n), make([]float64, n), make([]float64, n)
	}
	r.col, r.cov, r.mat, r.edge = r.col[:n], r.cov[:n], r.mat[:n], r.edge[:n]
	r.hub, r.shad, r.tmp = r.hub[:n], r.shad[:n], r.tmp[:n]
}

// boxBlur averages src over a window of 2r+1 samples along (dx, dy) into dst,
// counting samples beyond the edges as zero. The window slides: each step
// adds the sample it reaches and drops the one it leaves.
func boxBlur(src, dst []float64, w, h, r, dx, dy int) {
	n, lines, step, stride := w, h, 1, w // along rows
	if dy != 0 {
		n, lines, step, stride = h, w, w, 1 // along columns
	}
	for l := range lines {
		base := l * stride
		var sum float64
		for k := range min(r+1, n) {
			sum += src[base+k*step]
		}
		for i := range n {
			dst[base+i*step] = sum / float64(2*r+1)
			if k := i + r + 1; k < n {
				sum += src[base+k*step]
			}
			if k := i - r; k >= 0 {
				sum -= src[base+k*step]
			}
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
