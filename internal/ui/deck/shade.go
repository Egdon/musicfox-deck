package deck

import "math"

// keyLight comes from the top left in front, in camera space (x right, y
// down, z away from the viewer), pointing towards the light.
var keyLight = vec3{-0.45, -0.7, -0.55}.norm()

const (
	ambient = 0.35
	// A tall strip light to the right, which only the glass reflects: at
	// rest its reflection grazes the window's right end, and as the cassette
	// turns it sweeps across.
	stripAzim  = 0.6 // horizontal angle, radians
	stripWidth = 0.07
)

// restNormal is the A face's normal at the rest pose; lighting is normalized
// so that this face shows the palette colors exactly.
var restNormal = rotation(restYaw, restPitch, 0).apply(vec3{0, 0, -1})

var restNL = restNormal.dot(keyLight)

// scene shades one frame: the cassette pose and its moving parts.
type scene struct {
	m      *materials
	rot    mat3    // object to camera rotation
	inv    float64 // 1 / the object to camera scale
	o      vec3    // the eye in object space
	radii  [2]float64
	angles [2]float64
	blur   float64
	lObj   vec3    // key light in object space
	detail float64 // sub-pixel rows per face unit
}

// sample shades the ray from the eye in camera-space direction d.
func (sc *scene) sample(d vec3) (c rgb, m mat, covered bool) {
	dObj := sc.rot.applyT(d).mul(sc.inv)
	h, ok := intersectShell(sc.o, dObj)
	if !ok {
		return sc.m.bg, matBg, false
	}
	view := d.norm().mul(-1)
	switch h.surf {
	case surfSide:
		albedo, m := sc.m.shell, matSide
		if math.Abs(h.p.z) < 0.3 { // the seam between the two shell halves
			albedo, m = albedo.scale(0.8), matSeam
		}
		return sc.light(albedo, h.n, view, 0.3), m, true
	}
	side, u := 0, h.p.x+faceW/2
	if h.surf == surfBack {
		side, u = 1, faceW/2-h.p.x
	}
	v := h.p.y + faceH/2
	fs := face(sc.m, u, v, side, sc.detail)
	if fs.window {
		c, m := sc.recess(h, dObj, view, side)
		return c, m, true
	}
	return sc.light(fs.albedo, sc.faceNormal(fs.n, side), view, fs.gloss), fs.m, true
}

// faceNormal turns a face-space normal into object space.
func (sc *scene) faceNormal(n vec3, side int) vec3 {
	if side == 1 {
		return vec3{-n.x, n.y, n.z}
	}
	return vec3{n.x, n.y, -n.z}
}

// light applies Lambert shading normalized to the rest face, plus a soft
// specular highlight that never goes brighter than the hero color.
func (sc *scene) light(albedo rgb, nObj, view vec3, gloss float64) rgb {
	n := sc.rot.apply(nObj)
	diffuse := ambient + (1-ambient)*max(n.dot(keyLight), 0)/restNL
	c := albedo.scale(diffuse)
	if gloss > 0 {
		half := keyLight.add(view).norm()
		x := max(n.dot(half), 0)
		x3 := x * x * x
		x6 := x3 * x3
		x12 := x6 * x6
		spec := x12 * x12 * 0.14 * gloss // a power of 24
		c = c.mix(sc.m.hero, min(spec, 0.25))
	}
	return c
}

// recess looks through the window glass: the reels below it shift with the
// viewing angle, the walls catch the light or fall into the shadow of the
// window edge, and the glass reflects the strip light.
func (sc *scene) recess(h shellHit, d, view vec3, side int) (rgb, mat) {
	faceZ, dir := -depth/2, 1.0
	if side == 1 {
		faceZ, dir = depth/2, -1.0
	}
	reelZ := faceZ + dir*reelDepth
	t := (reelZ - sc.o.z) / d.z
	p := sc.o.at(d, t)
	u, v := p.x+faceW/2, p.y+faceH/2
	var c rgb
	var m mat
	if windowDist(u, v) > windowR {
		// The ray meets the recess wall before reaching the reels.
		x := min(max(u, reelCenters[0][0]), reelCenters[1][0])
		l := hypot(u-x, v-reelCenters[0][1])
		wall := vec3{-(u - x) / l, -(v - reelCenters[0][1]) / l, 0}
		c, m = sc.light(sc.m.shell.scale(0.95), wall, view, 0.2), matWall
	} else {
		c, m = sc.reel(p.x, p.y)
		// The window edge shadows the reels where the light comes in at an
		// angle: follow the light back up to the face.
		if math.Abs(sc.lObj.z) > 1e-6 {
			q := p.at(sc.lObj, (faceZ-p.z)/sc.lObj.z)
			if over := windowDist(q.x+faceW/2, q.y+faceH/2) - windowR; over > 0 {
				c = c.scale(1 - 0.42*min(over/0.8, 1))
			}
		}
	}
	// Smoked glass: a faint tint and the reflection of the strip light.
	n := sc.rot.apply(vec3{0, 0, -dir})
	r := view.mul(-1).sub(n.mul(2 * view.mul(-1).dot(n)))
	az := math.Atan2(r.x, -r.z)
	q := (az - stripAzim) / stripWidth
	band := math.Exp(-q*q) * smoothstep(-0.9, -0.2, -r.y)
	return c.mix(sc.m.glass, 0.12).mix(sc.m.hero, 0.2*band), m
}

func smoothstep(e0, e1, x float64) float64 {
	t := min(max((x-e0)/(e1-e0), 0), 1)
	return t * t * (3 - 2*t)
}

func angleDist(a, b float64) float64 {
	d := math.Mod(math.Abs(a-b), 2*math.Pi)
	return min(d, 2*math.Pi-d)
}

// angleNear reports whether angle a lies within tol of target modulo step.
func angleNear(a, target, step, tol float64) bool {
	d := math.Mod(a-target, step)
	if d < 0 {
		d += step
	}
	return d < tol || step-d < tol
}

// reel shades the reel plane at object coordinates (x, y): spindle with six
// teeth, hub ring with the leader clamp, tape pack with grooves and a sheen
// that stays put while the reel turns.
func (sc *scene) reel(x, y float64) (rgb, mat) {
	m := sc.m
	for i, rc := range reelCenters {
		dx, dy := x-(rc[0]-faceW/2), y-(rc[1]-faceH/2)
		r := hypot(dx, dy)
		if r >= sc.radii[i] {
			continue
		}
		ang := math.Atan2(dy, dx)
		rel := ang - sc.angles[i]
		switch {
		case r < toothR:
			return m.hole, matSpindle
		case r < holeR:
			teeth := m.hub.scale(0.92)
			if angleNear(rel, 0, math.Pi/3, 11*math.Pi/180) {
				return teeth.mix(m.hole.mix(teeth, 0.37), sc.blur), matTooth
			}
			return m.hole.mix(m.hole.mix(teeth, 0.37), sc.blur), matSpindle
		case r < hubR:
			c := m.hub
			if r > hubR-0.8 || r < holeR+0.6 {
				c = c.scale(0.87)
			}
			if r > 4.2 && r < 5.9 && angleNear(rel, math.Pi, 2*math.Pi, 0.32) {
				return c.mix(m.clamp, 1-0.9*sc.blur), matClamp
			}
			return c, matHub
		default:
			c := m.tape.scale(1 + 0.05*math.Sin(r*2.6))
			if s := angleDist(ang, -2.2); s < 0.55 {
				c = c.mix(m.hub, 0.12*(1-s/0.55))
			}
			if r > sc.radii[i]-0.8 {
				c = c.scale(1.12)
			}
			return c, matTape
		}
	}
	if math.Abs(y-7.4) < 0.55 && math.Abs(x) < 16 {
		return m.tape.scale(1.1), matTape // tape running along the head side
	}
	return m.glass, matGlass
}
