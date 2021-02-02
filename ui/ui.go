package ui

import (
	"bufio"
	"image/png"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/michaelilao/rpg/game"
	"github.com/veandco/go-sdl2/sdl"
)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

const winWidth, winHeight = 1280, 720

type ui struct {
	winWidth          int
	winHeight         int
	renderer          *sdl.Renderer
	window            *sdl.Window
	textureAtlas      *sdl.Texture
	textureIndex      map[game.Tile][]sdl.Rect
	prevKeyBoardState []uint8
	keyboardState     []uint8
	centerX           int
	centerY           int
	r                 *rand.Rand
	levelChan         chan *game.Level
	inputChan         chan *game.Input
}

func (ui *ui) loadTextureIndex() {
	ui.textureIndex = make(map[game.Tile][]sdl.Rect)
	infile, err := os.Open("ui/assets/atlas-index.txt")
	checkError(err)
	scanner := bufio.NewScanner(infile)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		tileRune := game.Tile(line[0])
		xy := line[1:]
		splitXYC := strings.Split(xy, ",")
		x, err := strconv.ParseInt(strings.TrimSpace(splitXYC[0]), 10, 64)
		checkError(err)
		y, err := strconv.ParseInt(strings.TrimSpace(splitXYC[1]), 10, 64)
		checkError(err)
		variationCount, err := strconv.ParseInt(strings.TrimSpace(splitXYC[2]), 10, 64)
		checkError(err)

		var rects []sdl.Rect
		for i := int64(0); i < variationCount; i++ {
			rects = append(rects, sdl.Rect{int32(x * 32), int32(y * 32), 32, 32})
			x++
			if x > 62 {
				x = 0
				y++
			}
		}
		ui.textureIndex[tileRune] = rects
	}

}
func (ui *ui) imgFileToTexture(filename string) *sdl.Texture {
	infile, err := os.Open(filename)
	checkError(err)

	defer infile.Close()
	img, err := png.Decode(infile)
	checkError(err)

	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y
	pixels := make([]byte, w*h*4)
	index := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixels[index] = byte(r / 256)
			index++
			pixels[index] = byte(g / 256)
			index++
			pixels[index] = byte(b / 256)
			index++
			pixels[index] = byte(a / 256)
			index++
		}
	}
	tex, err := ui.renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, int32(w), int32(h))
	checkError(err)

	tex.Update(nil, pixels, w*4)
	err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	checkError(err)

	return tex
}

func init() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	checkError(err)
}

func NewUI(inputChan chan *game.Input, levelChan chan *game.Level) *ui {
	ui := &ui{}
	ui.inputChan = inputChan
	ui.levelChan = levelChan
	ui.r = rand.New(rand.NewSource(1))
	ui.winWidth = winWidth
	ui.winHeight = winHeight

	var err error
	ui.window, err = sdl.CreateWindow("RPG", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(ui.winWidth), int32(ui.winHeight), sdl.WINDOW_SHOWN)
	checkError(err)

	ui.renderer, err = sdl.CreateRenderer(ui.window, -1, sdl.RENDERER_ACCELERATED)
	checkError(err)

	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	ui.textureAtlas = ui.imgFileToTexture("ui/assets/tiles.png")
	ui.loadTextureIndex()

	ui.keyboardState = sdl.GetKeyboardState()
	ui.prevKeyBoardState = make([]uint8, len(ui.keyboardState))
	for i, v := range ui.keyboardState {
		ui.prevKeyBoardState[i] = v
	}
	ui.centerX = -1
	ui.centerY = -1
	checkError(err)

	return ui
}
func (ui *ui) Draw(level *game.Level) {
	if ui.centerX == -1 && ui.centerY == -1 {
		ui.centerX = level.Player.X
		ui.centerY = level.Player.Y
	}
	limit := 7
	if level.Player.X > ui.centerX+limit {
		ui.centerX++
	} else if level.Player.X < ui.centerX-limit {
		ui.centerX--
	} else if level.Player.Y > ui.centerY+limit {
		ui.centerY++
	} else if level.Player.Y < ui.centerY-limit {
		ui.centerY--
	}
	offSetX := int32((ui.winWidth / 2) - ui.centerX*32)
	offSetY := int32((ui.winHeight / 2) - ui.centerY*32)

	ui.r.Seed(1)
	for y, row := range level.Map {
		for x, tile := range row {
			if tile != game.Blank {
				srcRects := ui.textureIndex[tile]
				srcRect := srcRects[ui.r.Intn(len(srcRects))]
				destRect := sdl.Rect{int32(x*32) + offSetX, int32(y*32) + offSetY, 32, 32}

				pos := game.Pos{x, y}
				if level.Debug[pos] {
					ui.textureAtlas.SetColorMod(128, 0, 0)
				} else {
					ui.textureAtlas.SetColorMod(255, 255, 255)
				}
				ui.renderer.Copy(ui.textureAtlas, &srcRect, &destRect)
			}
		}
	}
	//21,59

	for pos, monster := range level.Monsters {
		monsterSrcRect := ui.textureIndex[game.Tile(monster.Rune)][0]
		ui.renderer.Copy(ui.textureAtlas, &monsterSrcRect, &sdl.Rect{int32(pos.X*32) + offSetX, int32(pos.Y*32) + offSetY, 32, 32})
	}
	playerSrcRect := ui.textureIndex['@'][0]
	ui.renderer.Copy(ui.textureAtlas, &playerSrcRect, &sdl.Rect{int32(level.Player.X*32) + offSetX, int32(level.Player.Y*32) + offSetY, 32, 32})
	ui.renderer.Present()
	ui.renderer.Clear()

}

func (ui *ui) Run() {
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				ui.inputChan <- &game.Input{Typ: game.QuitGame}
			case *sdl.WindowEvent:
				if e.Event == sdl.WINDOWEVENT_CLOSE {
					ui.inputChan <- &game.Input{Typ: game.CloseWindow, LevelChannel: ui.levelChan}
				}
			}

		}

		select {
		case newLevel, ok := <-ui.levelChan:
			if ok {
				ui.Draw(newLevel)
			}
		default:
		}

		if sdl.GetKeyboardFocus() == ui.window && sdl.GetMouseFocus() == ui.window {

			var input game.Input
			if ui.keyboardState[sdl.SCANCODE_UP] == 1 && ui.prevKeyBoardState[sdl.SCANCODE_UP] == 0 {
				input.Typ = game.Up
			}
			if ui.keyboardState[sdl.SCANCODE_DOWN] == 1 && ui.prevKeyBoardState[sdl.SCANCODE_DOWN] == 0 {
				input.Typ = game.Down
			}
			if ui.keyboardState[sdl.SCANCODE_LEFT] == 1 && ui.prevKeyBoardState[sdl.SCANCODE_LEFT] == 0 {
				input.Typ = game.Left
			}
			if ui.keyboardState[sdl.SCANCODE_RIGHT] == 1 && ui.prevKeyBoardState[sdl.SCANCODE_RIGHT] == 0 {
				input.Typ = game.Right
			}
			if ui.keyboardState[sdl.SCANCODE_S] == 1 && ui.prevKeyBoardState[sdl.SCANCODE_S] == 0 {
				input.Typ = game.Search
			}
			for i, v := range ui.keyboardState {
				ui.prevKeyBoardState[i] = v
			}
			if input.Typ != game.None {
				ui.inputChan <- &input
			}
			sdl.Delay(10)
		}
	}
}
