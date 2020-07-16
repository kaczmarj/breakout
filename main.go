package main

import (
	"github.com/gdamore/tcell"
	"log"
	"math"
	"math/rand"
	"os"
	"time"
)

// TODO: use channels to communicate left/right from the key loop.
// TODO: use channels to implement pause function.

var defaultStyle = tcell.StyleDefault.Background(tcell.NewRGBColor(0, 0, 0))

const refreshRate = time.Second / 30

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err = screen.Init(); err != nil {
		log.Fatal(err)
	}
	defer screen.Fini()
	screen.SetStyle(defaultStyle)
	screen.Clear()

	mainLoop(screen)

}

func mainLoop(screen tcell.Screen) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	directions := []float64{math.Pi / 4, 3 * math.Pi / 4}
	direction := directions[rng.Intn(len(directions))]
	projectile := NewProjectile(screen, Vector2D{direction: direction, magnitude: 1})
	bar := NewBar(screen, 4)

	go keyEventLoop(bar, screen)

	grid, err := gridOfDestructibleRectangles(6, 20, 4, 1, 0, 0, screen)
	if err != nil {
		log.Fatal(err)
	}
	for _, row := range grid {
		for _, block := range row {
			block.Draw()
		}
	}
	screen.Show()

	ready := false
	for !ready {
		event := screen.PollEvent() // This blocks.
		switch event := event.(type) {
		case *tcell.EventKey:
			switch event.Key() {
			case tcell.KeyEnter:
				ready = true
			}
		}
	}

	finished := false
	won := false
	for !finished {
		// Check for collision between projectile and grid of destructible rectangles.
		for _, row := range grid {
			for _, block := range row {
				if !block.destroyed && projectile.CollidesWith(block) {
					block.destroyed = true
					block.Clear()
					projectile.velocity.direction = -projectile.velocity.direction
				}
			}
		}
		// TODO: get fancy here. Modify angle based on where the projectile hit.
		if projectile.CollidesWith(bar) {
			projectile.velocity.direction = -projectile.velocity.direction
		} else if projectile.CollidesWithScreenBottom() {
			finished = true
			won = false
		}
		projectile.Update()
		bar.Draw()
		screen.Show()
		time.Sleep(refreshRate)

		if grid.AllDestroyed() {
			finished = true
			won = true
		}
	}

	if won {
		print("YOU WIN")
	} else {
		print("LOSER")
	}

	time.Sleep(time.Second * 5)

}

func keyEventLoop(b *Bar, s tcell.Screen) {
	for {
		event := b.screen.PollEvent() // This blocks.
		sx, _ := s.Size()
		switch event := event.(type) {
		case *tcell.EventKey:
			switch event.Key() {
			case tcell.KeyLeft:
				b.Clear()
				b.x -= b.step
				if b.x <= 0 {
					b.x = 0
				}
			case tcell.KeyRight:
				b.Clear()
				b.x += b.step
				if b.x+b.w >= sx {
					b.x = sx - b.w
				}
			case tcell.KeyEsc, tcell.KeyCtrlC:
				s.Fini()
				os.Exit(0)
			case tcell.KeyCtrlL:
				s.Sync()
			}
		case *tcell.EventResize:
			s.Sync()
		}
	}
}
