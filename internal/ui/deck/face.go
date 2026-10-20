package deck

import "math"

// mat tells surfaces apart: neighboring samples of different materials mark
// an edge worth supersampling.
type mat uint8

const (
	matBg mat = iota
	matShell
	matBevel
	matSide
	matSeam
	matScrew
	matRim
	matLabel
	matStripe
	matInk
	matHead
	matHole
	matWall
	matGlass
	matSpindle
	matTooth
	matHub
	matClamp
	matTape
)

// faceSample describes a point of the A or B face before lighting. The normal
// is in face space: x along u, y along v, z pointing out of the face.
type faceSample struct {
	albedo rgb
	n      vec3
	gloss  float64
	m      mat
	window bool // the window: look into the recess instead
}

var flatOut = vec3{0, 0, 1}

// Side letters as 5×7 bitmaps, one string per row.
var letters = [2][7]string{
	{".###.", "#...#", "#...#", "#####", "#...#", "#...#", "#...#"},
	{"####.", "#...#", "#...#", "####.", "#...#", "#...#", "####."},
}

const (
	letterU, letterV = 6.6, 3.9 // top-left of the side letter
	letterPx         = 1.05     // size of one bitmap pixel
)

// face shades the printed and molded details of one face; side 0 is A.
func face(m *materials, u, v float64, side int) faceSample {
	sd := shellSD(u, v)
	if sd > -1 { // chamfered rim
		nx, ny := shellNormal2D(u, v)
		return faceSample{albedo: m.shell, n: vec3{nx * 0.7, ny * 0.7, 0.7}.norm(), gloss: 0.6, m: matBevel}
	}
	for _, s := range screws {
		if dx, dy := u-s[0], v-s[1]; dx*dx+dy*dy < 1.3 {
			return faceSample{albedo: m.screw, n: vec3{dx, dy, 1.2}.norm(), gloss: 0.8, m: matScrew}
		}
	}
	dw := windowDist(u, v)
	if dw <= windowR {
		return faceSample{window: true, m: matGlass}
	}
	if dw <= windowR+0.9 {
		// The rim slopes down into the window.
		x := min(max(u, reelCenters[0][0]), reelCenters[1][0])
		nx, ny := (u-x)/dw, (v-reelCenters[0][1])/dw
		return faceSample{albedo: m.shell.scale(0.9), n: vec3{-nx * 0.5, -ny * 0.5, 0.85}.norm(), gloss: 0.5, m: matRim}
	}
	if roundRectSD(u-4, v-3, 68, 32, 1.5) <= 0 {
		return labelAt(m, u, v, side)
	}
	if v >= 37 { // raised head area with two capstan holes and the head opening
		spread := (v - 37) * 4 / 11
		if u >= 20-spread && u <= 56+spread {
			for _, hx := range [2]float64{26.5, 49.5} {
				if math.Hypot(u-hx, v-44) < 1.4 {
					return faceSample{albedo: m.hole, n: flatOut, m: matHole}
				}
			}
			if u >= 35 && u <= 41 && v >= 44.5 {
				return faceSample{albedo: m.hole, n: flatOut, m: matHole}
			}
			if v < 38 {
				return faceSample{albedo: m.shell, n: vec3{0, -0.6, 0.8}, gloss: 0.5, m: matHead}
			}
			return faceSample{albedo: m.shell.scale(1.05), n: flatOut, gloss: 0.35, m: matHead}
		}
	}
	return faceSample{albedo: m.shell, n: flatOut, gloss: 0.35, m: matShell}
}

// labelAt prints the label: the side letter, three writing lines, the signal
// orange stripe and a row of small boxes under the window.
func labelAt(m *materials, u, v float64, side int) faceSample {
	paper := faceSample{albedo: m.label, n: flatOut, gloss: 0.1, m: matLabel}
	ink := faceSample{albedo: m.ink, n: flatOut, gloss: 0.1, m: matInk}
	if v >= 11.6 && v < 13.1 && u >= 6 && u <= 70 {
		return faceSample{albedo: m.stripe, n: flatOut, gloss: 0.1, m: matStripe}
	}
	if bx, by := (u-letterU)/letterPx, (v-letterV)/letterPx; bx >= 0 && by >= 0 && bx < 5 && by < 7 {
		if letters[side][int(by)][int(bx)] == '#' {
			return ink
		}
		return paper
	}
	if u >= 15 && u <= 69 {
		for _, lv := range [3]float64{5.0, 7.5, 10.0} {
			if v >= lv && v < lv+0.9 {
				return ink
			}
		}
	}
	if v >= 33.2 && v < 34.4 {
		for i := range 4 {
			if bu := 59.0 + 2.6*float64(i); u >= bu && u < bu+1.2 {
				return ink
			}
		}
	}
	return paper
}
