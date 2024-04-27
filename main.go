package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"os"
	"slices"
	"time"

	"github.com/adrg/sysfont"
	"github.com/bwmarrin/discordgo"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/joho/godotenv"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	discordChannelID = "960497802830573618"
)

const (
	cellWidth  = 40
	cellHeight = 40

	screenWidth  = 840 + cellWidth*2
	screenHeight = 840 + cellWidth*2
	botTokenKey  = "botToken"

	rowSize       = screenWidth / cellWidth
	columnSize    = screenHeight / cellHeight
	sizeOfGameMap = rowSize * columnSize
)

const (
	emptyCell     = iota
	borderCell    = iota
	snakeHeadCell = iota
	snakeBodyCell = iota
	foodCell      = iota
)

const (
	directionUp    = iota
	directionRight = iota
	directionDown  = iota
	directionLeft  = iota
)

const (
	upMessage    = "up"
	rightMessage = "right"
	downMessage  = "down"
	leftMessage  = "left"
)

var (
	borderColor    = color.RGBA{44, 5, 144, 0.8 * 255}
	snakeHeadColor = color.RGBA{230, 195, 11, 1}
	snakeBodyColor = color.RGBA{200, 175, 5, 1}
	foodColor      = color.RGBA{255, 0, 0, 1}
)

var (
	messagesArray = make([]string, 0)
)

const (
	moveDelay  = 40
	resetDelay = 5000
)

var (
	lastUpdateTime time.Time
	moveCounter    int64
	resetCounter   int64
)

var (
	defaultFont font.Face
)

func indexToColumnRow(index int) (int, int) {
	row := index % rowSize
	column := index / columnSize
	return column, row
}

func helperDrawRect(screen *ebiten.Image, row int, column int, color color.RGBA) {
	vector.DrawFilledRect(screen, float32(row*cellWidth), float32(column*cellHeight), cellWidth, cellHeight, color, false)
}

type Game struct {
	GameMap           []int
	SnakeHead         int
	PreviousSnakeHead int
	SnakeBody         []int
	MovementDirection int
	FoodPosition      int
	IsGameOver        bool
	Score             int
}

func (g *Game) Reset() {
	g.Score = 0
	g.SnakeBody = make([]int, 0)
	g.GameMap = make([]int, sizeOfGameMap)
	g.SnakeHead = (rowSize * rowSize / 2)
	g.MovementDirection = directionUp
	g.FoodPosition = -1
	g.IsGameOver = false
	moveCounter = 0
	resetCounter = 0
}

func (g *Game) Update() error {
	currentTime := time.Now()
	deltaTime := currentTime.Sub(lastUpdateTime).Milliseconds()

	if g.IsGameOver {
		resetCounter += deltaTime

		if resetCounter > resetDelay {
			g.Reset()
		} else {
			lastUpdateTime = currentTime
		}

		return nil
	}

	if len(messagesArray) != 0 {
		lastMessage := messagesArray[len(messagesArray)-1]
		messagesArray = make([]string, 0)

		switch lastMessage {
		case upMessage:
			g.MovementDirection = directionUp
		case rightMessage:
			g.MovementDirection = directionRight
		case downMessage:
			g.MovementDirection = directionDown
		case leftMessage:
			g.MovementDirection = directionLeft
		}
	}

	// if ebiten.IsKeyPressed(ebiten.KeyW) {
	// 	g.MovementDirection = directionUp
	// } else if ebiten.IsKeyPressed(ebiten.KeyD) {
	// 	g.MovementDirection = directionRight
	// } else if ebiten.IsKeyPressed(ebiten.KeyS) {
	// 	g.MovementDirection = directionDown
	// } else if ebiten.IsKeyPressed(ebiten.KeyA) {
	// 	g.MovementDirection = directionLeft
	// }

	if moveCounter > moveDelay {
		g.PreviousSnakeHead = g.SnakeHead
		switch g.MovementDirection {
		case directionUp:
			g.SnakeHead -= columnSize
		case directionRight:
			g.SnakeHead += 1
		case directionDown:
			g.SnakeHead += columnSize
		case directionLeft:
			g.SnakeHead -= 1
		}
		moveCounter = 0

		if g.SnakeHead == g.FoodPosition {
			g.SnakeBody = append(g.SnakeBody, g.PreviousSnakeHead)
			g.FoodPosition = -1
			g.Score += 1
		} else {
			g.SnakeBody = append(g.SnakeBody, g.PreviousSnakeHead)
			g.SnakeBody = g.SnakeBody[1:]
		}

		column, row := indexToColumnRow(g.SnakeHead)
		if slices.Contains(g.SnakeBody, g.SnakeHead) || column == 0 || column == columnSize-1 || row == 0 || row == rowSize-1 {
			g.IsGameOver = true
		}
	} else {
		moveCounter += deltaTime
	}

	if g.FoodPosition == -1 {
		for {
			randNumber := rand.Intn(sizeOfGameMap)
			column, row := indexToColumnRow(randNumber)
			if column > 2 && column < columnSize-3 && row > 2 && row < rowSize-3 && randNumber != g.SnakeHead && !slices.Contains(g.SnakeBody, randNumber) {
				g.FoodPosition = randNumber
				break
			}
		}
	}

	for index, _ := range g.GameMap {
		column, row := indexToColumnRow(index)
		if column == 0 || column == columnSize-1 || row == 0 || row == rowSize-1 {
			g.GameMap[index] = borderCell
		} else if index == g.SnakeHead {
			g.GameMap[index] = snakeHeadCell
		} else if index == g.FoodPosition {
			g.GameMap[index] = foodCell
		} else if slices.Contains(g.SnakeBody, index) {
			g.GameMap[index] = snakeBodyCell
		} else {
			g.GameMap[index] = emptyCell
		}
	}

	lastUpdateTime = currentTime

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	for index, value := range g.GameMap {
		column, row := indexToColumnRow(index)
		switch value {
		case borderCell:
			helperDrawRect(screen, row, column, borderColor)
		case snakeHeadCell:
			helperDrawRect(screen, row, column, snakeHeadColor)
		case foodCell:
			helperDrawRect(screen, row, column, foodColor)
		case snakeBodyCell:
			helperDrawRect(screen, row, column, snakeBodyColor)
		}
	}

	if g.IsGameOver {
		gameOverString := fmt.Sprintf("You lost the game.\nYour score is %v.\nRestarting in %v seconds.", g.Score, resetDelay/1000)
		text.Draw(screen, gameOverString, defaultFont, screenWidth/5, screenHeight/3, color.RGBA{255, 255, 255, 1})
	}

	text.Draw(screen, fmt.Sprintf("%v", g.Score), defaultFont, screenWidth-100, 70, color.RGBA{255, 255, 255, 1})
}

func (g *Game) Layout(width int, height int) (int, int) {
	return width, height
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.ChannelID != discordChannelID {
		return
	}

	// log.Printf("[%s] %s: %s\n", m.ChannelID, m.Author.Username, m.Content)

	messagesArray = append(messagesArray, m.Content)
}

func main() {
	finder := sysfont.NewFinder(nil)
	pathToFont := finder.Match("Arial").Filename

	ttfData, err := os.ReadFile(pathToFont)
	if err != nil {
		log.Fatal(err)
	}

	ttf, err := opentype.Parse(ttfData)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	const fontSize = 60
	defaultFont, err = opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = godotenv.Load()
	if err != nil {
		log.Fatalln(err.Error())
	}

	token := os.Getenv(botTokenKey)
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln(err.Error())
	}

	session.AddHandler(onMessageCreate)

	err = session.Open()
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer session.Close()

	log.Println("Bot is running")

	game := &Game{}
	game.Reset()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Snake Game")

	err = ebiten.RunGame(game)

	if err != nil {
		log.Fatalln(err.Error())
	}

	log.Println("Exiting")
}
