package main

import (
	"fmt"
	"image"
	"math"

	"github.com/fogleman/gg"
	"github.com/goki/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/font/gofont/gobold"
)

const (
	Width = 800                                    // Window width.
	Size  = 4                                      // Between 2 and 5.
	Level = Medium                                 // Difficulty.
	Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ" // Lowercase not supported.
)

type Point struct {
	X, Y float64
}

type Game struct {
	Canvas   *gg.Context
	Board    *Sudoku
	Current  Point
	width    float64
	offset   float64
	cellsize float64
	hrow     float64 // Highlighted row.
	hcol     float64 // Highlighted column.
	htimer   int     // Highlighting timer.
	conflict int     // Conflict flag.
	number   int     // Conflicting number.
}

func NewGame(screenwidth, boardsize int) *Game {
	game := &Game{
		Canvas:  gg.NewContext(screenwidth, screenwidth),
		Board:   NewSudoku(boardsize),
		Current: Point{0.0, 0.0},
		width:   float64(screenwidth),
	}
	game.offset = game.width / 20.0
	game.cellsize = (game.width - game.offset*2.0) / float64(boardsize*boardsize)

	font, err := truetype.Parse(gobold.TTF)
	if err != nil {
		panic(err)
	}

	// Find fitting font size.
	size := 100.0
	for {
		face := truetype.NewFace(font, &truetype.Options{Size: size})
		game.Canvas.SetFontFace(face)
		w, h := game.Canvas.MeasureString("8")
		if !(w > game.cellsize || h > game.cellsize) {
			break
		}
		size--
	}

	return game
}

func (g *Game) drawHighlight() {
	dc := g.Canvas
	size := g.Board.Size()

	dc.SetRGBA255(255, 70, 70, g.htimer)

	if g.conflict&Col != 0 {
		dc.DrawRectangle(g.offset, g.hcol*g.cellsize+g.offset, g.width-g.offset*2.0, g.cellsize)
	}

	if g.conflict&Row != 0 {
		dc.DrawRectangle(g.hrow*g.cellsize+g.offset, g.offset, g.cellsize, g.width-g.offset*2.0)
	}

	if g.conflict&Region != 0 {
		region := g.Board.Region(int(g.hrow), int(g.hcol))
		rr, rc := region/size, region%size
		dc.DrawRectangle(
			float64(rr*size)*g.cellsize+g.offset,
			float64(rc*size)*g.cellsize+g.offset,
			g.cellsize*float64(size),
			g.cellsize*float64(size),
		)
	}

	dc.Fill()

	dc.SetRGBA255(0, 0, 0, g.htimer)
	dc.DrawStringAnchored(
		string(Chars[g.number]),
		(g.hrow+1.0)*g.cellsize-0.5*g.cellsize+g.offset,
		(g.hcol+1.0)*g.cellsize-0.6*g.cellsize+g.offset,
		0.5,
		0.5,
	)
}

func (g *Game) Update() error {
	dc := g.Canvas
	size := g.Board.Size()

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		// Highlight square when clicked.
		x, y := ebiten.CursorPosition()
		if x >= 0 && x < dc.Width() && y >= 0 && y < dc.Height() {
			g.Current.X = math.Floor((float64(x) - g.offset) / g.cellsize)
			g.Current.Y = math.Floor((float64(y) - g.offset) / g.cellsize)
		}
	} else {
		// Move selected square with arrow keys.
		if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
			g.Current.X += 1.0
		} else if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
			g.Current.X -= 1.0
		} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
			g.Current.Y += 1.0
		} else if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
			g.Current.Y -= 1.0
		}
		// Keep selected square in bounds.
		g.Current.X = math.Min(math.Max(0, g.Current.X), float64(size*size-1))
		g.Current.Y = math.Min(math.Max(0, g.Current.Y), float64(size*size-1))
	}

	// Backspace removes number on selected square.
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		x, y := int(g.Current.X), int(g.Current.Y)
		if !g.Board.IsLocked(x, y) {
			g.Board.Remove(x, y)
		}
	}
	// Generate Sudoku with difficulty.
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if !g.Board.Generating() {
			go g.Board.Generate(Level)
		}
	}
	// Solve current board.
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if !g.Board.Generating() {
			go g.Board.Finish()
		}
	}
	// Clear Sudoku board.
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if !g.Board.Generating() {
			g.Board.Clear()
		}
	}

	// Clear Canvas.
	dc.SetRGB255(255, 255, 255)
	dc.Clear()

	// Draw Gray Squares for locked numbers.
	for row := 0; row < size*size; row++ {
		for col := 0; col < size*size; col++ {
			if g.Board.IsLocked(row, col) {
				dc.SetRGB255(170, 170, 170)
				dc.DrawRectangle(float64(row)*g.cellsize+g.offset, float64(col)*g.cellsize+g.offset, g.cellsize, g.cellsize)
				dc.Fill()
			}
		}
	}

	// Draw Selected Square.
	dc.SetRGB255(88, 150, 236)
	dc.DrawRectangle(g.Current.X*g.cellsize+g.offset, g.Current.Y*g.cellsize+g.offset, g.cellsize, g.cellsize)
	dc.Fill()

	// Add Number to selected square when corresponding key is pressed.
	for i := 1; i <= size*size || i >= len(Chars); i++ {
		var key ebiten.Key
		if err := key.UnmarshalText([]byte(string(Chars[i]))); err != nil {
			continue
		}
		if inpututil.IsKeyJustPressed(key) {
			x, y := int(g.Current.X), int(g.Current.Y)
			if !g.Board.Set(x, y, i) {
				g.conflict = g.Board.Conflict(x, y, i)
				g.hrow = g.Current.X
				g.hcol = g.Current.Y
				g.htimer = 255
				g.number = i
				g.drawHighlight()
			}
			break
		}
	}

	if g.htimer > 0 {
		g.drawHighlight()
		g.htimer--
	}

	// Draw Numbers.
	for row := 0; row < size*size; row++ {
		for col := 0; col < size*size; col++ {
			if g.Board.Has(row, col) {
				dc.SetRGB255(0, 0, 0)
				dc.DrawStringAnchored(
					string(Chars[g.Board.At(row, col)]),
					float64(row+1)*g.cellsize-0.5*g.cellsize+g.offset,
					float64(col+1)*g.cellsize-0.6*g.cellsize+g.offset,
					0.5,
					0.5,
				)
			}
		}
	}

	// Draw Surrounding Box.
	dc.SetRGB255(0, 0, 0)
	dc.SetLineWidth(8.0)
	dc.DrawRectangle(g.offset, g.offset, g.width-g.offset*2.0, g.width-g.offset*2.0)
	dc.Stroke()

	// Draw Grid Lines.
	for i := 1; i < size*size; i++ {
		if i%size == 0 {
			dc.SetLineWidth(5.0)
		} else {
			dc.SetLineWidth(2.0)
		}
		pos := g.offset + float64(i)*g.cellsize
		dc.DrawLine(pos, g.offset, pos, g.width-g.offset)
		dc.DrawLine(g.offset, pos, g.width-g.offset, pos)
		dc.Stroke()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	img := g.Canvas.Image().(*image.RGBA)
	screen.WritePixels(img.Pix)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.Canvas.Width(), g.Canvas.Width()
}

func main() {
	max := int(math.Sqrt(float64(len(Chars) - 1)))
	if Size < 2 || Size > max {
		panic(fmt.Sprintf("board size: %d must be between 2 and %d", Size, max))
	}

	game := NewGame(Width, Size)

	ebiten.SetWindowSize(Width, Width)
	ebiten.SetWindowTitle(fmt.Sprintf("%dx%d Sudoku", Size, Size))

	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
