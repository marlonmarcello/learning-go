package main

import "github.com/gdamore/tcell/v2"

type Sprite struct {
	Char rune
	X, Y int
}

func NewSprite(char rune, x, y int) *Sprite {
	return &Sprite{
		Char: char,
		X:    x,
		Y:    y,
	}
}

func (s *Sprite) Draw(screen tcell.Screen) {
	screen.SetContent(
		s.X,
		s.Y,
		s.Char,
		nil,
		tcell.StyleDefault,
	)
}
