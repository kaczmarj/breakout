package main

import (
	"errors"
	"github.com/gdamore/tcell"
	"math"
	"math/rand"
	"time"
)

type Boundser interface {
	Bounds() (int, int, int, int)
}

type Rectangle struct {
	x, y   int
	w, h   int
	screen tcell.Screen
	style  tcell.Style
	mainc  rune
	combc  []rune
}

// Bounds returns the bounds of the object, x0, x1, y0, y1.
func (r *Rectangle) Bounds() (int, int, int, int) {
	return r.x, r.x + r.w, r.y, r.y + r.h
}

func (r *Rectangle) Draw() {
	for j := r.y; j < r.y+r.h; j++ {
		for i := r.x; i < r.x+r.w; i++ {
			r.screen.SetContent(i, j, r.mainc, r.combc, r.style)
		}
	}
}

func (r *Rectangle) Clear() {
	for j := r.y; j < r.y+r.h; j++ {
		for i := r.x; i < r.x+r.w; i++ {
			r.screen.SetContent(i, j, r.mainc, r.combc, defaultStyle)
		}
	}
}

func (r *Rectangle) CollidesWithScreenTop() bool {
	return r.y <= 0
}

func (r *Rectangle) CollidesWithScreenBottom() bool {
	_, sy := r.screen.Size()
	return (r.y + r.h) >= sy
}

func (r *Rectangle) CollidesWithScreenLeft() bool {
	return r.x <= 0
}

func (r *Rectangle) CollidesWithScreenRight() bool {
	sx, _ := r.screen.Size()
	return (r.x + r.w) >= sx
}

func (r *Rectangle) CollidesWith(other Boundser) bool {
	ax0, ax1, ay0, ay1 := r.Bounds()
	bx0, bx1, by0, by1 := other.Bounds()
	return ax0 < bx1 && bx0 < ax1 && ay0 < by1 && by0 < ay1
}

type Bar struct {
	*Rectangle
	step int
}

func NewBar(s tcell.Screen, w, step int) *Bar {
	sx, sy := s.Size()
	h := 1
	x := sx/2 - (w / 2)
	y := sy - 5
	style := tcell.StyleDefault.Background(tcell.NewRGBColor(0, 128, 128))
	r := &Rectangle{x: x, y: y, w: w, h: h, screen: s, style: style, mainc: ' ', combc: nil}
	return &Bar{Rectangle: r, step: step}
}

type Vector2D struct {
	direction float64
	magnitude float64
}

type Projectile struct {
	*Rectangle
	velocity Vector2D
	// We're in space! Forget gravity.
}

func NewProjectile(s tcell.Screen, velocity Vector2D) *Projectile {
	sx, sy := s.Size()
	w := 2
	h := 1
	x := sx / 2
	y := sy/2 + 2*w
	style := tcell.StyleDefault.Background(tcell.NewRGBColor(128, 128, 128))
	r := &Rectangle{x: x, y: y, w: w, h: h, screen: s, style: style, mainc: ' ', combc: nil}
	return &Projectile{Rectangle: r, velocity: velocity}
}

func (p *Projectile) Update() {
	p.Clear()
	if p.CollidesWithScreenTop() || p.CollidesWithScreenBottom() {
		p.velocity.direction = -p.velocity.direction
	}
	if p.CollidesWithScreenLeft() || p.CollidesWithScreenRight() {
		p.velocity.direction = math.Pi - p.velocity.direction
	}
	// TODO: Modify direction if it is close to horizontal or vertical.
	//  This only applies if we change the angles based on where the projectile hit the bar.
	p.velocity.direction = math.Mod(p.velocity.direction, 2*math.Pi)

	// Update origin.
	p.x += int(math.Round(p.velocity.magnitude * math.Cos(p.velocity.direction)))
	p.y -= int(math.Round(p.velocity.magnitude * math.Sin(p.velocity.direction)))
	p.Draw()
}

type DestructibleRectangle struct {
	*Rectangle
	destroyed bool
}

type DestructibleGrid [][]*DestructibleRectangle

func (d DestructibleGrid) AllDestroyed() bool {
	for _, row := range d {
		for _, block := range row {
			if !block.destroyed {
				return false
			}
		}
	}
	return true
}

func NewDestructibleGrid(r, c int, w, h int, marginx, marginy int, s tcell.Screen) (DestructibleGrid, error) {
	var err error = nil

	gridWidth := c * (w + marginx)
	gridHeight := r * (h + marginy)

	sx, sy := s.Size()
	if gridWidth > sx || gridHeight >= sy {
		s.Fini()
		err = errors.New("grid out of bounds of screen")
	}

	x := sx/2 - (gridWidth / 2)
	y := 0

	colors := [5][3]int32{{46, 134, 171}, {162, 59, 114}, {241, 143, 1}, {199, 62, 29}, {59, 31, 43}}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	grid := make([][]*DestructibleRectangle, c)
	for j := 0; j < c; j++ {
		grid[j] = make([]*DestructibleRectangle, r)
		for i := 0; i < r; i++ {
			rgb := colors[rng.Intn(len(colors))]
			grid[j][i] = &DestructibleRectangle{Rectangle: &Rectangle{
				x:      j*(w+marginx) + x,
				y:      i*(h+marginy) + y,
				w:      w,
				h:      h,
				screen: s,
				style:  tcell.StyleDefault.Background(tcell.NewRGBColor(rgb[0], rgb[1], rgb[2])),
				mainc:  ' ',
				combc:  nil},
				destroyed: false,
			}
		}
	}
	return grid, err
}
