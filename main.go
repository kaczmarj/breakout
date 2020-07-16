package main

import (
	"flag"
	"github.com/gdamore/tcell"
	"log"
	"math"
	"math/rand"
	"os"
	"time"
)

// TODO: use channels to communicate left/right from the key loop.
// TODO: use channels to implement pause function.
// TODO: use channels to start game. Put the Enter key in the key event loop.

var defaultStyle = tcell.StyleDefault.Background(tcell.NewRGBColor(0, 0, 0))

const refreshRate = time.Second / 30

func main() {

	platformWidth := flag.Int("platform-width", 16, "width of platform")
	platformStep := flag.Int("platform-step", 6, "number of characters platform shifts per move")
	speed := flag.Float64("speed", 1, "speed of projectile")
	rows := flag.Int("rows", 6, "rows in grid of breakable blocks")
	cols := flag.Int("cols", 20, "columns in grid of breakable blocks")
	blockWidth := flag.Int("block-width", 4, "width of each breakable block")
	blockHeight := flag.Int("block-height", 1, "height of each breakable block")
	flag.Parse()

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

	// rows, cols
	mainLoop(screen, *platformWidth, *platformStep, *speed, *rows, *cols, *blockWidth, *blockHeight)

}

func mainLoop(screen tcell.Screen, platformWidth, platformStep int, speed float64, rows, cols, blockWidth, blockHeight int) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	directions := []float64{math.Pi / 4, 3 * math.Pi / 4}
	direction := directions[rng.Intn(len(directions))]
	projectile := NewProjectile(screen, Vector2D{direction: direction, magnitude: speed})
	bar := NewBar(screen, platformWidth, platformStep)

	go keyEventLoop(bar, screen)

	grid, err := NewDestructibleGrid(rows, cols, 4, 1, 1, 0, screen)
	if err != nil {
		log.Fatal(err)
	}
	for _, row := range grid {
		for _, block := range row {
			block.Draw()
		}
	}
	projectile.Draw()
	bar.Draw()

	sx, sy := screen.Size()
	style := defaultStyle.Foreground(tcell.NewRGBColor(162, 59, 114))
	// Write instructions.
	instructions := "Press enter to begin."
	x0 := sx/2 - (len(instructions) / 2)
	y0 := sy / 2
	for i, ch := range []rune(instructions) {
		screen.SetContent(x0+i, y0, ch, nil, style)
	}
	quitInstructions := "Press escape to quit."
	for i, ch := range []rune(quitInstructions) {
		screen.SetContent(x0+i, y0+1, ch, nil, style)
	}
	screen.Show()

	// Wait for user input.
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

	// Remove instructions.
	for i, _ := range []rune(instructions) {
		screen.SetContent(x0+i, y0, ' ', nil, defaultStyle)
	}
	for i, _ := range []rune(quitInstructions) {
		screen.SetContent(x0+i, y0+1, ' ', nil, style)
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

	time.Sleep(time.Second * 3)

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
