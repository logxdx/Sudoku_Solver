package main

import (
	"math/rand"
	"time"
)

type Sudoku struct {
	size       int    // Number of regions.
	data       []int  // Numbers on the board.
	locked     []bool // Initial Numbers given. Locked in place.
	rows       *Bitmap
	cols       *Bitmap
	regions    *Bitmap
	rand       *rand.Rand
	generating bool
}

func NewSudoku(size int) *Sudoku {
	sq := size * size
	return &Sudoku{
		size:       size,
		data:       make([]int, sq*sq),
		locked:     make([]bool, sq*sq),
		rows:       NewBitmap(sq, sq),
		cols:       NewBitmap(sq, sq),
		regions:    NewBitmap(sq, sq),
		rand:       rand.New(rand.NewSource(time.Now().UnixNano())),
		generating: false,
	}
}

func (s *Sudoku) Size() int {
	return s.size
}

func (s *Sudoku) Clear() {
	for i := range s.data {
		s.data[i] = 0
		s.locked[i] = false
	}
	s.rows.Clear()
	s.cols.Clear()
	s.regions.Clear()
}

func (s *Sudoku) Generating() bool {
	return s.generating
}

func (s *Sudoku) Has(row, col int) bool {
	return s.At(row, col) != 0
}

func (s *Sudoku) IsLocked(row, col int) bool {
	return s.locked[row*(s.size*s.size)+col]
}

func (s *Sudoku) At(row, col int) int {
	return s.data[row*(s.size*s.size)+col]
}

func (s *Sudoku) Region(row, col int) int {
	return (row/s.size)*s.size + (col / s.size)
}

func (s *Sudoku) Set(row, col, val int) bool {
	region := s.Region(row, col)

	if s.rows.At(row, val-1) || s.cols.At(col, val-1) || s.regions.At(region, val-1) {
		return false
	}

	if s.Has(row, col) {
		old := s.At(row, col) - 1
		s.rows.Set(row, old, false)
		s.cols.Set(col, old, false)
		s.regions.Set(region, old, false)
	}

	s.data[row*(s.size*s.size)+col] = val
	s.rows.Set(row, val-1, true)
	s.cols.Set(col, val-1, true)
	s.regions.Set(region, val-1, true)

	return true
}

func (s *Sudoku) Remove(row, col int) {
	val := s.data[row*(s.size*s.size)+col]
	s.data[row*(s.size*s.size)+col] = 0
	s.rows.Set(row, val-1, false)
	s.cols.Set(col, val-1, false)
	s.regions.Set(s.Region(row, col), val-1, false)
}

const (
	Row    = 0b001
	Col    = 0b010
	Region = 0b100
)

func (s *Sudoku) Conflict(row, col, val int) int {
	if s.At(row, col) == val {
		return 0
	}

	result := 0

	if s.rows.At(row, val-1) {
		result |= Row
	}

	if s.cols.At(col, val-1) {
		result |= Col
	}

	if s.regions.At(s.Region(row, col), val-1) {
		result |= Region
	}

	return result
}

type Difficulty float64

const (
	Easy   Difficulty = 75.0
	Medium Difficulty = 50.0
	Hard   Difficulty = 30.0
)

func (s *Sudoku) Generate(difficulty Difficulty) bool {
	s.generating = true
	s.Clear()

	sq := s.size * s.size

	// Fill in diagonal regions randomly.
	for region := 0; region < s.size; region++ {
		start := s.rand.Intn(1000) + 1
		val := s.rand.Intn(1000) + 1
		i := 0
		for row := 0; row < s.size; row++ {
			for col := 0; col < s.size; col++ {
				r, c := (row+start)%s.size, (col+start)%s.size
				s.Set(region*s.size+r, region*s.size+c, (val+start+i)%sq+1)
				i++
			}
		}
	}

	solvable := s.Solve()
	if !solvable {
		return false
	}

	removed := 0
	remove := sq*sq - int(float64(sq*sq)*float64(difficulty)/100.0)

	for removed < remove {
		row, col := s.rand.Intn(sq), s.rand.Intn(sq)
		if s.Has(row, col) {
			s.Remove(row, col)
			removed++
		}
	}

	for i := range s.data {
		if s.data[i] != 0 {
			s.locked[i] = true
		}
	}

	s.generating = false

	return true
}

func (s *Sudoku) FindNextEmpty(r, c *int) bool {
	sq := s.size * s.size
	for col := 0; col < sq; col++ {
		for row := 0; row < sq; row++ {
			if s.At(row, col) == 0 {
				*r = row
				*c = col
				return true
			}
		}
	}

	return false
}

func (s *Sudoku) Finish() bool {
	s.generating = true
	res := s.Solve()
	s.generating = false

	return res
}

func (s *Sudoku) Solve() bool {
	row, col := 0, 0
	if !s.FindNextEmpty(&row, &col) {
		return true
	}

	start := s.rand.Intn(1000) + 1
	for i := start; i < start+s.size*s.size; i++ {
		if s.Set(row, col, i%(s.size*s.size)+1) {
			if s.Solve() {
				return true
			}
			s.Remove(row, col)
		}
	}

	return false
}
