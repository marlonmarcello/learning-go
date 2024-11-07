package main

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

func drawString(screen tcell.Screen, x, y int, msg string) {
	for index, char := range msg {
		screen.SetContent(x+index, y, char, nil, tcell.StyleDefault)
	}
}

func generateCoins(level int) []*Sprite {
	total := level + 2
	coins := make([]*Sprite, total)

	for index := range total {
		coins[index] = NewSprite(
			'0',
			rand.Intn(20),
			rand.Intn(20),
		)
	}

	return coins
}

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	defer screen.Fini()

	err = screen.Init()
	if err != nil {
		log.Fatal(err)
	}

	// ----------------------- game init
	player := NewSprite('@', 10, 10)

	level := 1

	coins := generateCoins(level)

	score := 0

	// game loop
	running := true
	for running {
		// --------------------- draw screen

		screen.Clear()

		// player
		player.Draw(screen)

		// coins
		for _, coin := range coins {
			coin.Draw(screen)
		}

		// ui
		drawString(
			screen,
			1,
			1,
			fmt.Sprintf("Score: %d", score),
		)

		drawString(
			screen,
			1,
			2,
			fmt.Sprintf("Level: %d", level),
		)

		screen.Show()

		// --------------------- update state

		playerMoved := false

		// getting event
		ev := screen.PollEvent()

		// checking event type
		switch ev := ev.(type) {
		case *tcell.EventKey:
			// checking event key
			switch ev.Rune() {
			case 'q':
				running = false
			case 'w':
				playerMoved = true
				player.Y -= 1
			case 's':
				playerMoved = true
				player.Y += 1
			case 'a':
				playerMoved = true
				player.X -= 1
			case 'd':
				playerMoved = true
				player.X += 1
			}
		}

		if playerMoved {
			coinCollectedIndex := -1

			for index, coin := range coins {
				if coin.X == player.X && coin.Y == player.Y {
					// collect coin
					coinCollectedIndex = index
					score++
				}
			}

			if coinCollectedIndex >= 0 {
				// swap collected with last item
				coins[coinCollectedIndex] = coins[len(coins)-1]
				// remove last item
				coins = coins[0 : len(coins)-1]

				if len(coins) == 0 {
					level++
					coins = generateCoins(level)
				}
			}
		}
	}
}
