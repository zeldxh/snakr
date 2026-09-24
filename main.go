package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/nsf/termbox-go"
)

type point struct{ x, y int }

type direction int

const (
	up direction = iota
	down
	left
	right
)

const (
	boardW     = 40
	boardH     = 20
	baseTickMS = 140
)

type game struct {
	snake       []point
	dir         direction
	food        point
	score       int
	highScore   int
	over        bool
	wrap        bool
	obstaclesOn bool
	obstacles   []point
	tickMS      time.Duration
	scores      *highScoreStore
}

func newGame(scores *highScoreStore, wrap, obstaclesOn bool) *game {
	g := &game{
		snake:       []point{{boardW / 2, boardH / 2}},
		dir:         right,
		tickMS:      baseTickMS * time.Millisecond,
		scores:      scores,
		wrap:        wrap,
		obstaclesOn: obstaclesOn,
	}
	if scores != nil {
		g.highScore = scores.load()
	}
	if obstaclesOn {
		g.obstacles = genObstacles(g.snake)
	}
	g.spawnFood()
	return g
}

func (g *game) spawnFood() {
	for {
		p := point{rand.Intn(boardW), rand.Intn(boardH)}
		clash := false
		for _, s := range g.snake {
			if s == p {
				clash = true
				break
			}
		}
		if !clash && g.obstaclesOn && containsPoint(g.obstacles, p) {
			clash = true
		}
		if !clash {
			g.food = p
			return
		}
	}
}

func (g *game) setDirection(d direction) {
	opposite := map[direction]direction{up: down, down: up, left: right, right: left}
	if opposite[g.dir] == d {
		return
	}
	g.dir = d
}

func (g *game) step() {
	if g.over {
		return
	}
	head := g.snake[0]
	next := head
	switch g.dir {
	case up:
		next.y--
	case down:
		next.y++
	case left:
		next.x--
	case right:
		next.x++
	}

	if next.x < 0 || next.x >= boardW || next.y < 0 || next.y >= boardH {
		if !g.wrap {
			g.finish()
			return
		}
		next.x = (next.x + boardW) % boardW
		next.y = (next.y + boardH) % boardH
	}
	for _, s := range g.snake {
		if s == next {
			g.finish()
			return
		}
	}
	if g.obstaclesOn && containsPoint(g.obstacles, next) {
		g.finish()
		return
	}

	g.snake = append([]point{next}, g.snake...)
	if next == g.food {
		g.score++
		g.spawnFood()
		if g.tickMS > 60*time.Millisecond {
			g.tickMS -= 4 * time.Millisecond
		}
	} else {
		g.snake = g.snake[:len(g.snake)-1]
	}
}

func (g *game) finish() {
	g.over = true
	if g.score > g.highScore {
		g.highScore = g.score
		if g.scores != nil {
			_ = g.scores.save(g.highScore)
		}
	}
}

func (g *game) draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	for x := -1; x <= boardW; x++ {
		termbox.SetCell(x+1, 0, '─', termbox.ColorDarkGray, termbox.ColorDefault)
		termbox.SetCell(x+1, boardH+1, '─', termbox.ColorDarkGray, termbox.ColorDefault)
	}
	for y := 0; y <= boardH+1; y++ {
		termbox.SetCell(0, y, '│', termbox.ColorDarkGray, termbox.ColorDefault)
		termbox.SetCell(boardW+1, y, '│', termbox.ColorDarkGray, termbox.ColorDefault)
	}

	for i, s := range g.snake {
		ch := '█'
		color := termbox.ColorGreen
		if i == 0 {
			color = termbox.ColorLightGreen
		}
		termbox.SetCell(s.x+1, s.y+1, ch, color, termbox.ColorDefault)
	}

	termbox.SetCell(g.food.x+1, g.food.y+1, '●', termbox.ColorRed, termbox.ColorDefault)

	for _, o := range g.obstacles {
		termbox.SetCell(o.x+1, o.y+1, '▓', termbox.ColorMagenta, termbox.ColorDefault)
	}

	status := fmt.Sprintf(" score: %d   best: %d ", g.score, g.highScore)
	for i, r := range status {
		termbox.SetCell(i, boardH+2, r, termbox.ColorYellow, termbox.ColorDefault)
	}

	if g.over {
		msg := fmt.Sprintf(" game over — final score %d — press r to restart, q to quit ", g.score)
		if g.score > 0 && g.score == g.highScore {
			msg = fmt.Sprintf(" new high score: %d! — press r to restart, q to quit ", g.score)
		}
		for i, r := range msg {
			termbox.SetCell(i, boardH+3, r, termbox.ColorRed, termbox.ColorDefault)
		}
	} else {
		wrapState := "off"
		if g.wrap {
			wrapState = "on"
		}
		obstaclesState := "off"
		if g.obstaclesOn {
			obstaclesState = "on"
		}
		hint := fmt.Sprintf(" wasd / arrows to move — m: wrap (%s) — o: obstacles (%s) — q to quit ", wrapState, obstaclesState)
		for i, r := range hint {
			termbox.SetCell(i, boardH+3, r, termbox.ColorDarkGray, termbox.ColorDefault)
		}
	}

	termbox.Flush()
}

func main() {
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()

	rand.Seed(time.Now().UnixNano())

	scores, err := newHighScoreStore()
	if err != nil {
		scores = nil
	}
	wrap := false
	obstaclesOn := false
	g := newGame(scores, wrap, obstaclesOn)
	events := make(chan termbox.Event)
	go func() {
		for {
			events <- termbox.PollEvent()
		}
	}()

	g.draw()
	ticker := time.NewTicker(g.tickMS)
	defer ticker.Stop()

	for {
		select {
		case ev := <-events:
			if ev.Type != termbox.EventKey {
				continue
			}
			switch {
			case ev.Key == termbox.KeyEsc, ev.Ch == 'q':
				return
			case ev.Ch == 'r' && g.over:
				g = newGame(scores, wrap, obstaclesOn)
				ticker.Reset(g.tickMS)
			case ev.Ch == 'm' && !g.over:
				wrap = !wrap
				g.wrap = wrap
			case ev.Ch == 'o' && !g.over:
				obstaclesOn = !obstaclesOn
				g.obstaclesOn = obstaclesOn
				if obstaclesOn {
					g.obstacles = genObstacles(g.snake)
					if containsPoint(g.obstacles, g.food) {
						g.spawnFood()
					}
				} else {
					g.obstacles = nil
				}
			case ev.Key == termbox.KeyArrowUp, ev.Ch == 'w':
				g.setDirection(up)
			case ev.Key == termbox.KeyArrowDown, ev.Ch == 's':
				g.setDirection(down)
			case ev.Key == termbox.KeyArrowLeft, ev.Ch == 'a':
				g.setDirection(left)
			case ev.Key == termbox.KeyArrowRight, ev.Ch == 'd':
				g.setDirection(right)
			}
			g.draw()
		case <-ticker.C:
			prevTick := g.tickMS
			g.step()
			g.draw()
			if g.tickMS != prevTick {
				ticker.Reset(g.tickMS)
			}
		}
	}
}
