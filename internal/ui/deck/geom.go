package deck

import "math"

// The cassette in face units: one unit is about 1.3 mm of a real compact
// cassette. Object space puts the origin at its center with x to the right,
// y down and z away from the viewer; the A side faces -z.
const (
	faceW  = 76.0
	faceH  = 48.0
	depth  = 9.0
	radius = 3.0 // corner radius of the shell outline

	windowR    = 8.5 // radius of the stadium-shaped window
	glassDepth = 0.8 // the window glass sits this far below the face
	reelDepth  = 3.0 // the reels sit this far below the face

	toothR = 1.6 // spindle hole; six teeth reach in to this radius
	holeR  = 3.2
	hubR   = 6.5 // hub ring; also the pack radius of an empty reel
	packR  = 15.0

	// tapeSpeed is the tape speed at the pack surface in face units per
	// second: an empty reel turns at about 41 rpm, a full one at about 18.
	tapeSpeed = 27.9
)

// Reel centers and screws in face coordinates (u right, v down from the
// face's top-left corner).
var (
	reelCenters = [2][2]float64{{22, 24}, {54, 24}}
	screws      = [2][2]float64{{3.6, 44.4}, {72.4, 44.4}}
)

// ReelRadii returns the tape pack radii of the supply and take-up reels.
// Tape area is conserved, so one pack shrinks as the other grows.
func ReelRadii(progress float64) [2]float64 {
	p := min(max(progress, 0), 1)
	span := packR*packR - hubR*hubR
	return [2]float64{math.Sqrt(hubR*hubR + (1-p)*span), math.Sqrt(hubR*hubR + p*span)}
}

type vec3 struct{ x, y, z float64 }

func (a vec3) add(b vec3) vec3           { return vec3{a.x + b.x, a.y + b.y, a.z + b.z} }
func (a vec3) sub(b vec3) vec3           { return vec3{a.x - b.x, a.y - b.y, a.z - b.z} }
func (a vec3) mul(k float64) vec3        { return vec3{a.x * k, a.y * k, a.z * k} }
func (a vec3) dot(b vec3) float64        { return a.x*b.x + a.y*b.y + a.z*b.z }
func (a vec3) norm() vec3                { return a.mul(1 / math.Sqrt(a.dot(a))) }
func (a vec3) at(d vec3, t float64) vec3 { return a.add(d.mul(t)) }

// mat3 is a row-major 3×3 rotation matrix.
type mat3 [9]float64

func (m mat3) apply(v vec3) vec3 {
	return vec3{
		m[0]*v.x + m[1]*v.y + m[2]*v.z,
		m[3]*v.x + m[4]*v.y + m[5]*v.z,
		m[6]*v.x + m[7]*v.y + m[8]*v.z,
	}
}

// applyT applies the transpose, which inverts a rotation.
func (m mat3) applyT(v vec3) vec3 {
	return vec3{
		m[0]*v.x + m[3]*v.y + m[6]*v.z,
		m[1]*v.x + m[4]*v.y + m[7]*v.z,
		m[2]*v.x + m[5]*v.y + m[8]*v.z,
	}
}

func (m mat3) mul(n mat3) mat3 {
	var r mat3
	for i := range 3 {
		for j := range 3 {
			r[i*3+j] = m[i*3]*n[j] + m[i*3+1]*n[3+j] + m[i*3+2]*n[6+j]
		}
	}
	return r
}

// rotation turns the cassette: roll about its own z axis first, then pitch
// about x (positive shows the top wall), then yaw about the vertical axis
// (negative shows the left wall).
func rotation(yaw, pitch, roll float64) mat3 {
	cy, sy := math.Cos(yaw), math.Sin(yaw)
	cp, sp := math.Cos(pitch), math.Sin(pitch)
	cr, sr := math.Cos(roll), math.Sin(roll)
	ry := mat3{cy, 0, sy, 0, 1, 0, -sy, 0, cy}
	rx := mat3{1, 0, 0, 0, cp, -sp, 0, sp, cp}
	rz := mat3{cr, -sr, 0, sr, cr, 0, 0, 0, 1}
	return ry.mul(rx).mul(rz)
}

// slab returns where a ray is inside lo..hi along one axis.
func slab(o, d, lo, hi float64) (t0, t1 float64, ok bool) {
	if math.Abs(d) < 1e-12 {
		if o < lo || o > hi {
			return 0, 0, false
		}
		return math.Inf(-1), math.Inf(1), true
	}
	t0, t1 = (lo-o)/d, (hi-o)/d
	if t0 > t1 {
		t0, t1 = t1, t0
	}
	return t0, t1, true
}

// circleT solves |o + t·d − c| = r for t in the plane.
func circleT(ox, oy, dx, dy, cx, cy, r float64) (t0, t1 float64, ok bool) {
	px, py := ox-cx, oy-cy
	a := dx*dx + dy*dy
	if a < 1e-12 {
		return 0, 0, false
	}
	b := px*dx + py*dy
	c := px*px + py*py - r*r
	disc := b*b - a*c
	if disc < 0 {
		return 0, 0, false
	}
	s := math.Sqrt(disc)
	return (-b - s) / a, (-b + s) / a, true
}

// roundRect intersects a ray in the plane with the shell outline, a rounded
// rectangle centered at the origin. It returns the entry and exit parameters
// and the outward normal at the entry.
func roundRect(ox, oy, dx, dy float64) (tin, tout, nx, ny float64, ok bool) {
	hw, hh := faceW/2, faceH/2
	ax, ay := hw-radius, hh-radius
	tx0, tx1, okx := slab(ox, dx, -hw, hw)
	ty0, ty1, oky := slab(oy, dy, -hh, hh)
	if !okx || !oky {
		return 0, 0, 0, 0, false
	}
	tin, tout = max(tx0, ty0), min(tx1, ty1)
	if tin > tout {
		return 0, 0, 0, 0, false
	}
	if tx0 > ty0 {
		nx = -math.Copysign(1, dx)
	} else {
		ny = -math.Copysign(1, dy)
	}
	// An entry or exit point in a corner square really lies on the corner
	// arc; if the ray misses that circle it passes outside the corner.
	if px, py := ox+tin*dx, oy+tin*dy; math.Abs(px) > ax && math.Abs(py) > ay {
		cx, cy := math.Copysign(ax, px), math.Copysign(ay, py)
		t0, _, hit := circleT(ox, oy, dx, dy, cx, cy, radius)
		if !hit {
			return 0, 0, 0, 0, false
		}
		tin = t0
		nx, ny = (ox+t0*dx-cx)/radius, (oy+t0*dy-cy)/radius
	}
	if px, py := ox+tout*dx, oy+tout*dy; math.Abs(px) > ax && math.Abs(py) > ay {
		cx, cy := math.Copysign(ax, px), math.Copysign(ay, py)
		_, t1, hit := circleT(ox, oy, dx, dy, cx, cy, radius)
		if !hit {
			return 0, 0, 0, 0, false
		}
		tout = t1
	}
	return tin, tout, nx, ny, tin <= tout
}

type surface uint8

const (
	surfNone surface = iota
	surfFront
	surfBack
	surfSide
)

// shellHit is where a ray enters the shell, in object space.
type shellHit struct {
	t    float64
	surf surface
	p    vec3
	n    vec3 // outward normal
}

// intersectShell intersects an object-space ray with the shell, a rounded
// rectangle extruded along z. The shell is convex, so the ray enters where it
// is inside both the z slab and the outline.
func intersectShell(o, d vec3) (shellHit, bool) {
	if h, ok := faceHit(o, d); ok {
		return h, true
	}
	tz0, tz1, ok := slab(o.z, d.z, -depth/2, depth/2)
	if !ok {
		return shellHit{}, false
	}
	t0, t1, nx, ny, ok := roundRect(o.x, o.y, d.x, d.y)
	if !ok {
		return shellHit{}, false
	}
	tin, tout := max(tz0, t0), min(tz1, t1)
	if tin > tout || tout <= 0 {
		return shellHit{}, false
	}
	h := shellHit{t: tin, p: o.at(d, tin)}
	switch {
	case tz0 >= t0 && d.z > 0:
		h.surf, h.n = surfFront, vec3{0, 0, -1}
	case tz0 >= t0:
		h.surf, h.n = surfBack, vec3{0, 0, 1}
	default:
		h.surf, h.n = surfSide, vec3{nx, ny, 0}
	}
	return h, true
}

// faceHit is the quick way in for most rays: through the face turned to the
// eye, at a point clear of the outline's edges and corners. intersectShell
// finds the same hit there, by the same arithmetic.
func faceHit(o, d vec3) (shellHit, bool) {
	var h shellHit
	z := 0.0
	switch {
	case o.z < -depth/2 && d.z > 0:
		z, h.surf, h.n = -depth/2, surfFront, vec3{0, 0, -1}
	case o.z > depth/2 && d.z < 0:
		z, h.surf, h.n = depth/2, surfBack, vec3{0, 0, 1}
	default:
		return shellHit{}, false
	}
	const e = 1e-6
	t := (z - o.z) / d.z
	p := o.at(d, t)
	ax, ay := math.Abs(p.x), math.Abs(p.y)
	if ax > faceW/2-e || ay > faceH/2-e || (ax > faceW/2-radius-e && ay > faceH/2-radius-e) {
		return shellHit{}, false
	}
	h.t, h.p = t, p
	return h, true
}

// shellSD is the signed distance to the shell outline in face coordinates.
func shellSD(u, v float64) float64 {
	return roundRectSD(u, v, faceW, faceH, radius)
}

func roundRectSD(u, v, w, h, r float64) float64 {
	qx := math.Abs(u-w/2) - (w/2 - r)
	qy := math.Abs(v-h/2) - (h/2 - r)
	return hypot(max(qx, 0), max(qy, 0)) + min(max(qx, qy), 0) - r
}

// hypot is math.Hypot for the small lengths of the cassette, several times
// faster. It is exact when either side is zero, as most points away from
// corners and reel centers are.
func hypot(x, y float64) float64 {
	switch {
	case x == 0:
		return math.Abs(y)
	case y == 0:
		return math.Abs(x)
	}
	return math.Sqrt(x*x + y*y)
}

// shellNormal2D is the outward normal of the outline nearest to (u, v).
func shellNormal2D(u, v float64) (float64, float64) {
	const e = 0.05
	nx := shellSD(u+e, v) - shellSD(u-e, v)
	ny := shellSD(u, v+e) - shellSD(u, v-e)
	l := hypot(nx, ny)
	if l == 0 {
		return 0, -1
	}
	return nx / l, ny / l
}

// windowDist is the distance to the axis of the stadium-shaped window.
func windowDist(u, v float64) float64 {
	x := min(max(u, reelCenters[0][0]), reelCenters[1][0])
	return hypot(u-x, v-reelCenters[0][1])
}
