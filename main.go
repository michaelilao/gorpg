package main

import (
	"fmt"
	"runtime"

	"github.com/michaelilao/gorpg/game"
	"github.com/michaelilao/gorpg/ui"
)

const numWindows = 1

func main() {
	game := game.NewGame(numWindows)

	for i := 0; i < numWindows; i++ {
		go func(i int) {
			runtime.LockOSThread()
			ui := ui.NewUI(game.InputChan, game.LevelChans[i])
			ui.Run()
		}(i)
	}
	game.Run()
	fmt.Println("Done")
}
