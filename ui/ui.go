package ui

import (
	"bufio"
	"image/png"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/michaelilao/gorpg/game"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
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
	textureIndex      map[rune][]sdl.Rect
	prevKeyBoardState []uint8
	keyboardState     []uint8
	centerX           int
	centerY           int
	r                 *rand.Rand
	levelChan         chan *game.Level
	inputChan         chan *game.Input
	fontSmall         *ttf.Font
	fontMedium        *ttf.Font
	fontLarge         *ttf.Font
	eventBackground   *sdl.Texture
	str2TexSm         map[string]*sdl.Texture
	str2TexMd         map[string]*sdl.Texture
	str2TexLg         map[string]*sdl.Texture
}

func (ui *ui) loadTextureIndex() {
	ui.textureIndex = make(map[rune][]sdl.Rect)
	infile, err := os.Open("ui/assets/atlas-index.txt")
	checkError(err)
	scanner := bufio.NewScanner(infile)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		tileRune := rune(line[0])
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

	err = ttf.Init()
	checkError(err)
}

func NewUI(inputChan chan *game.Input, levelChan chan *game.Level) *ui {
	ui := &ui{}
	ui.inputChan = inputChan
	ui.str2TexSm = make(map[string]*sdl.Texture)
	ui.str2TexMd = make(map[string]*sdl.Texture)
	ui.str2TexLg = make(map[string]*sdl.Texture)

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

	ui.fontSmall, err = ttf.OpenFont("ui/assets/font.ttf", int(float64(ui.winWidth)*0.015))
	checkError(err)

	ui.fontMedium, err = ttf.OpenFont("ui/assets/font.ttf", 32)
	checkError(err)

	ui.fontLarge, err = ttf.OpenFont("ui/assets/font.ttf", 64)
	checkError(err)

	ui.eventBackground = ui.getSinglePixelTex(sdl.Color{0, 0, 0, 128})
	ui.eventBackground.SetBlendMode(sdl.BLENDMODE_BLEND)
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
			if tile.Rune != game.Blank {
				srcRects := ui.textureIndex[tile.Rune]
				srcRect := srcRects[ui.r.Intn(len(srcRects))]
				if tile.Visible || tile.Seen {
					destRect := sdl.Rect{int32(x*32) + offSetX, int32(y*32) + offSetY, 32, 32}
					pos := game.Pos{x, y}
					if level.Debug[pos] {
						ui.textureAtlas.SetColorMod(128, 0, 0)
					} else if tile.Seen && !tile.Visible {
						ui.textureAtlas.SetColorMod(128, 128, 128)
					} else {
						ui.textureAtlas.SetColorMod(255, 255, 255)
					}
					ui.renderer.Copy(ui.textureAtlas, &srcRect, &destRect)

					if tile.OverlayRune != game.Blank {
						srcRect := ui.textureIndex[tile.OverlayRune][0]
						ui.renderer.Copy(ui.textureAtlas, &srcRect, &destRect)
					}
				}
			}
		}
	}
	//21,59
	ui.textureAtlas.SetColorMod(255, 255, 255)
	for pos, monster := range level.Monsters {
		if level.Map[pos.Y][pos.X].Visible {
			monsterSrcRect := ui.textureIndex[(monster.Rune)][0]
			ui.renderer.Copy(ui.textureAtlas, &monsterSrcRect, &sdl.Rect{int32(pos.X*32) + offSetX, int32(pos.Y*32) + offSetY, 32, 32})
		}
	}
	playerSrcRect := ui.textureIndex['@'][0]
	ui.renderer.Copy(ui.textureAtlas, &playerSrcRect, &sdl.Rect{int32(level.Player.X*32) + offSetX, int32(level.Player.Y*32) + offSetY, 32, 32})

	textStart := int(float64(ui.winHeight) * .68)
	textWidth := int(float64(ui.winWidth) * .25)
	ui.renderer.Copy(ui.eventBackground, nil, &sdl.Rect{0, int32(textStart), int32(textWidth), int32(ui.winHeight - textStart)})

	i := level.EventPos
	count := 0

	_, fontSizeY, _ := ui.fontSmall.SizeUTF8("A")
	for {
		event := level.Events[i]
		if event != "" {
			tex := ui.stringToTexture(event, sdl.Color{255, 0, 0, 0}, FontSmall)
			_, _, w, h, err := tex.Query()
			checkError(err)
			ui.renderer.Copy(tex, nil, &sdl.Rect{5, int32(count*fontSizeY) + int32(textStart), w, h})
		}
		i = (i + 1) % (len(level.Events))
		count++
		if i == level.EventPos {
			break
		}
	}
	ui.renderer.Present()
	ui.renderer.Clear()

}

type FontSize int

const (
	FontSmall FontSize = iota
	FontMedium
	FontLarge
)

func (ui *ui) stringToTexture(s string, color sdl.Color, size FontSize) *sdl.Texture {
	var font *ttf.Font
	switch size {
	case FontSmall:
		font = ui.fontSmall
		tex, exists := ui.str2TexSm[s]
		if exists {
			return tex
		}
	case FontMedium:
		font = ui.fontMedium
		tex, exists := ui.str2TexMd[s]
		if exists {
			return tex
		}
	case FontLarge:
		font = ui.fontLarge
		tex, exists := ui.str2TexLg[s]
		if exists {
			return tex
		}
	}

	fontSurface, err := font.RenderUTF8Blended(s, color)
	checkError(err)

	tex, err := ui.renderer.CreateTextureFromSurface(fontSurface)
	checkError(err)
	switch size {
	case FontSmall:
		ui.str2TexSm[s] = tex
	case FontMedium:
		ui.str2TexMd[s] = tex
	case FontLarge:
		ui.str2TexLg[s] = tex
	}

	return tex
}
func (ui *ui) keyDownOnce(key uint8) bool {
	return ui.keyboardState[key] == 1 && ui.prevKeyBoardState[key] == 0
}

func (ui *ui) keyPressed(key uint8) bool {
	return ui.keyboardState[key] == 0 && ui.prevKeyBoardState[key] == 1
}

func (ui *ui) getSinglePixelTex(color sdl.Color) *sdl.Texture {
	tex, err := ui.renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, 1, 1)
	checkError(err)

	pixels := make([]byte, 4)
	pixels[0] = color.R
	pixels[1] = color.G
	pixels[2] = color.B
	pixels[3] = color.A
	tex.Update(nil, pixels, 4)
	return tex
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
			if ui.keyDownOnce(sdl.SCANCODE_UP) {
				input.Typ = game.Up
			}
			if ui.keyDownOnce(sdl.SCANCODE_DOWN) {
				input.Typ = game.Down
			}
			if ui.keyDownOnce(sdl.SCANCODE_LEFT) {
				input.Typ = game.Left
			}
			if ui.keyDownOnce(sdl.SCANCODE_RIGHT) {
				input.Typ = game.Right
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
