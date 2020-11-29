package elements

import (
	"image/color"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/text"
	"golang.org/x/image/font/inconsolata"
)

// Game contains everything needed for the game
type Game struct {
	player         *Player
	playerChannel  chan int
	gui            *GUI
	score          int
	enemies        []*Enemy
	enemiesChannel []chan int
	enemiesAmount  int
	foods          []*Food
	foodAmount     int
	dotTime        int
	alive          bool
}

// Start will begin all variables needed for the game
func Start(nFood, nEnemies int) Game {
	game := Game{
		score:         0,
		foodAmount:    nFood,
		enemiesAmount: nEnemies,
		dotTime:       0,
		alive:         true,
	}
	foodArray := make([]*Food, game.foodAmount)
	for i := 0; i < game.foodAmount; i++ {
		foodArray[i] = RandomFood(&game)
	}
	enemyArray := make([]*Enemy, game.enemiesAmount)
	for i := 0; i < len(enemyArray); i++ {
		enemyArray[i] = SpawnEnemy(&game)
	}
	enemiesChannel := make([]chan int, game.enemiesAmount)
	for i := 0; i < len(enemiesChannel); i++ {
		enemiesChannel[i] = make(chan int)
		enemyArray[i].behaviour = enemiesChannel[i]
		go enemyArray[i].Behaviour()
	}
	game.enemiesChannel = enemiesChannel
	game.enemies = enemyArray
	game.foods = foodArray
	game.player = Spawn(&game)
	game.playerChannel = make(chan int)
	go game.player.Behaviour()
	game.gui = initializeGUI(&game)
	return game
}

// Update proceeds the game state.
func (g *Game) Update() error {
	if g.alive {
		if g.foodAmount == 0 {
			g.alive = false
		}
		g.dotTime = (g.dotTime + 1) % 10 //game speed

		if err := g.player.Update(g.dotTime); err != nil {
			g.playerChannel <- g.dotTime
		}
		for i := 0; i < len(g.enemies); i++ {
			g.enemiesChannel[i] <- g.dotTime
		}

		x, y := g.player.getHead()

		for i := 0; i < len(g.foods); i++ {
			if x == g.foods[i].foodX && y == g.foods[i].foodY {
				g.foods[i].foodX = -20
				g.foods[i].foodY = -20 //set them out of screen
				g.gui.ateFood()
				g.foodAmount--
				g.player.ateFood()
				break
			}
		}

		for i := 0; i < len(g.enemies); i++ {
			x, y := g.enemies[i].getHead()
			for j := 0; j < len(g.foods); j++ {
				if x == g.foods[j].foodX && y == g.foods[j].foodY {
					g.foods[j].foodX = -20
					g.foods[j].foodY = -20
					g.foodAmount--
					g.enemies[i].ateFood()
					break
				}
			}
		}

	} else {

	}
	for i := 0; i < g.foodAmount; i++ {
		if err := g.foods[i].Update(g.dotTime); err != nil {
			return err
		}
	}
	// Write your game's logical update.
	return nil
}

func (g *Game) enemyDied() {
	g.enemiesAmount--
	g.gui.enemyDied()
}

// GameOver ends the game
func (g *Game) GameOver() {
	g.alive = false
}

// Draw interface, this follows a hierarchy, so snakes have to go at last, while background would have to be the first
func (g *Game) Draw(screen *ebiten.Image) error {

	drawer := &ebiten.DrawImageOptions{}
	drawer.GeoM.Translate(0, 0)
	background, _, _ := ebitenutil.NewImageFromFile("files/background.png", ebiten.FilterDefault)
	screen.DrawImage(background, drawer)

	if err := g.gui.Draw(screen); err != nil {
		return err
	}

	for i := 0; i < len(g.foods); i++ {
		if err := g.foods[i].Draw(screen, g.dotTime); err != nil {
			return err
		}
	}

	for _, enemy := range g.enemies {
		if err := enemy.Draw(screen, g.dotTime); err != nil {
			return err
		}
	}

	if err := g.player.Draw(screen, g.dotTime); err != nil {
		if g.alive == false {
			text.Draw(screen, "game over", inconsolata.Bold8x16, 400, 300, color.White)
		}
		return err
	}
	if g.alive == false {
		largest := g.enemies[0]
		for i := 1; i < len(g.enemies); i++ {
			if g.enemies[i].length > largest.length {
				largest = g.enemies[i]
			}
		}

		textGameOver := ""
		if g.player.length > largest.length {
			textGameOver = "Game over\nYou win. You ate the most food"
		} else {
			textGameOver = "Game over\nYou lose. You did not ate the most food :("
		}
		text.Draw(screen, textGameOver, inconsolata.Bold8x16, 400, 300, color.White)
	}

	return nil
}
