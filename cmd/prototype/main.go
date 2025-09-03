package main

import (
	"github.com/joyfuldevs/project-lumos/cmd/prototype/app"
	_ "github.com/joyfuldevs/project-lumos/cmd/prototype/app/chat"
	_ "github.com/joyfuldevs/project-lumos/cmd/prototype/app/embedding"
	_ "github.com/joyfuldevs/project-lumos/cmd/prototype/app/hybridsearch"
	_ "github.com/joyfuldevs/project-lumos/cmd/prototype/app/index"
	_ "github.com/joyfuldevs/project-lumos/cmd/prototype/app/insert"
	_ "github.com/joyfuldevs/project-lumos/cmd/prototype/app/search"
)

func main() {
	app.Execute()
}
