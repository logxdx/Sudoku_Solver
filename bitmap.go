package main

import "math/bits"

type Bitmap struct {
	width  int
	height int
	data   []uint64
}

func NewBitmap(width, height int) *Bitmap {
	total := width * height
	size := total / 64

	if total%64 != 0 {
		size++
	}

	return &Bitmap{
		width:  width,
		height: height,
		data:   make([]uint64, size),
	}
}

func (b *Bitmap) Width() int {
	return b.width
}

func (b *Bitmap) Height() int {
	return b.height
}

func (b *Bitmap) Size() int {
	total := 0
	for i := range b.data {
		total += bits.OnesCount64(b.data[i])
	}
	return total
}

func (b *Bitmap) Copy() *Bitmap {
	new := NewBitmap(b.width, b.height)
	copy(new.data, b.data)
	return new
}

func (b *Bitmap) At(x, y int) bool {
	index := y*b.width + x
	return (b.data[index/64] & (1 << (index & 63))) != 0
}

func (b *Bitmap) Set(x, y int, value bool) {
	index := y*b.width + x
	if value {
		b.data[index/64] |= (1 << (index & 63))
	} else {
		b.data[index/64] &= ^(1 << (index & 63))
	}
}

func (b *Bitmap) Clear() {
	for i := range b.data {
		b.data[i] = 0
	}
}
